package core

type DiffKind int

const (
    DiffInsert DiffKind = iota
    DiffDelete
    DiffModify
)

type SnapshotDiff struct {
    LineID  LineID
    Before *LineSnapshot
    After  *LineSnapshot
    Change DiffKind
}

func DiffSnapshot(pre, post Snapshot) []SnapshotDiff {
    diffs := []SnapshotDiff{}

    // deletions & modifications
    for id, preIdx := range pre.Index {
        preLine := pre.Lines[preIdx]
        postIdx, ok := post.Index[id]

        if !ok {
            diffs = append(diffs, SnapshotDiff{
                LineID: id,
                Before: &preLine,
                After:  nil,
                Change: DiffDelete,
            })
            continue
        }

        postLine := post.Lines[postIdx]
        if preLine.Hash != postLine.Hash {
            diffs = append(diffs, SnapshotDiff{
                LineID: id,
                Before: &preLine,
                After:  &postLine,
                Change: DiffModify,
            })
        }
    }

    // insertions
    for id, postIdx := range post.Index {
        if _, ok := pre.Index[id]; !ok {
            postLine := post.Lines[postIdx]
            diffs = append(diffs, SnapshotDiff{
                LineID: id,
                Before: nil,
                After:  &postLine,
                Change: DiffInsert,
            })
        }
    }

    return diffs
}