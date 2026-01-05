/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GraphQLClient is a client for interacting with the BSS API GraphQL endpoint
type GraphQLClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient(endpoint string) *GraphQLClient {
	return &GraphQLClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// Execute executes a GraphQL query and returns the response
func (c *GraphQLClient) Execute(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	req := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var graphqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return &graphqlResp, fmt.Errorf("graphql errors: %v", graphqlResp.Errors)
	}

	return &graphqlResp, nil
}

// ClusterData represents cluster data from GraphQL
type ClusterData struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Replicas       int32     `json:"replicas"`
	Version        string    `json:"version"`
	State          string    `json:"state"`
	ReadyReplicas  int32     `json:"readyReplicas"`
	CreatedAt      time.Time `json:"createdAt"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// GetCluster retrieves a single cluster by ID
func (c *GraphQLClient) GetCluster(id string) (*ClusterData, error) {
	query := `
		query GetCluster($id: String!) {
			cluster(id: $id) {
				id
				name
				replicas
				version
				state
				readyReplicas
				createdAt
				lastUpdateTime
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	resp, err := c.Execute(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Cluster *ClusterData `json:"cluster"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster data: %w", err)
	}

	return result.Cluster, nil
}

// ListClusters retrieves all clusters
func (c *GraphQLClient) ListClusters() ([]*ClusterData, error) {
	query := `
		query ListClusters {
			clusters {
				id
				name
				replicas
				version
				state
				readyReplicas
				createdAt
				lastUpdateTime
			}
		}
	`

	resp, err := c.Execute(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Clusters []*ClusterData `json:"clusters"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clusters data: %w", err)
	}

	return result.Clusters, nil
}

// CreateCluster creates a new cluster
func (c *GraphQLClient) CreateCluster(name string, replicas int32, version string) (*ClusterData, error) {
	query := `
		mutation CreateCluster($name: String!, $replicas: Int!, $version: String!) {
			createCluster(name: $name, replicas: $replicas, version: $version) {
				id
				name
				replicas
				version
				state
				readyReplicas
				createdAt
				lastUpdateTime
			}
		}
	`

	variables := map[string]interface{}{
		"name":     name,
		"replicas": replicas,
		"version":  version,
	}

	resp, err := c.Execute(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		CreateCluster *ClusterData `json:"createCluster"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created cluster data: %w", err)
	}

	return result.CreateCluster, nil
}

// DeleteCluster deletes a cluster by ID
func (c *GraphQLClient) DeleteCluster(id string) (bool, error) {
	query := `
		mutation DeleteCluster($id: String!) {
			deleteCluster(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	resp, err := c.Execute(query, variables)
	if err != nil {
		return false, err
	}

	var result struct {
		DeleteCluster bool `json:"deleteCluster"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return false, fmt.Errorf("failed to unmarshal delete result: %w", err)
	}

	return result.DeleteCluster, nil
}
