package main

import (
	"log"
	"net/http"

	"github.com/brmorris/bss-operator/hack/bss-api/api"
	"github.com/brmorris/bss-operator/hack/bss-api/store"
)

func main() {
	store := store.NewMemoryStore()
	server := api.NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/clusters", server.CreateCluster)
	mux.HandleFunc("GET /api/v1/clusters/{id}", server.GetCluster)
	mux.HandleFunc("DELETE /api/v1/clusters/{id}", server.DeleteCluster)

	log.Println("BSS API listening on :8880")
	log.Fatal(http.ListenAndServe(":8880", mux))
}
