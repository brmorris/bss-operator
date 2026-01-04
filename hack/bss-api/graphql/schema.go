package graphql

import (
	"github.com/brmorris/bss-operator/hack/bss-api/store"
	"github.com/graphql-go/graphql"
)

var clusterType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Cluster",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"replicas": &graphql.Field{
				Type: graphql.Int,
			},
			"version": &graphql.Field{
				Type: graphql.String,
			},
			"state": &graphql.Field{
				Type: graphql.String,
			},
			"readyReplicas": &graphql.Field{
				Type: graphql.Int,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"lastUpdateTime": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)

func NewSchema(store *store.MemoryStore) (graphql.Schema, error) {
	queryType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"cluster": &graphql.Field{
					Type:        clusterType,
					Description: "Get cluster by ID",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						id, ok := p.Args["id"].(string)
						if !ok {
							return nil, nil
						}
						cluster, exists := store.Get(id)
						if !exists {
							return nil, nil
						}
						return cluster, nil
					},
				},
				"clusters": &graphql.Field{
					Type:        graphql.NewList(clusterType),
					Description: "List all clusters",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return store.List(), nil
					},
				},
			},
		},
	)

	mutationType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"createCluster": &graphql.Field{
					Type:        clusterType,
					Description: "Create a new cluster",
					Args: graphql.FieldConfigArgument{
						"name": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"replicas": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Int),
						},
						"version": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						name, _ := p.Args["name"].(string)
						replicas, _ := p.Args["replicas"].(int)
						version, _ := p.Args["version"].(string)

						return createCluster(store, name, int32(replicas), version), nil
					},
				},
				"deleteCluster": &graphql.Field{
					Type:        graphql.Boolean,
					Description: "Delete a cluster by ID",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						id, ok := p.Args["id"].(string)
						if !ok {
							return false, nil
						}
						return deleteCluster(store, id), nil
					},
				},
			},
		},
	)

	return graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)
}
