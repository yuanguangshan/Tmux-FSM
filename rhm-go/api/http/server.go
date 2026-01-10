package httpapi

import (
	"fmt"
	"net/http"
)

func Start(addr string) {
	// Register handlers from handlers.go
	http.HandleFunc("/solve", solveHandler)

	// Add Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	fmt.Printf("🚀 RHM Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
