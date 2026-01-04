package main

import (
	"log"
	"net/http"

	"github.com/brmorris/bss-operator/hack/bss-api/api"
	bssGraphQL "github.com/brmorris/bss-operator/hack/bss-api/graphql"
	"github.com/brmorris/bss-operator/hack/bss-api/store"
	"github.com/graphql-go/handler"
)

func main() {
	store := store.NewMemoryStore()
	server := api.NewServer(store)

	// Create GraphQL schema
	schema, err := bssGraphQL.NewSchema(store)
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}

	// Create GraphQL handler
	graphqlHandler := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})

	mux := http.NewServeMux()

	// REST API endpoints
	mux.HandleFunc("POST /api/v1/clusters", server.CreateCluster)
	mux.HandleFunc("GET /api/v1/clusters/{id}", server.GetCluster)
	mux.HandleFunc("DELETE /api/v1/clusters/{id}", server.DeleteCluster)

	// GraphQL endpoint
	mux.Handle("/graphql", graphqlHandler)

	log.Println("BSS API listening on :8880")
	log.Println("REST API: http://localhost:8880/api/v1/clusters")
	log.Println("GraphQL: http://localhost:8880/graphql")
	log.Fatal(http.ListenAndServe(":8880", mux))
}
