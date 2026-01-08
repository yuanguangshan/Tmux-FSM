# Weaver System Design Axioms (Phases 5-7)

This document consolidates the core architectural principles (Axioms) that govern the Weaver system as of Phase 7.

## Phase 5: Semantic Foundations
- **Axiom 5.1: Anchor Primacy**: Locations are never hardcoded; they are resolved from semantic descriptions at the last possible microsecond.
- **Axiom 5.2: Planner Detachment**: The Planner generates "what should happen" based on intent, oblivious to physical coordinates.
- **Axiom 5.3: Inverse Integrity**: Every fact generated must store its inverse content (captured from reality) during the Resolve phase to ensure lossless Undo.

## Phase 6: Temporal Freezing
- **Axiom 6.1: Snapshot Atomicity**: All planning for a single intent must occur against a single, frozen world snapshot.
- **Axiom 6.2: Universal Intent Hash**: Every intent carries the hash of the world it was born in.
- **Axiom 6.3: Reality Readers**: Resolvers should prioritize reading from provided snapshots over direct IO.

## Phase 7: Deterministic Replay & Temporal Integrity
- **Axiom 7.1: Intent Is Timeless, Execution Is Temporal**: Intents are descriptions; they only enter history when verified against a specific world state.
- **Axiom 7.2: Replay Is Re-Execution**: History is an auditable chain of causal effects (Intent + Snapshot -> Verdict), not a buffer of restored text.
- **Axiom 7.3: Determinism Is a Contract**: In identical conditions (Intent + Hash + Version), the result must be identical.
- **Axiom 7.4: World Drift Is Final**: If the world has moved, the system must refuse execution. No guessing, no silent fallbacks.
- **Axiom 7.5: Undo Is Verified Replay**: Undo must verify the "Post-State" hash before attempting to invert an action.
- **Axiom 7.6: Engine Owns Temporal Authority**: Only the Engine can adjudicate "World Drift." Resolvers merely follow the coordinates of the chosen reality.
- **Axiom 7.7: Two-Phase Replay**: To prevent partial state corruption, all anchors in a transaction must be successfully resolved before any single fact in that transaction is projected.

---
*End of Axioms v0.7.0*
