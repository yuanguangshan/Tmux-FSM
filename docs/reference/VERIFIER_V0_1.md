# Verifier Protocol v0.1

## 1. Purpose

The verifier validates that a given application state
was produced **only** by a specific set of Facts,
under deterministic replay rules,
without trusting the editor engine or runtime environment.

This is a *verification protocol*, not an execution engine.

---

## 2. Trust Model

The verifier trusts:

- Fact DAG structure
- Canonical Fact payloads
- Deterministic replay rules

The verifier does NOT trust:

- Engine implementation
- Event IDs
- Timestamps
- CRDT positions
- Network order
- Local actor state

---

## 3. Data Model

### 3.1 Fact

```go
type Fact struct {
	ID        Hash
	Actor     ActorID
	Parents   []Hash
	Timestamp int64
	Payload   CanonicalSemanticEvent
	PolicyRef Hash
}
```

#### Fact ID

```
Fact.ID = hash(
  Actor,
  Parents,
  Timestamp,
  Payload,
  PolicyRef
)
```

- Fact.ID MUST be content-addressed
- Fact.ID MUST NOT depend on itself
- Fact.ID MUST be reproducible byte-for-byte

---

### 3.2 CanonicalSemanticEvent

```go
type CanonicalSemanticEvent struct {
	Actor         ActorID
	CausalParents []EventID
	Fact          semantic.BaseFact
}
```

The following fields are explicitly excluded:

- EventID
- Timestamp
- LocalParent
- CRDT internal metadata

---

## 4. Structural Invariants

### INV-1: Fact Self-Consistency

```
RecomputedHash(Fact) == Fact.ID
```

---

### INV-2: Parent Equivalence

```
Fact.Parents ≡ hash(Payload.CausalParents)
```

Fact DAG order MUST match semantic causal order.

---

### INV-3: DAG Acyclicity

All Facts MUST be topologically sortable.
Failure indicates invalid history.

---

## 5. Replay Rules

### 5.1 Determinism

Replay MUST be:

- Pure (no side effects)
- Deterministic
- Order-dependent only on Fact DAG

---

### 5.2 Order Resolution

Replay order is defined as:

1. Topological sort over Fact.Parents
2. Stable tie-breaking by Fact.ID

---

## 6. State Commitment

### 6.1 State Root Definition

```
StateRoot = hash(
  FinalState,
  LastFactID,
  FactCount
)
```

This prevents history erasure attacks.

---

## 7. Policy Rules (v0.1)

Policies may:

- Accept or reject Facts

Policies may NOT:

- Modify state
- Reorder Facts
- Inject events

Policy execution MUST be deterministic.

---

## 8. Verification Outcome

Verification succeeds iff:

- All invariants hold
- Replay completes
- Computed StateRoot matches expected root

Any violation results in verification failure.

---

## 9. Non-Goals

Verifier v0.1 does NOT address:

- Cryptographic signatures
- Key distribution
- Zero-knowledge proofs
- Partial verification

These are deferred to v0.2+.