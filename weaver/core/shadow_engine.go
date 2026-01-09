package core

import (
	"fmt"
	"log"
	"time"
)

// ShadowEngine 核心执行引擎
// 负责处理 Intent，生成并应用 Transaction，维护 History
type ShadowEngine struct {
	planner    Planner
	history    History
	resolver   AnchorResolver
	projection Projection
	reality    RealityReader
}

func NewShadowEngine(planner Planner, resolver AnchorResolver, projection Projection, reality RealityReader) *ShadowEngine {
	return &ShadowEngine{
		planner:    planner,
		history:    NewInMemoryHistory(100),
		resolver:   resolver,
		projection: projection,
		reality:    reality,
	}
}

func (e *ShadowEngine) ApplyIntent(intent Intent, snapshot Snapshot) (*Verdict, error) {
	log.Printf("Applying intent: Kind=%d, PaneID=%s, SnapshotHash=%s",
		intent.GetKind(), intent.GetPaneID(), intent.GetSnapshotHash())

	var audit []AuditEntry

	// Phase 6.3: Temporal Adjudication (World Drift Check)
	// Engine owns the authority to reject execution if current reality != intent's expectation.
	if intent.GetSnapshotHash() != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			if string(current.Hash) != intent.GetSnapshotHash() {
				log.Printf("World drift detected: expected %s, got %s", intent.GetSnapshotHash(), string(current.Hash))
				audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Rejected: World Drift detected"})
				return &Verdict{
					Kind:    VerdictRejected,
					Safety:  SafetyUnsafe,
					Message: "World drift detected",
					Audit:   audit,
				}, ErrWorldDrift
			}
			log.Printf("Time consistency verified for intent in pane %s", intent.GetPaneID())
			audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Success: Time consistency verified"})
		} else {
			log.Printf("Could not read current reality for pane %s: %v", intent.GetPaneID(), err)
		}
		// If Reality check fails (IO error), we might proceed with warning or fail fast.
		// For now, assume if we can't read reality, it's a structural error but not necessarily drift.
	}

	// 1. Handle Undo/Redo explicitly
	kind := intent.GetKind()
	if kind == IntentUndo {
		log.Printf("Processing undo intent for pane %s", intent.GetPaneID())
		return e.performUndo()
	}
	if kind == IntentRedo {
		log.Printf("Processing redo intent for pane %s", intent.GetPaneID())
		return e.performRedo()
	}

	// 2. Plan: Generate Facts
	log.Printf("Planning facts for intent in pane %s", intent.GetPaneID())
	facts, inverseFacts, err := e.planner.Build(intent, snapshot)
	if err != nil {
		log.Printf("Failed to plan facts for intent in pane %s: %v", intent.GetPaneID(), err)
		audit = append(audit, AuditEntry{Step: "Plan", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	log.Printf("Successfully planned %d facts for intent in pane %s", len(facts), intent.GetPaneID())
	audit = append(audit, AuditEntry{Step: "Plan", Result: "Success"})

	// [Phase 5.1] 4. Resolve: 定位权移交
	// [Phase 5.4] 包含 Reconciliation 检查
	// [Phase 6.3] 包含 World Drift 检查 (SnapshotHash)
	log.Printf("Resolving facts for intent in pane %s", intent.GetPaneID())
	resolvedFacts, err := e.resolver.ResolveFacts(facts, intent.GetSnapshotHash())
	if err != nil {
		log.Printf("Failed to resolve facts for intent in pane %s: %v", intent.GetPaneID(), err)
		audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	log.Printf("Successfully resolved %d facts for intent in pane %s", len(resolvedFacts), intent.GetPaneID())
	audit = append(audit, AuditEntry{Step: "Resolve", Result: "Success"})

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
		return &Verdict{
			Kind:    VerdictRejected,
			Safety:  SafetyUnsafe,
			Message: "Fuzzy resolution disallowed by policy",
			Audit:   audit,
		}, ErrWorldDrift // Or a new error like ErrSafetyViolation
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

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot := snapshot

	// 5. Project: Execute
	log.Printf("Projecting %d resolved facts for intent in pane %s", len(resolvedFacts), intent.GetPaneID())
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to project facts for intent in pane %s: %v", intent.GetPaneID(), err)
		audit = append(audit, AuditEntry{Step: "Project", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	log.Printf("Successfully projected facts for intent in pane %s", intent.GetPaneID())
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})
	tx.Applied = true

	// [Phase 7] Capture PostSnapshotHash for Undo verification
	var postSnap Snapshot
	if e.reality != nil {
		var err error
		postSnap, err = e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			tx.PostSnapshotHash = string(postSnap.Hash)
			log.Printf("Captured post-snapshot hash %s for transaction %s", tx.PostSnapshotHash, txID)
			audit = append(audit, AuditEntry{Step: "Record", Result: fmt.Sprintf("PostHash: %s", tx.PostSnapshotHash)})
		} else {
			log.Printf("Failed to capture post-snapshot for transaction %s: %v", txID, err)
		}
	}

	// [Phase 9] Verify that the projection achieved the expected result
	if e.projection != nil && e.reality != nil {
		verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
		if !verification.OK {
			log.Printf("Projection verification failed for transaction %s: %s", txID, verification.Message)
			audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Verification failed: %s", verification.Message)})
			// For now, we still consider this applied but log the verification issue
			log.Printf("[WEAVER] Projection verification failed: %s", verification.Message)
		} else {
			log.Printf("Projection verification succeeded for transaction %s", txID)
			audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Projection matched expectations"})
		}
	}

	// 6. Update History
	if len(facts) > 0 {
		log.Printf("Pushing transaction %s to history for pane %s", txID, intent.GetPaneID())
		e.history.Push(tx)
	}

	log.Printf("Successfully applied intent for pane %s, transaction %s", intent.GetPaneID(), txID)
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: tx,
		Safety:      safety,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performUndo() (*Verdict, error) {
	log.Printf("Starting undo operation")
	tx := e.history.PopUndo()
	if tx == nil {
		log.Printf("No transaction to undo")
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to undo"}, nil
	}

	log.Printf("Attempting to undo transaction %s for pane %s", tx.ID, tx.Intent.GetPaneID())

	// [Phase 7] Axiom 7.5: Undo Is Verified Replay
	if tx.PostSnapshotHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != tx.PostSnapshotHash {
			log.Printf("World drift detected during undo: expected %s, got %s", tx.PostSnapshotHash, string(current.Hash))
			// Put it back to undo stack since we didn't apply it
			e.history.PushBack(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot undo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
		log.Printf("Undo context verified for transaction %s", tx.ID)
	}

	var audit []AuditEntry
	audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Undo context verified"})

	// [Phase 5.1] Resolve InverseFacts
	// [Phase 6.3] Use recorded PostHash if available (passed as expectedHash)
	log.Printf("Resolving %d inverse facts for undo of transaction %s", len(tx.InverseFacts), tx.ID)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.InverseFacts, tx.PostSnapshotHash)
	if err != nil {
		log.Printf("Failed to resolve inverse facts for undo of transaction %s: %v", tx.ID, err)
		e.history.PushBack(tx)
		return nil, err
	}
	log.Printf("Successfully resolved %d inverse facts for undo of transaction %s", len(resolvedFacts), tx.ID)
	audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Success: %d facts", len(resolvedFacts))})

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		log.Printf("Failed to capture pre-snapshot for undo of transaction %s: %v", tx.ID, err)
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if len(resolvedFacts) > 0 {
		log.Printf("[WEAVER] Undo: Applying %d inverse facts for transaction %s. Text length: %d chars.",
			len(resolvedFacts), tx.ID, len(resolvedFacts[0].Payload.Text))
	}
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to apply inverse facts for undo of transaction %s: %v", tx.ID, err)
		e.history.PushBack(tx)
		return nil, err
	}
	log.Printf("Successfully applied inverse facts for undo of transaction %s", tx.ID)
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})

	// [Phase 9] Verify undo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("Undo verification failed for transaction %s: %s", tx.ID, verification.Message)
				audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Undo verification failed: %s", verification.Message)})
				log.Printf("[WEAVER] Undo projection verification failed: %s", verification.Message)
			} else {
				log.Printf("Undo verification succeeded for transaction %s", tx.ID)
				audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Undo projection matched expectations"})
			}
		} else {
			log.Printf("Failed to read post-snapshot for undo verification of transaction %s: %v", tx.ID, err)
		}
	}

	// Move to Redo Stack
	log.Printf("Moving transaction %s from undo to redo stack", tx.ID)
	e.history.AddRedo(tx)

	log.Printf("Successfully undone transaction %s", tx.ID)
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Undone tx: %s", tx.ID),
		Transaction: tx,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performRedo() (*Verdict, error) {
	log.Printf("Starting redo operation")
	tx := e.history.PopRedo()
	if tx == nil {
		log.Printf("No transaction to redo")
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to redo"}, nil
	}

	log.Printf("Attempting to redo transaction %s for pane %s", tx.ID, tx.Intent.GetPaneID())

	// [Phase 7] Redo verification (must match Pre-state)
	preHash := tx.Intent.GetSnapshotHash()
	if preHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != preHash {
			log.Printf("World drift detected during redo: expected %s, got %s", preHash, string(current.Hash))
			e.history.AddRedo(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot redo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
		log.Printf("Redo context verified for transaction %s", tx.ID)
	}

	// [Phase 5.1] Resolve Facts
	log.Printf("Resolving %d facts for redo of transaction %s", len(tx.Facts), tx.ID)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.Facts, preHash)
	if err != nil {
		log.Printf("Failed to resolve facts for redo of transaction %s: %v", tx.ID, err)
		e.history.AddRedo(tx)
		return nil, err
	}
	log.Printf("Successfully resolved %d facts for redo of transaction %s", len(resolvedFacts), tx.ID)

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		log.Printf("Failed to capture pre-snapshot for redo of transaction %s: %v", tx.ID, err)
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	log.Printf("Projecting %d resolved facts for redo of transaction %s", len(resolvedFacts), tx.ID)
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		log.Printf("Failed to apply facts for redo of transaction %s: %v", tx.ID, err)
		e.history.AddRedo(tx)
		return nil, err
	}
	log.Printf("Successfully applied facts for redo of transaction %s", tx.ID)

	// [Phase 9] Verify redo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("Redo verification failed for transaction %s: %s", tx.ID, verification.Message)
				log.Printf("[WEAVER] Redo projection verification failed: %s", verification.Message)
			} else {
				log.Printf("Redo verification succeeded for transaction %s", tx.ID)
				// Verification successful
			}
		} else {
			log.Printf("Failed to read post-snapshot for redo verification of transaction %s: %v", tx.ID, err)
		}
	}

	// Restore to Undo Stack
	log.Printf("Moving transaction %s from redo back to undo stack", tx.ID)
	e.history.PushBack(tx)

	log.Printf("Successfully redone transaction %s", tx.ID)
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Redone tx: %s", tx.ID),
		Transaction: tx,
	}, nil
}

// GetHistory 获取历史管理器 (用于 Reverse Bridge)
func (e *ShadowEngine) GetHistory() History {
	return e.history
}
