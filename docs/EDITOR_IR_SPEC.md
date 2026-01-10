# Editor IR Design Specification

## 1. Overview
The Editor Intermediate Representation (IR) is the backbone of the Tmux-FSM's next-generation editing engine. It represents the editing history not as a linear sequence of states, but as a Directed Acyclic Graph (DAG) of atomic, semantic operations. This structure enables advanced features like non-linear undo/redo, collaborative editing, and semantic diffing.

## 2. Data Structure

### 2.1. DAG Node
Each node in the DAG represents an atomic edit operation.

```go
type DAGNode struct {
	ID        DAGNodeID          `json:"id"`        // Unique UUID
	Operation ResolvedOperation  `json:"operation"` // The atomic edit
	Parents   []DAGNodeID        `json:"parents"`   // Causal dependencies
	Timestamp int64              `json:"timestamp"` // Unix/Lamport timestamp
	Meta      map[string]string  `json:"meta"`      // Extensible metadata
}
```

### 2.2. Resolved Operation
The payload of a node is a `ResolvedOperation`, which is a strictly typed, location-aware description of the edit.

```go
type ResolvedOperation struct {
    Kind     ResolvedOperationKind // OpInsert, OpDelete, OpMove
    BufferID BufferID
    Anchor   Cursor                // Starting position
    // For Insert:
    Text     string
    // For Delete:
    Range       *TextRange
    DeletedText string             // Captured for reversibility
}
```

## 3. Serialization
The DAG is serialized to JSON. This format is human-readable and easy to parse, making it suitable for debugging, storage, and inter-process communication.

### 3.1. Schema
```json
{
  "nodes": {
    "node_123": {
      "id": "node_123",
      "operation": { ... },
      "parents": ["node_122"]
    },
    ...
  },
  "roots": ["node_0"],
  "tips": ["node_123"]
}
```

## 4. Semantic Diffing
Diffing in an Operation DAG is fundamentally different from text diffing. It answers the question: "What operations happened in Branch B that did not happen in Branch A?"

### 4.1. Algorithm
1.  **Identify Ancestry**: Compute the set of all ancestors for both Key nodes (Base and Target).
2.  **Set Subtraction**: `Diff = Ancestors(Target) - Ancestors(Base)`.
3.  **Topological Sort**: Order the resulting set of nodes by dependency to ensure a valid execution order.

### 4.2. Output
The output of a semantic diff is a "Patch" — a sequence of `ResolvedOperation`s. This patch can be applied to the Base state to reach the Target state (assuming no conflicts).

## 5. Git Integration Strategy
While the internal IR is a DAG, we can project this onto Git's version control model.

1.  **Commit Mapping**: A Git Commit corresponds to a snapshot of the DAG. The commit message can reference specific DAG Node IDs.
2.  **Semantic Blame**: Instead of line-based blame, we can trace the DAG backwards to find the node responsible for the current state of a text range.
3.  **Conflict Resolution**: When Git detects a merge conflict, we can use the DAG structure to identify if the conflict is purely textual or semantically non-colliding (e.g., disjoint edits), potentially resolving it automatically.

## 6. Future Extensions
*   **Signatures**: Cryptographic signing of DAG nodes for author verification.
*   **Compression**: Snapshotting state at intervals to avoid traversing the entire history.
*   **CRDT Integration**: If real-time character-by-character collaboration is needed, nodes can be CRDT operations.
