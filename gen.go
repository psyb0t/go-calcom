// Package calcom is a Go client for the Cal.com API v2,
// generated with oapi-codegen from the OpenAPI document Cal.com publishes.
package calcom

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/patch-spec.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.patched.json
