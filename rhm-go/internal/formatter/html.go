package formatter

import (
	"bytes"
	"html/template"
	"rhm-go/core/narrative"
)

const htmlTemplateStr = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8"/>
<title>RHM Resolution Report</title>
<style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; margin: 40px; line-height: 1.6; color: #333; }
    h1 { border-bottom: 2px solid #eee; padding-bottom: 10px; }
    .summary-box { background: #f4fcf4; border: 1px solid #c3e6cb; padding: 15px; border-radius: 5px; color: #155724; margin-bottom: 30px; }
    .cost-badge { background: #e2e3e5; color: #383d41; padding: 2px 6px; border-radius: 4px; font-weight: bold; font-family: monospace; }
    .step { border-left: 4px solid #007bff; padding-left: 15px; margin-bottom: 30px; }
    .step h3 { margin-top: 0; color: #0056b3; }
    .decision-box { background: #f8f9fa; padding: 15px; border-radius: 5px; border: 1px solid #ddd; }
    .rejected-table { width: 100%; border-collapse: collapse; margin-top: 15px; font-size: 0.9em; }
    .rejected-table th { text-align: left; border-bottom: 2px solid #ddd; padding: 8px; color: #666; }
    .rejected-table td { border-bottom: 1px solid #eee; padding: 8px; }
    .reason { color: #888; font-style: italic; }
</style>
</head>
<body>

<h1>RHM Causal Resolution Report</h1>

<div class="summary-box">
    <strong>Summary:</strong> {{.Summary}}<br>
    <strong>Total Semantic Cost:</strong> {{.TotalCost}} SLU
</div>

<h2>Decision Trail</h2>

{{range .Steps}}
<div class="step">
    <h3>Step: {{.ProblemContext}}</h3>
    <div class="decision-box">
        <div><strong>Selected Strategy:</strong> <code>{{.Decision}}</code></div>
        <div><strong>Cost:</strong> <span class="cost-badge">{{.DecisionCost}}</span></div>
    </div>

    {{if .Rejected}}
    <h4>Alternatives Rejected</h4>
    <table class="rejected-table">
        <thead>
            <tr><th>Strategy</th><th>Cost</th><th>Reason</th></tr>
        </thead>
        <tbody>
        {{range .Rejected}}
        <tr>
            <td><code>{{.Description}}</code></td>
            <td>{{.Cost}}</td>
            <td class="reason">{{.Reason}}</td>
        </tr>
        {{end}}
        </tbody>
    </table>
    {{end}}
</div>
{{end}}

</body>
</html>
`

func ToHTML(n narrative.Narrative) (string, error) {
	tpl, err := template.New("report").Parse(htmlTemplateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, n); err != nil {
		return "", err
	}
	return buf.String(), nil
}
