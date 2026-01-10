package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"tmux-fsm/editor"
)

// ShadowEngine 核心执行引擎
// 负责处理 Intent，生成并应用 Transaction，维护 History
type ShadowEngine struct {
	planner      Planner
	history      History
	resolver     AnchorResolver
	projection   Projection
	reality      RealityReader
	proofBuilder *ProofBuilder
	dag          *editor.OperationDAG
}

func NewShadowEngine(planner Planner, resolver AnchorResolver, projection Projection, reality RealityReader) *ShadowEngine {
	return &ShadowEngine{
		planner:      planner,
		history:      NewInMemoryHistory(100),
		resolver:     resolver,
		projection:   projection,
		reality:      reality,
		proofBuilder: NewProofBuilder(),
		dag:          editor.NewOperationDAG(),
	}
}

func (e *ShadowEngine) ApplyIntent(hctx HandleContext, intent Intent, snapshot Snapshot) (*Verdict, error) {
	requestID := hctx.RequestID
	actorID := hctx.ActorID

	log.Printf("Applying intent: RequestID=%s, Kind=%d, PaneID=%s, SnapshotHash=%s",
		requestID, intent.GetKind(), intent.GetPaneID(), intent.GetSnapshotHash())

	// Initialize AuditRecord v2
	auditRecord := &AuditRecord{
		Version:      "v2",
		RequestID:    requestID,
		ActorID:      actorID,
		TimestampUTC: time.Now().Unix(),
		IntentKind:   fmt.Sprintf("%d", intent.GetKind()),
		DecisionPath: "Intent",
		Entries:      []AuditEntryV2{},
		Result:       AuditResult{Status: "Pending", WorldDrift: false},
	}

	// Phase 6.3: Temporal Adjudication (World Drift Check)
	// Engine owns the authority to reject execution if current reality != intent's expectation.
	if intent.GetSnapshotHash() != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			if string(current.Hash) != intent.GetSnapshotHash() {
				log.Printf("World drift detected: expected %s, got %s. Proceeding anyway (Optimistic).", intent.GetSnapshotHash(), string(current.Hash))

				// Add audit entry as warning
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Adjudicate",
					Action:  "Warning",
					Outcome: "Proceed",
					Detail:  "World drift detected but ignored (Optimistic Execution)",
					Meta:    map[string]string{"expected": intent.GetSnapshotHash(), "actual": string(current.Hash)},
					At:      time.Now().Unix(),
				})
			} else {
				log.Printf("Time consistency verified for intent in pane %s", intent.GetPaneID())

				// Add audit entry
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Adjudicate",
					Action:  "Verify",
					Outcome: "Success",
					Detail:  "Time consistency verified",
					Meta:    map[string]string{"pane": intent.GetPaneID()},
					At:      time.Now().Unix(),
				})
			}
		} else {
			log.Printf("Could not read current reality for pane %s: %v", intent.GetPaneID(), err)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Adjudicate",
				Action:  "Verify",
				Outcome: "Warning",
				Detail:  fmt.Sprintf("Could not read current reality: %v", err),
				Meta:    map[string]string{"pane": intent.GetPaneID()},
				At:      time.Now().Unix(),
			})
		}
		// If Reality check fails (IO error), we might proceed with warning or fail fast.
		// For now, assume if we can't read reality, it's a structural error but not necessarily drift.
	}

	// 1. Handle Undo/Redo explicitly
	kind := intent.GetKind()
	if kind == IntentUndo {
		log.Printf("Processing undo intent for pane %s", intent.GetPaneID())
		return e.performUndoWithRequestID(requestID, auditRecord)
	}
	if kind == IntentRedo {
		log.Printf("Processing redo intent for pane %s", intent.GetPaneID())
		return e.performRedoWithRequestID(requestID, auditRecord)
	}

	// 2. Plan: Generate Facts
	log.Printf("Planning facts for intent in pane %s", intent.GetPaneID())
	facts, inverseFacts, err := e.planner.Build(intent, snapshot)
	if err != nil {
		log.Printf("Failed to plan facts for intent in pane %s: %v", intent.GetPaneID(), err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Plan",
			Action:  "Build",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to plan facts: %v", err),
			Meta:    map[string]string{"pane": intent.GetPaneID()},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to plan facts: %v", err),
		}

		return &Verdict{Kind: VerdictBlocked, Audit: convertAuditRecordToLegacy(auditRecord)}, err
	}
	log.Printf("Successfully planned %d facts for intent in pane %s", len(facts), intent.GetPaneID())

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Plan",
		Action:  "Build",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully planned %d facts", len(facts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(facts)), "pane": intent.GetPaneID()},
		At:      time.Now().Unix(),
	})

	// [Phase 5.1] 4. Resolve: 定位权移交
	// [Phase 5.4] 包含 Reconciliation 检查
	// [Phase 6.3] 包含 World Drift 检查 (SnapshotHash)
	log.Printf("Resolving facts for intent in pane %s", intent.GetPaneID())
	// Contextual Logic: If intent doesn't specify an expected state (fresh intent),
	// we bind it to the snapshot we just took (Current Reality).
	// This ensures consistency between Planning (using snapshot) and Resolution.
	expectedHash := intent.GetSnapshotHash()
	if expectedHash == "" {
		expectedHash = string(snapshot.Hash)
	}
	resolvedFacts, err := e.resolver.ResolveFacts(facts, expectedHash)
	if err != nil {
		log.Printf("Failed to resolve facts for intent in pane %s: %v", intent.GetPaneID(), err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Resolve",
			Action:  "Resolve",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to resolve facts: %v", err),
			Meta:    map[string]string{"pane": intent.GetPaneID()},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to resolve facts: %v", err),
		}

		return &Verdict{Kind: VerdictBlocked, Audit: convertAuditRecordToLegacy(auditRecord)}, err
	}
	log.Printf("Successfully resolved %d facts for intent in pane %s", len(resolvedFacts), intent.GetPaneID())

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Resolve",
		Action:  "Resolve",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully resolved %d facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "pane": intent.GetPaneID()},
		At:      time.Now().Unix(),
	})

	// [Phase 7] Determine overall safety
	safety := SafetyExact
	for _, rf := range resolvedFacts {
		if rf.Safety > safety {
			safety = rf.Safety
		}
	}
	log.Printf("Determined safety level %d for intent in pane %s", safety, intent.GetPaneID())

	if safety == SafetyFuzzy && !intent.IsPartialAllowed() {
		log.Printf("Fuzzy resolution disallowed by policy for intent in pane %s", intent.GetPaneID())

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Policy",
			Action:  "Validate",
			Outcome: "Rejected",
			Detail:  "Fuzzy resolution disallowed by policy",
			Meta:    map[string]string{"safety": fmt.Sprintf("%d", safety), "partial_allowed": fmt.Sprintf("%t", intent.IsPartialAllowed())},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  "Fuzzy resolution disallowed by policy",
		}

		return &Verdict{
				Kind:    VerdictRejected,
				Safety:  SafetyUnsafe,
				Message: "Fuzzy resolution disallowed by policy",
				Audit:   convertAuditRecordToLegacy(auditRecord),
			}, &WorldDriftError{
				Reason:   DriftSnapshotMismatch,
				Expected: intent.GetSnapshotHash(),
				Actual:   intent.GetSnapshotHash(), // Not actually a snapshot mismatch, but using for policy violation
				Message:  "Fuzzy resolution disallowed by policy",
			}
	}

	// [Phase 7] Inverse Fact Enrichment:
	// If the planner couldn't generate inverse facts (common for semantic deletes),
	// we generate them now using the reality captured during resolution.
	if len(inverseFacts) == 0 && len(resolvedFacts) > 0 {
		log.Printf("Generating inverse facts for intent in pane %s", intent.GetPaneID())
		for _, rf := range resolvedFacts {
			if rf.Kind == FactDelete && rf.Payload.OldText != "" {
				// [Phase 7] Axiom 7.6: Paradox Resolved
				// Undo is return-to-origin, not a new fork.
				// Line-level semantic fingerprints are ignored because global post-hash already secured the timeline.
				invAnchor := Anchor{
					PaneID: rf.Anchor.PaneID,
					Kind:   AnchorAbsolute,
					Ref:    []int{rf.Anchor.Line, rf.Anchor.Start},
				}

				invMeta := make(map[string]interface{})
				for k, v := range rf.Meta {
					invMeta[k] = v
				}
				invMeta["operation"] = "undo_restore"

				inverseFacts = append(inverseFacts, Fact{
					Kind:   FactInsert,
					Anchor: invAnchor,
					Payload: FactPayload{
						Text: rf.Payload.OldText,
					},
					Meta: invMeta,
				})
			}
		}
		log.Printf("Generated %d inverse facts for intent in pane %s", len(inverseFacts), intent.GetPaneID())

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Prepare",
			Action:  "Generate",
			Outcome: "Success",
			Detail:  fmt.Sprintf("Generated %d inverse facts", len(inverseFacts)),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(inverseFacts)), "pane": intent.GetPaneID()},
			At:      time.Now().Unix(),
		})
	}

	// 3. Create Transaction
	txID := TransactionID(fmt.Sprintf("tx-%d", time.Now().UnixNano()))
	log.Printf("Creating transaction %s for intent in pane %s", txID, intent.GetPaneID())
	tx := &Transaction{
		ID:           txID,
		Intent:       intent,
		Facts:        facts,
		InverseFacts: inverseFacts,
		Safety:       safety,
		Timestamp:    time.Now().Unix(),
		AllowPartial: intent.IsPartialAllowed(),
	}

	// Update audit record with transaction ID
	auditRecord.TransactionID = string(txID)

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot := snapshot

	// 5. Project: Execute
	log.Printf("Projecting %d resolved facts for intent in pane %s", len(resolvedFacts), intent.GetPaneID())
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to project facts for intent in pane %s: %v", intent.GetPaneID(), err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Project",
			Action:  "Apply",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to project facts: %v", err),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "pane": intent.GetPaneID()},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to project facts: %v", err),
		}

		return &Verdict{Kind: VerdictBlocked, Audit: convertAuditRecordToLegacy(auditRecord)}, err
	}
	log.Printf("Successfully projected facts for intent in pane %s", intent.GetPaneID())

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Project",
		Action:  "Apply",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully projected %d facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "pane": intent.GetPaneID()},
		At:      time.Now().Unix(),
	})
	tx.Applied = true

	// [Phase 7] Capture PostSnapshotHash for Undo verification
	var postSnap Snapshot
	if e.reality != nil {
		var err error
		postSnap, err = e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			tx.PostSnapshotHash = string(postSnap.Hash)
			log.Printf("Captured post-snapshot hash %s for transaction %s", tx.PostSnapshotHash, txID)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Record",
				Action:  "Capture",
				Outcome: "Success",
				Detail:  fmt.Sprintf("Captured post-snapshot hash: %s", tx.PostSnapshotHash),
				Meta:    map[string]string{"hash": tx.PostSnapshotHash, "tx": string(txID)},
				At:      time.Now().Unix(),
			})
		} else {
			log.Printf("Failed to capture post-snapshot for transaction %s: %v", txID, err)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Record",
				Action:  "Capture",
				Outcome: "Failure",
				Detail:  fmt.Sprintf("Failed to capture post-snapshot: %v", err),
				Meta:    map[string]string{"tx": string(txID)},
				At:      time.Now().Unix(),
			})
		}
	}

	// [Phase 9] Verify that the projection achieved the expected result
	if e.projection != nil && e.reality != nil {
		verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
		if !verification.OK {
			log.Printf("Projection verification failed for transaction %s: %s", txID, verification.Message)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Verify",
				Action:  "Validate",
				Outcome: "Failure",
				Detail:  fmt.Sprintf("Verification failed: %s", verification.Message),
				Meta:    map[string]string{"tx": string(txID), "message": verification.Message},
				At:      time.Now().Unix(),
			})

			// For now, we still consider this applied but log the verification issue
			log.Printf("[WEAVER] Projection verification failed: %s", verification.Message)
		} else {
			log.Printf("Projection verification succeeded for transaction %s", txID)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Verify",
				Action:  "Validate",
				Outcome: "Success",
				Detail:  "Projection matched expectations",
				Meta:    map[string]string{"tx": string(txID)},
				At:      time.Now().Unix(),
			})
		}
	}

	// 6. Update History
	if len(facts) > 0 {
		log.Printf("Pushing transaction %s to history for pane %s", txID, intent.GetPaneID())
		e.history.Push(tx)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "History",
			Action:  "Push",
			Outcome: "Success",
			Detail:  fmt.Sprintf("Transaction %s pushed to history", txID),
			Meta:    map[string]string{"tx": string(txID), "pane": intent.GetPaneID()},
			At:      time.Now().Unix(),
		})
	}

	// Update final result
	auditRecord.Result = AuditResult{
		Status:     "Committed",
		WorldDrift: false,
	}

	// Generate proof for this transaction
	if e.proofBuilder != nil {
		proof := e.proofBuilder.BuildProof(tx, auditRecord)
		log.Printf("Generated proof for transaction %s: PreState=%s, PostState=%s, Facts=%s, Audit=%s",
			txID, proof.PreStateHash, proof.PostStateHash, proof.FactsHash, proof.AuditHash)

		// ✅ Bind ProofHash to Transaction (Authority anchoring)
		proofHash := HashProof(proof)
		tx.ProofHash = proofHash

		log.Printf("Bound ProofHash to transaction %s: %s", txID, tx.ProofHash)
	}

	// Phase 6.0: Populate DAG
	if e.dag != nil && len(resolvedFacts) > 0 {
		// Use the first fact as the primary operation? or Create a node for each?
		// Usually atomic intent -> atomic DAG node.
		// If multiple facts (e.g. multiple cursors), we might need composite node or multiple nodes.
		// For now, let's assume 1:1 or 1:N mapping where intent is the grouper.
		// But DAGNode stores 'ResolvedOperation'.
		// If we store the *Intent* as the semantic parent, we might want one Node per Intent.
		// However, editor.ResolvedOperation is fine-grained.

		parentIDs := e.dag.Tips // Use current tips as parents

		for _, rf := range resolvedFacts {
			op := convertFactToOp(rf)
			_, err := e.dag.AddNode(op, parentIDs)
			if err != nil {
				log.Printf("Failed to add node to DAG: %v", err)
			}
			// Sequence them? If we add all with same parents, they are concurrent.
			// Facts in a transaction are atomic/simultaneous.
			// So using same 'parentIDs' (previous tips) is correct for "parallel" application on state?
			// Or should they be sequenced?
			// If facts are ordered (e.g. sequential edits), we should chain them.
			// Current Planner usually produces independent facts or sequenced?
			// Assumption: Sequenced.
			// Let's update parentIDs for next fact to chain them.
			// But Transaction is Atomic.
			// Let's chain them for safety.
			// Actually, reusing same parents means they are parallel forks.
			// Ideally, we want a single DAG Node representing the Transaction?
			// But DAGNode holds ResolvedOperation (singular).
			// Let's chain them.
			// Note: We need to retrieve the new node's ID to use as parent for next.
			// But AddNode returns *DAGNode.
			// Since we just added it, it becomes a Tip.
			// So for the next iteration, we should use the *new* tips?
			// e.dag.Tips will be updated by AddNode.
			// So if we just pass e.dag.Tips, are we implicitly chaining?
			// e.dag.Tips will contain the *newly added node*.
			// So yes, chaining happens naturally if we use e.dag.Tips.
			// But for the *first* fact, we use pre-tx tips.
			// For *subsequent* facts in same tx, we use the tip created by previous fact.
			parentIDs = e.dag.Tips
		}
	}

	log.Printf("Successfully applied intent for pane %s, transaction %s", intent.GetPaneID(), intent.GetPaneID())
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: tx,
		Safety:      safety,
		Audit:       convertAuditRecordToLegacy(auditRecord),
	}, nil
}

// Helper function to convert AuditRecord to legacy AuditEntry format
func convertAuditRecordToLegacy(record *AuditRecord) []AuditEntry {
	var legacy []AuditEntry

	for _, entry := range record.Entries {
		legacy = append(legacy, AuditEntry{
			Step:   fmt.Sprintf("[%s] %s", entry.Phase, entry.Action),
			Result: fmt.Sprintf("%s: %s", entry.Outcome, entry.Detail),
		})
	}

	// Add a summary entry for the result
	legacy = append(legacy, AuditEntry{
		Step:   "FinalResult",
		Result: fmt.Sprintf("%s (Drift: %t)", record.Result.Status, record.Result.WorldDrift),
	})

	return legacy
}

func (e *ShadowEngine) performUndo() (*Verdict, error) {
	// Generate a RequestID for this undo operation - this should be derived from parent context
	// For now, using a default since we don't have the parent context here
	// In a proper implementation, undo should be called with the parent request context
	parentRequestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Create a minimal audit record for this operation
	auditRecord := &AuditRecord{
		Version:      "v2",
		RequestID:    parentRequestID + ":undo", // Derived from parent
		ActorID:      "system",                  // Undo is system-triggered
		TimestampUTC: time.Now().Unix(),
		IntentKind:   "Undo",
		DecisionPath: "System",
		Entries:      []AuditEntryV2{},
		Result:       AuditResult{Status: "Pending", WorldDrift: false},
	}

	return e.performUndoWithRequestID(parentRequestID, auditRecord)
}

// performUndoWithRequestID performs undo with a specific RequestID and audit record
func (e *ShadowEngine) performUndoWithRequestID(parentRequestID string, auditRecord *AuditRecord) (*Verdict, error) {
	// ✅ Undo RequestID derivation (not new generation)
	requestID := parentRequestID + ":undo"
	log.Printf("Starting undo operation: RequestID=%s", requestID)
	tx := e.history.PopUndo()
	if tx == nil {
		log.Printf("No transaction to undo")

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Undo",
			Action:  "Pop",
			Outcome: "NoOp",
			Detail:  "Nothing to undo",
			Meta:    map[string]string{"request_id": requestID},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Skipped",
		}

		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to undo", Audit: convertAuditRecordToLegacy(auditRecord)}, nil
	}

	log.Printf("Attempting to undo transaction %s for pane %s", tx.ID, tx.Intent.GetPaneID())

	// [Phase 7] Axiom 7.5: Undo Is Verified Replay
	if tx.PostSnapshotHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != tx.PostSnapshotHash {
			log.Printf("World drift detected during undo: expected %s, got %s", tx.PostSnapshotHash, string(current.Hash))

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Adjudicate",
				Action:  "Verify",
				Outcome: "Rejected",
				Detail:  "World drift detected during undo",
				Meta:    map[string]string{"expected": tx.PostSnapshotHash, "actual": string(current.Hash), "tx": string(tx.ID)},
				At:      time.Now().Unix(),
			})

			// Update result
			auditRecord.Result = AuditResult{
				Status:      "Rejected",
				WorldDrift:  true,
				DriftReason: string(DriftUndoMismatch),
				Error:       "World drift: cannot undo safely",
			}

			// Put it back to undo stack since we didn't apply it
			e.history.PushBack(tx)
			return &Verdict{
					Kind:    VerdictRejected,
					Message: "World drift: cannot undo safely",
					Safety:  SafetyUnsafe,
					Audit:   convertAuditRecordToLegacy(auditRecord),
				}, &WorldDriftError{
					Reason:   DriftUndoMismatch,
					Expected: tx.PostSnapshotHash,
					Actual:   string(current.Hash),
					Message:  "World drift: cannot undo safely",
				}
		}
		log.Printf("Undo context verified for transaction %s", tx.ID)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Adjudicate",
			Action:  "Verify",
			Outcome: "Success",
			Detail:  "Undo context verified",
			Meta:    map[string]string{"tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})
	}

	// [Phase 5.1] Resolve InverseFacts
	// [Phase 6.3] Use recorded PostHash if available (passed as expectedHash)
	log.Printf("Resolving %d inverse facts for undo of transaction %s", len(tx.InverseFacts), tx.ID)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.InverseFacts, tx.PostSnapshotHash)
	if err != nil {
		log.Printf("Failed to resolve inverse facts for undo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Resolve",
			Action:  "Resolve",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to resolve inverse facts: %v", err),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(tx.InverseFacts)), "tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		e.history.PushBack(tx)

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to resolve inverse facts: %v", err),
		}

		return nil, err
	}
	log.Printf("Successfully resolved %d inverse facts for undo of transaction %s", len(resolvedFacts), tx.ID)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Resolve",
		Action:  "Resolve",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully resolved %d inverse facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		log.Printf("Failed to capture pre-snapshot for undo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Verify",
			Action:  "Capture",
			Outcome: "Warning",
			Detail:  fmt.Sprintf("Failed to capture pre-snapshot: %v", err),
			Meta:    map[string]string{"tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if len(resolvedFacts) > 0 {
		log.Printf("[WEAVER] Undo: Applying %d inverse facts for transaction %s. Text length: %d chars.",
			len(resolvedFacts), tx.ID, len(resolvedFacts[0].Payload.Text))
	}
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to apply inverse facts for undo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Project",
			Action:  "Apply",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to apply inverse facts: %v", err),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		e.history.PushBack(tx)

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to apply inverse facts: %v", err),
		}

		return nil, err
	}
	log.Printf("Successfully applied inverse facts for undo of transaction %s", tx.ID)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Project",
		Action:  "Apply",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully applied %d inverse facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// [Phase 9] Verify undo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("Undo verification failed for transaction %s: %s", tx.ID, verification.Message)

				// Add audit entry
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Verify",
					Action:  "Validate",
					Outcome: "Failure",
					Detail:  fmt.Sprintf("Undo verification failed: %s", verification.Message),
					Meta:    map[string]string{"tx": string(tx.ID), "message": verification.Message},
					At:      time.Now().Unix(),
				})

				log.Printf("[WEAVER] Undo projection verification failed: %s", verification.Message)
			} else {
				log.Printf("Undo verification succeeded for transaction %s", tx.ID)

				// Add audit entry
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Verify",
					Action:  "Validate",
					Outcome: "Success",
					Detail:  "Undo projection matched expectations",
					Meta:    map[string]string{"tx": string(tx.ID)},
					At:      time.Now().Unix(),
				})
			}
		} else {
			log.Printf("Failed to read post-snapshot for undo verification of transaction %s: %v", tx.ID, err)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Verify",
				Action:  "Validate",
				Outcome: "Warning",
				Detail:  fmt.Sprintf("Failed to read post-snapshot: %v", err),
				Meta:    map[string]string{"tx": string(tx.ID)},
				At:      time.Now().Unix(),
			})
		}
	}

	// Move to Redo Stack
	log.Printf("Moving transaction %s from undo to redo stack", tx.ID)
	e.history.AddRedo(tx)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "History",
		Action:  "Move",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Transaction %s moved from undo to redo", tx.ID),
		Meta:    map[string]string{"tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// Update final result
	auditRecord.Result = AuditResult{
		Status: "Committed",
	}

	// Update audit record with transaction ID
	auditRecord.TransactionID = string(tx.ID)

	// Generate proof for this undo transaction
	if e.proofBuilder != nil {
		proof := e.proofBuilder.BuildProof(tx, auditRecord)
		log.Printf("Generated proof for undo transaction %s: PreState=%s, PostState=%s, Facts=%s, Audit=%s",
			tx.ID, proof.PreStateHash, proof.PostStateHash, proof.FactsHash, proof.AuditHash)
	}

	log.Printf("Successfully undone transaction %s", tx.ID)
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Undone tx: %s", tx.ID),
		Transaction: tx,
		Audit:       convertAuditRecordToLegacy(auditRecord),
	}, nil
}

func (e *ShadowEngine) performRedo() (*Verdict, error) {
	// Generate a RequestID for this redo operation - this should be derived from parent context
	// For now, using a default since we don't have the parent context here
	// In a proper implementation, redo should be called with the parent request context
	parentRequestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Create a minimal audit record for this operation
	auditRecord := &AuditRecord{
		Version:      "v2",
		RequestID:    parentRequestID + ":redo", // Derived from parent
		ActorID:      "system",                  // Redo is system-triggered
		TimestampUTC: time.Now().Unix(),
		IntentKind:   "Redo",
		DecisionPath: "System",
		Entries:      []AuditEntryV2{},
		Result:       AuditResult{Status: "Pending", WorldDrift: false},
	}

	return e.performRedoWithRequestID(parentRequestID, auditRecord)
}

// performRedoWithRequestID performs redo with a specific RequestID and audit record
func (e *ShadowEngine) performRedoWithRequestID(parentRequestID string, auditRecord *AuditRecord) (*Verdict, error) {
	// ✅ Redo RequestID derivation (not new generation)
	requestID := parentRequestID + ":redo"
	log.Printf("Starting redo operation: RequestID=%s", requestID)
	tx := e.history.PopRedo()
	if tx == nil {
		log.Printf("No transaction to redo")

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Redo",
			Action:  "Pop",
			Outcome: "NoOp",
			Detail:  "Nothing to redo",
			Meta:    map[string]string{"request_id": requestID},
			At:      time.Now().Unix(),
		})

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Skipped",
		}

		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to redo", Audit: convertAuditRecordToLegacy(auditRecord)}, nil
	}

	log.Printf("Attempting to redo transaction %s for pane %s", tx.ID, tx.Intent.GetPaneID())

	// [Phase 7] Redo verification (must match Pre-state)
	preHash := tx.Intent.GetSnapshotHash()
	if preHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != preHash {
			log.Printf("World drift detected during redo: expected %s, got %s", preHash, string(current.Hash))

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Adjudicate",
				Action:  "Verify",
				Outcome: "Rejected",
				Detail:  "World drift detected during redo",
				Meta:    map[string]string{"expected": preHash, "actual": string(current.Hash), "tx": string(tx.ID)},
				At:      time.Now().Unix(),
			})

			// Update result
			auditRecord.Result = AuditResult{
				Status:      "Rejected",
				WorldDrift:  true,
				DriftReason: string(DriftRedoMismatch),
				Error:       "World drift: cannot redo safely",
			}

			e.history.AddRedo(tx)
			return &Verdict{
					Kind:    VerdictRejected,
					Message: "World drift: cannot redo safely",
					Safety:  SafetyUnsafe,
					Audit:   convertAuditRecordToLegacy(auditRecord),
				}, &WorldDriftError{
					Reason:   DriftRedoMismatch,
					Expected: preHash,
					Actual:   string(current.Hash),
					Message:  "World drift: cannot redo safely",
				}
		}
		log.Printf("Redo context verified for transaction %s", tx.ID)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Adjudicate",
			Action:  "Verify",
			Outcome: "Success",
			Detail:  "Redo context verified",
			Meta:    map[string]string{"tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})
	}

	// [Phase 5.1] Resolve Facts
	log.Printf("Resolving %d facts for redo of transaction %s", len(tx.Facts), tx.ID)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.Facts, preHash)
	if err != nil {
		log.Printf("Failed to resolve facts for redo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Resolve",
			Action:  "Resolve",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to resolve facts: %v", err),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(tx.Facts)), "tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		e.history.AddRedo(tx)

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to resolve facts: %v", err),
		}

		return nil, err
	}
	log.Printf("Successfully resolved %d facts for redo of transaction %s", len(resolvedFacts), tx.ID)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Resolve",
		Action:  "Resolve",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully resolved %d facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		log.Printf("Failed to capture pre-snapshot for redo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Verify",
			Action:  "Capture",
			Outcome: "Warning",
			Detail:  fmt.Sprintf("Failed to capture pre-snapshot: %v", err),
			Meta:    map[string]string{"tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	log.Printf("Projecting %d resolved facts for redo of transaction %s", len(resolvedFacts), tx.ID)
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to apply facts for redo of transaction %s: %v", tx.ID, err)

		// Add audit entry
		auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
			Phase:   "Project",
			Action:  "Apply",
			Outcome: "Failure",
			Detail:  fmt.Sprintf("Failed to apply facts: %v", err),
			Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
			At:      time.Now().Unix(),
		})

		e.history.AddRedo(tx)

		// Update result
		auditRecord.Result = AuditResult{
			Status: "Rejected",
			Error:  fmt.Sprintf("Failed to apply facts: %v", err),
		}

		return nil, err
	}
	log.Printf("Successfully applied facts for redo of transaction %s", tx.ID)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "Project",
		Action:  "Apply",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Successfully applied %d facts", len(resolvedFacts)),
		Meta:    map[string]string{"count": fmt.Sprintf("%d", len(resolvedFacts)), "tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// [Phase 9] Verify redo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("Redo verification failed for transaction %s: %s", tx.ID, verification.Message)

				// Add audit entry
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Verify",
					Action:  "Validate",
					Outcome: "Failure",
					Detail:  fmt.Sprintf("Redo verification failed: %s", verification.Message),
					Meta:    map[string]string{"tx": string(tx.ID), "message": verification.Message},
					At:      time.Now().Unix(),
				})

				log.Printf("[WEAVER] Redo projection verification failed: %s", verification.Message)
			} else {
				log.Printf("Redo verification succeeded for transaction %s", tx.ID)

				// Add audit entry
				auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
					Phase:   "Verify",
					Action:  "Validate",
					Outcome: "Success",
					Detail:  "Redo projection matched expectations",
					Meta:    map[string]string{"tx": string(tx.ID)},
					At:      time.Now().Unix(),
				})
			}
		} else {
			log.Printf("Failed to read post-snapshot for redo verification of transaction %s: %v", tx.ID, err)

			// Add audit entry
			auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
				Phase:   "Verify",
				Action:  "Validate",
				Outcome: "Warning",
				Detail:  fmt.Sprintf("Failed to read post-snapshot: %v", err),
				Meta:    map[string]string{"tx": string(tx.ID)},
				At:      time.Now().Unix(),
			})
		}
	}

	// Restore to Undo Stack
	log.Printf("Moving transaction %s from redo back to undo stack", tx.ID)
	e.history.PushBack(tx)

	// Add audit entry
	auditRecord.Entries = append(auditRecord.Entries, AuditEntryV2{
		Phase:   "History",
		Action:  "Move",
		Outcome: "Success",
		Detail:  fmt.Sprintf("Transaction %s moved from redo back to undo", tx.ID),
		Meta:    map[string]string{"tx": string(tx.ID)},
		At:      time.Now().Unix(),
	})

	// Update final result
	auditRecord.Result = AuditResult{
		Status: "Committed",
	}

	// Update audit record with transaction ID
	auditRecord.TransactionID = string(tx.ID)

	log.Printf("Successfully redone transaction %s", tx.ID)
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Redone tx: %s", tx.ID),
		Transaction: tx,
		Audit:       convertAuditRecordToLegacy(auditRecord),
	}, nil
}

// GetHistory 获取历史管理器 (用于 Reverse Bridge)
func (e *ShadowEngine) GetHistory() History {
	return e.history
}

// HashProof generates a hash of the proof object
func HashProof(p *Proof) string {
	b, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error marshaling proof: %v", err)
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Convert ResolvedFact to Editor Operation for DAG
func convertFactToOp(f ResolvedFact) editor.ResolvedOperation {
	opID := editor.OperationID(fmt.Sprintf("fact_%d", time.Now().UnixNano()))
	bufferID := editor.BufferID(f.Anchor.PaneID)
	anchor := editor.Cursor{Row: f.Anchor.Line, Col: f.Anchor.Start}

	switch f.Kind {
	case FactInsert:
		return &editor.InsertOperation{
			ID:     opID,
			Buffer: bufferID,
			At:     anchor,
			Text:   f.Payload.Text,
		}
	case FactDelete:
		return &editor.DeleteOperation{
			ID:     opID,
			Buffer: bufferID,
			Range: editor.TextRange{
				Start: anchor,
				End:   editor.Cursor{Row: f.Anchor.Line, Col: f.Anchor.End},
			},
			DeletedText: f.Payload.OldText,
		}
	case FactReplace:
		// Replace = Delete + Insert
		delOp := &editor.DeleteOperation{
			ID:     editor.OperationID(fmt.Sprintf("%s_del", opID)),
			Buffer: bufferID,
			Range: editor.TextRange{
				Start: anchor,
				End:   editor.Cursor{Row: f.Anchor.Line, Col: f.Anchor.End},
			},
			DeletedText: f.Payload.OldText,
		}
		insOp := &editor.InsertOperation{
			ID:     editor.OperationID(fmt.Sprintf("%s_ins", opID)),
			Buffer: bufferID,
			At:     anchor,
			Text:   f.Payload.NewText,
		}
		return &editor.CompositeOperation{
			ID:       opID,
			Children: []editor.ResolvedOperation{delOp, insOp},
		}
	case FactMove:
		// For now, treat Move as incomplete if we don't have To position
		return nil
	default:
		return nil
	}
}
