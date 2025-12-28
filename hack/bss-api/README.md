# bss-api

Brad's super simple API is a toy REST API that simulates some external api.

## Running Locally

```
go run main.go
```
and then:

```
 curl -X POST localhost:8880/api/v1/clusters -H "Content-Type: application/json" \
  -d '{"name":"demo","replicas":3,"version":"1.0.0"}'
```
should return something like

```
{"id":"0ac1339d-8d07-4075-9f19-18df006d0643","name":"demo","replicas":3,"version":"1.0.0","state":"creating","readyReplicas":0,"createdAt":"2025-12-28T15:04:10.514455284+10:00","lastUpdateTime":"2025-12-28T15:04:10.514455512+10:00"}
```

## Docker publish

docker build -t bss-api:1.0.0 .
