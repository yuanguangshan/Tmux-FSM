package formatter

import (
	"fmt"
	"rhm-go/core/narrative"
	"strings"
)

func ToMarkdown(n narrative.Narrative) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", n.Summary))
	sb.WriteString(fmt.Sprintf("**Total Semantic Cost:** `%d SLU`\n\n", n.TotalCost))
	sb.WriteString("## Decision Trail\n\n")

	for i, step := range n.Steps {
		sb.WriteString(fmt.Sprintf("### Step %d: %s\n", i+1, step.ProblemContext))
		sb.WriteString(fmt.Sprintf("> **Selected:** `%s` (Cost %d)\n\n", step.Decision, step.DecisionCost))

		if len(step.Rejected) > 0 {
			sb.WriteString("| Alternative | Cost | Reason |\n|---|---|---|\n")
			for _, alt := range step.Rejected {
				sb.WriteString(fmt.Sprintf("| `%s` | %d | %s |\n", alt.Description, alt.Cost, alt.Reason))
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
