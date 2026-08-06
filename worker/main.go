package main

import (
	"context"
	"net/http"
)

func runServer(ctx context.Context, handler func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
