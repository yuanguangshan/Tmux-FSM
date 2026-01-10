package narrative

type Narrative struct {
	Summary   string         `json:"summary"`
	Steps     []DecisionStep `json:"steps"`
	TotalCost int            `json:"totalCost"`
}

type DecisionStep struct {
	ProblemContext string                `json:"problem"`
	Decision       string                `json:"decision"`
	DecisionCost   int                   `json:"cost"`
	Rejected       []RejectedAlternative `json:"rejected,omitempty"`
}

type RejectedAlternative struct {
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Reason      string `json:"reason"`
}
