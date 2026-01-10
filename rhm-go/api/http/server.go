package httpapi

import (
	"encoding/json"
	"net/http"
	"rhm-go/core/solver"
	"rhm-go/internal/formatter"
	"rhm-go/internal/loader"
)

func Start(addr string) {
	http.HandleFunc("/solve", func(w http.ResponseWriter, r *http.Request) {
		dag, tipA, tipB := loader.LoadDemoScenario() // Mocked loader
		plan := solver.Solve(dag, tipA, tipB)

		if r.URL.Query().Get("format") == "markdown" {
			w.Header().Set("Content-Type", "text/markdown")
			w.Write([]byte(formatter.ToMarkdown(plan.Narrative)))
			return
		}
		json.NewEncoder(w).Encode(plan)
	})
	http.ListenAndServe(addr, nil)
}
