package httpapi

import (
	"encoding/json"
	"net/http"
	"rhm-go/core/solver"
	"rhm-go/internal/formatter"
	"rhm-go/internal/loader"
)

func solveHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Load World (Mocked for demo)
	dag, tipA, tipB := loader.LoadDemoScenario()

	// 2. Run Engine
	plan := solver.Solve(dag, tipA, tipB)

	// 3. Render Response
	format := r.URL.Query().Get("format")

	switch format {
	case "markdown":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write([]byte(formatter.ToMarkdown(plan.Narrative)))
	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html, err := formatter.ToHTML(plan.Narrative)
		if err != nil {
			http.Error(w, "Template Error", 500)
			return
		}
		w.Write([]byte(html))
	default:
		// JSON Default
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(plan)
	}
}
