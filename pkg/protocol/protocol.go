package protocol

// Anchor is "I mean this text, not the cursor"
type Anchor struct {
	PaneID   string  `json:"pane_id"`
	LineHint int     `json:"line_hint"`
	LineHash string  `json:"line_hash"`
	Cursor   *[2]int `json:"cursor_hint,omitempty"`
}

type Range struct {
	Anchor      Anchor `json:"anchor"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Text        string `json:"text"`
}

type Fact struct {
	Kind        string                 `json:"kind"` // delete / insert / replace
	Target      Range                  `json:"target"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	SideEffects []string               `json:"side_effects,omitempty"`
}

type ActionRecord struct {
	Fact    Fact `json:"fact"`
	Inverse Fact `json:"inverse"`
}
