# Collaborative Editing Model: Operation DAG

## Overview
This document outlines the foundational principles for the collaborative editing model in Tmux-FSM, based on **Operation DAGs** (Directed Acyclic Graphs). This approach departs from traditional linear undo/redo stacks and OT/CRDT approaches by treating the edit history as a causal graph of immutable semantic operations.

## Core Concepts

### 1. Operation DAG vs. Linear History
*   **Linear History (Legacy)**: A stack of states or operations (Undo/Redo). Branching (undoing then doing new work) destroys 'future' history.
*   **Operation DAG (Weaver)**: Every operation is a node. 
    *   **Node**: Contains a `ResolvedOperation` (Atomic, Semantic).
    *   **Edges**: Represent causal dependencies (`Parent` pointers).
    *   **Immutability**: Once created, a node is never modified.
    *   **Branching**: "Undoing" is simply moving the current view pointer to an ancestor. "Redoing" is moving it to a descendant. Creating a new edit from an old state creates a **New Branch**.

### 2. Semantic Diffing
Since edits are semantic (e.g., "Delete Function Foo", "Rename Variable X"), diffing is structured:
*   **Diff(A, B)**: The set of DAG nodes present in B's ancestry but not in A's.
*   **Path**: Topological ordering of these nodes represents the "Patch".

### 3. Collaboration & Merging
When two users edit concurrently:
*   User A creates Node `nA` with parent `P`.
*   User B creates Node `nB` with parent `P`.
*   **State divergence**: `Tips = {nA, nB}`.

#### Automatic Merging
To converge, we create a **Merge Node** `nM`:
*   `nM.Parents = {nA, nB}`.
*   `nM.Operation` = Result of reconciling `nA` and `nB`.

#### Conflict Detection
Unlike text-based merge (which fails on overlapping lines), we use **Semantic Collision**:
1.  **Spatial Conflict**: Do operations touch the same `LineID` ranges?
2.  **Semantic Conflict**: Does `nB` modify a variable that `nA` deleted?
3.  **Resolution Strategy**:
    *   **Conservative**: If collision detected, prompt user (Manual Merge).
    *   **Optimistic**: If spatially disjoint, apply both.

### 4. Git Integration
The Operation DAG maps naturally to Git's object model but at a finer granularity:
*   **Commit** ≈ Checkpoint of DAG state.
*   **Review**: Instead of reviewing "Changed lines 10-12", review "Refactor Function X (composed of nodes N1..N5)".

### 5. Implementation Status (Phase 7)
*   [x] **DAG Structure**: `editor/dag.go` defined `DAGNode` and `OperationDAG`.
*   [x] **Traversal Logic**: `editor/dag_traversal.go` implements `GetAncestors`, `FindLCA`, `Diff`.
*   [x] **Shadow Engine Integration**: `ShadowEngine` maintains a live DAG of local edits.
*   [ ] **Merge Logic**: To be implemented in Phase 8 (`editor/dag_merge.go`).

## Future Work
*   **Rebase**: Reparenting a chain of nodes onto a new base.
*   **Squash**: Collapsing a subgraph into a single composite semantic operation.
