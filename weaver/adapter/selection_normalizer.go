package adapter

import (
	"fmt"
	"sort"
	"tmux-fsm/weaver/core"
)

// Selection represents a user selection with start and end positions
type Selection struct {
	LineID core.LineID
	Anchor int
	Focus  int
}

type normRange struct {
	start int
	end   int
}

// NormalizeSelections normalizes user selections into a safe list of anchors
func NormalizeSelections(selections []Selection) ([]core.Anchor, error) {
	if len(selections) == 0 {
		return nil, nil
	}

	// 1️⃣ canonicalize + group by line
	group := make(map[core.LineID][]normRange)

	for _, sel := range selections {
		start := sel.Anchor
		end := sel.Focus
		if start > end {
			start, end = end, start
		}
		group[sel.LineID] = append(group[sel.LineID], normRange{
			start: start,
			end:   end,
		})
	}

	var anchors []core.Anchor

	// 2️⃣ process per line
	for lineID, ranges := range group {
		// 3️⃣ sort by start, then end
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i].start == ranges[j].start {
				return ranges[i].end < ranges[j].end
			}
			return ranges[i].start < ranges[j].start
		})

		// 4️⃣ reject overlap / containment
		var prev *normRange
		for i := range ranges {
			curr := &ranges[i]
			if prev != nil {
				if curr.start < prev.end {
					return nil, fmt.Errorf(
						"overlapping selections on line %s [%d,%d] vs [%d,%d]",
						lineID,
						prev.start, prev.end,
						curr.start, curr.end,
					)
				}
			}
			prev = curr
		}

		// 5️⃣ convert to anchors
		for _, r := range ranges {
			anchors = append(anchors, core.Anchor{
				LineID: lineID,
				Kind:   core.AnchorAbsolute,
				Ref:    []int{r.start, r.end}, // Store as [start, end] pair
			})
		}
	}

	return anchors, nil
}