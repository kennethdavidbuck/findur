// Package provider owns the reproducible, OAuth-bearer-only SnapTrade inventory overlay.
package provider

//go:generate go run ./cmd/materialize
//go:generate go tool oapi-codegen --config=oapi-codegen.yaml openapi.yaml
