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
	var audit []AuditEntry

	// Phase 6.3: Temporal Adjudication (World Drift Check)
	// Engine owns the authority to reject execution if current reality != intent's expectation.
	if intent.GetSnapshotHash() != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			if string(current.Hash) != intent.GetSnapshotHash() {
				audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Rejected: World Drift detected"})
				return &Verdict{
					Kind:    VerdictRejected,
					Safety:  SafetyUnsafe,
					Message: "World drift detected",
					Audit:   audit,
				}, ErrWorldDrift
			}
			audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Success: Time consistency verified"})
		}
		// If Reality check fails (IO error), we might proceed with warning or fail fast.
		// For now, assume if we can't read reality, it's a structural error but not necessarily drift.
	}

	// 1. Handle Undo/Redo explicitly
	kind := intent.GetKind()
	if kind == IntentUndo {
		return e.performUndo()
	}
	if kind == IntentRedo {
		return e.performRedo()
	}

	// 2. Plan: Generate Facts
	facts, inverseFacts, err := e.planner.Build(intent, snapshot)
	if err != nil {
		audit = append(audit, AuditEntry{Step: "Plan", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Plan", Result: "Success"})

	// [Phase 5.1] 4. Resolve: 定位权移交
	// [Phase 5.4] 包含 Reconciliation 检查
	// [Phase 6.3] 包含 World Drift 检查 (SnapshotHash)
	resolvedFacts, err := e.resolver.ResolveFacts(facts, intent.GetSnapshotHash())
	if err != nil {
		audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Resolve", Result: "Success"})

	// [Phase 7] Determine overall safety
	safety := SafetyExact
	for _, rf := range resolvedFacts {
		if rf.Safety > safety {
			safety = rf.Safety
		}
	}

	if safety == SafetyFuzzy && !intent.IsPartialAllowed() {
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
	}

	// 3. Create Transaction
	txID := TransactionID(fmt.Sprintf("tx-%d", time.Now().UnixNano()))
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
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
		audit = append(audit, AuditEntry{Step: "Project", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})
	tx.Applied = true

	// [Phase 7] Capture PostSnapshotHash for Undo verification
	var postSnap Snapshot
	if e.reality != nil {
		var err error
		postSnap, err = e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			tx.PostSnapshotHash = string(postSnap.Hash)
			audit = append(audit, AuditEntry{Step: "Record", Result: fmt.Sprintf("PostHash: %s", tx.PostSnapshotHash)})
		}
	}

	// [Phase 9] Verify that the projection achieved the expected result
	if e.projection != nil && e.reality != nil {
		verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
		if !verification.OK {
			audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Verification failed: %s", verification.Message)})
			// For now, we still consider this applied but log the verification issue
			log.Printf("[WEAVER] Projection verification failed: %s", verification.Message)
		} else {
			audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Projection matched expectations"})
		}
	}

	// 6. Update History
	if len(facts) > 0 {
		e.history.Push(tx)
	}

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: tx,
		Safety:      safety,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performUndo() (*Verdict, error) {
	tx := e.history.PopUndo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to undo"}, nil
	}

	// [Phase 7] Axiom 7.5: Undo Is Verified Replay
	if tx.PostSnapshotHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != tx.PostSnapshotHash {
			// Put it back to undo stack since we didn't apply it
			e.history.PushBack(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot undo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
	}

	var audit []AuditEntry
	audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Undo context verified"})

	// [Phase 5.1] Resolve InverseFacts
	// [Phase 6.3] Use recorded PostHash if available (passed as expectedHash)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.InverseFacts, tx.PostSnapshotHash)
	if err != nil {
		e.history.PushBack(tx)
		return nil, err
	}
	audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Success: %d facts", len(resolvedFacts))})

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if len(resolvedFacts) > 0 {
		log.Printf("[WEAVER] Undo: Applying %d inverse facts. Text length: %d chars.", len(resolvedFacts), len(resolvedFacts[0].Payload.Text))
	}
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
		e.history.PushBack(tx)
		return nil, err
	}
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})

	// [Phase 9] Verify undo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Undo verification failed: %s", verification.Message)})
				log.Printf("[WEAVER] Undo projection verification failed: %s", verification.Message)
			} else {
				audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Undo projection matched expectations"})
			}
		}
	}

	// Move to Redo Stack
	e.history.AddRedo(tx)

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Undone tx: %s", tx.ID),
		Transaction: tx,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performRedo() (*Verdict, error) {
	tx := e.history.PopRedo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to redo"}, nil
	}

	// [Phase 7] Redo verification (must match Pre-state)
	preHash := tx.Intent.GetSnapshotHash()
	if preHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != preHash {
			e.history.AddRedo(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot redo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
	}

	// [Phase 5.1] Resolve Facts
	resolvedFacts, err := e.resolver.ResolveFacts(tx.Facts, preHash)
	if err != nil {
		e.history.AddRedo(tx)
		return nil, err
	}

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
		e.history.AddRedo(tx)
		return nil, err
	}

	// [Phase 9] Verify redo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("[WEAVER] Redo projection verification failed: %s", verification.Message)
			} else {
				// Verification successful
			}
		}
	}

	// Restore to Undo Stack
	e.history.PushBack(tx)

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
