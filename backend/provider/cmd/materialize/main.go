// Command materialize projects the reviewed OAuth inventory operations and all
// of their transitive response schemas from the pinned official SnapTrade spec.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	upstreamPath = "upstream/api.yaml"
	overlayPath  = "oauth-bearer-overlay.yaml"
	outputPath   = "openapi.yaml"
	revision     = "9a7c008c1b0c7ea9f7944938556d04f66b381082"
	wantSHA256   = "6c6058c06284884dcf6f2710aa0c46c51701df0eeade1adb058de292e95b95f4"
)

type overlay struct {
	Upstream struct {
		Revision string `yaml:"revision"`
		SHA256   string `yaml:"sha256"`
	} `yaml:"upstream"`
	Operations []operationSelection `yaml:"operations"`
	OAuth      bearerSecurity       `yaml:"oauthBearer"`
}

type operationSelection struct {
	Path             string              `yaml:"path"`
	Method           string              `yaml:"method"`
	RemoveParameters []parameterSelector `yaml:"removeParameters"`
}

type parameterSelector struct {
	In   string `yaml:"in"`
	Name string `yaml:"name"`
}

type bearerSecurity struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Scheme string `yaml:"scheme"`
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	source, err := readPinnedSource()
	if err != nil {
		return err
	}
	selection, err := readOverlay()
	if err != nil {
		return err
	}
	selectedPaths, err := projectOperations(source, selection)
	if err != nil {
		return err
	}
	document, err := assembleDocument(source, selection, selectedPaths)
	if err != nil {
		return err
	}
	return writeDeterministicDocument(document)
}

func readPinnedSource() (map[string]any, error) {
	upstreamBytes, err := os.ReadFile(upstreamPath)
	if err != nil {
		return nil, fmt.Errorf("read pinned SnapTrade specification: %w", err)
	}
	sum := sha256.Sum256(upstreamBytes)
	if hex.EncodeToString(sum[:]) != wantSHA256 {
		return nil, errors.New("pinned SnapTrade specification checksum mismatch")
	}
	var source map[string]any
	if err := yaml.Unmarshal(upstreamBytes, &source); err != nil {
		return nil, fmt.Errorf("parse pinned SnapTrade specification: %w", err)
	}
	return source, nil
}

func readOverlay() (overlay, error) {
	overlayBytes, err := os.ReadFile(overlayPath)
	if err != nil {
		return overlay{}, fmt.Errorf("read OAuth bearer overlay: %w", err)
	}
	var selection overlay
	if err := yaml.Unmarshal(overlayBytes, &selection); err != nil {
		return overlay{}, fmt.Errorf("parse OAuth bearer overlay: %w", err)
	}
	if err := validateOverlay(selection); err != nil {
		return overlay{}, err
	}
	return selection, nil
}

func projectOperations(source map[string]any, selection overlay) (map[string]any, error) {
	paths, err := object(source, "paths")
	if err != nil {
		return nil, err
	}
	selectedPaths := make(map[string]any, len(selection.Operations))
	for _, chosen := range selection.Operations {
		pathItem, ok := paths[chosen.Path].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("pinned specification lacks selected path %q", chosen.Path)
		}
		operation, ok := pathItem[chosen.Method].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("pinned specification lacks selected operation %s %s", chosen.Method, chosen.Path)
		}
		operation["parameters"], err = filterParameters(operation["parameters"], chosen.RemoveParameters)
		if err != nil {
			return nil, fmt.Errorf("filter %s %s parameters: %w", chosen.Method, chosen.Path, err)
		}
		operation["security"] = []any{map[string]any{selection.OAuth.Name: []any{}}}
		selectedPaths[chosen.Path] = map[string]any{chosen.Method: operation}
	}
	return selectedPaths, nil
}

func assembleDocument(source map[string]any, selection overlay, selectedPaths map[string]any) (map[string]any, error) {
	components, err := object(source, "components")
	if err != nil {
		return nil, err
	}
	allSchemas, ok := components["schemas"].(map[string]any)
	if !ok {
		return nil, errors.New("pinned specification lacks component schemas")
	}
	allResponses, ok := components["responses"].(map[string]any)
	if !ok {
		return nil, errors.New("pinned specification lacks component responses")
	}
	selectedResponses, err := referencedResponses(selectedPaths, allResponses)
	if err != nil {
		return nil, err
	}
	allHeaders, ok := components["headers"].(map[string]any)
	if !ok {
		return nil, errors.New("pinned specification lacks component headers")
	}
	selectedHeaders, err := referencedComponents(selectedResponses, "#/components/headers/", allHeaders)
	if err != nil {
		return nil, err
	}
	selectedSchemas, err := transitiveSchemas([]any{selectedPaths, selectedResponses}, allSchemas)
	if err != nil {
		return nil, err
	}

	info, err := object(source, "info")
	if err != nil {
		return nil, err
	}
	info["x-upstream-revision"] = revision
	info["x-upstream-sha256"] = wantSHA256
	document := map[string]any{
		"openapi":  source["openapi"],
		"info":     info,
		"servers":  source["servers"],
		"security": []any{map[string]any{selection.OAuth.Name: []any{}}},
		"paths":    selectedPaths,
		"components": map[string]any{
			"securitySchemes": map[string]any{selection.OAuth.Name: map[string]any{"type": selection.OAuth.Type, "scheme": selection.OAuth.Scheme}},
			"schemas":         selectedSchemas,
			"responses":       selectedResponses,
			"headers":         selectedHeaders,
		},
	}
	return document, nil
}

func referencedResponses(root any, all map[string]any) (map[string]any, error) {
	return referencedComponents(root, "#/components/responses/", all)
}

func referencedComponents(root any, prefix string, all map[string]any) (map[string]any, error) {
	wanted := make(map[string]bool)
	if err := collectComponentRefs(root, prefix, wanted); err != nil {
		return nil, err
	}
	selected := make(map[string]any, len(wanted))
	for name := range wanted {
		component, exists := all[name]
		if !exists {
			return nil, fmt.Errorf("selected operation references missing component %q", name)
		}
		selected[name] = component
	}
	return selected, nil
}

func writeDeterministicDocument(document map[string]any) error {
	materialized, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("encode materialized provider contract: %w", err)
	}
	materialized = append(materialized, '\n')
	current, readErr := os.ReadFile(outputPath)
	if readErr == nil && string(current) == string(materialized) {
		return nil
	}
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("read materialized provider contract: %w", readErr)
	}
	if err := os.WriteFile(outputPath, materialized, 0o644); err != nil {
		return fmt.Errorf("write materialized provider contract: %w", err)
	}
	return nil
}

func validateOverlay(value overlay) error {
	if value.Upstream.Revision != revision || value.Upstream.SHA256 != wantSHA256 {
		return errors.New("OAuth bearer overlay does not identify the pinned upstream specification")
	}
	if len(value.Operations) != 5 || value.OAuth.Name == "" || value.OAuth.Type != "http" || value.OAuth.Scheme != "bearer" {
		return errors.New("OAuth bearer overlay must select exactly five operations and HTTP bearer security")
	}
	wanted := map[string]bool{
		"get /authorizations":                            false,
		"get /authorizations/{authorizationId}/accounts": false,
		"get /accounts/{accountId}/balances":             false,
		"get /accounts/{accountId}/positions/all":        false,
		"get /accounts/{accountId}/activities":           false,
	}
	for _, operation := range value.Operations {
		key := strings.ToLower(operation.Method) + " " + operation.Path
		seen, approved := wanted[key]
		if !approved || seen {
			return errors.New("OAuth bearer overlay must select only approved read-only portfolio operations")
		}
		wanted[key] = true
		removals := map[string]bool{"query:userId": false, "query:userSecret": false}
		for _, parameter := range operation.RemoveParameters {
			removalKey := parameter.In + ":" + parameter.Name
			present, approved := removals[removalKey]
			if !approved || present {
				return errors.New("OAuth bearer overlay must remove each Commercial user query parameter exactly once")
			}
			removals[removalKey] = true
		}
		for _, present := range removals {
			if !present {
				return errors.New("OAuth bearer overlay must remove both Commercial user query parameters")
			}
		}
	}
	return nil
}

func filterParameters(raw any, remove []parameterSelector) ([]any, error) {
	parameters, ok := raw.([]any)
	if !ok {
		return nil, errors.New("operation parameters are not an array")
	}
	removed := make(map[string]bool, len(remove))
	kept := make([]any, 0, len(parameters))
	for _, rawParameter := range parameters {
		parameter, ok := rawParameter.(map[string]any)
		if !ok {
			return nil, errors.New("operation parameter is not an object")
		}
		matched := false
		for _, selector := range remove {
			if parameter["in"] == selector.In && parameter["name"] == selector.Name {
				removed[selector.In+":"+selector.Name] = true
				matched = true
				break
			}
		}
		if !matched {
			kept = append(kept, parameter)
		}
	}
	for _, selector := range remove {
		if !removed[selector.In+":"+selector.Name] {
			return nil, fmt.Errorf("selected removal %s:%s was not present upstream", selector.In, selector.Name)
		}
	}
	return kept, nil
}

func transitiveSchemas(root any, all map[string]any) (map[string]any, error) {
	wanted := make(map[string]bool)
	if err := collectSchemaRefs(root, wanted); err != nil {
		return nil, err
	}
	selected := make(map[string]any)
	for {
		names := make([]string, 0, len(wanted))
		for name := range wanted {
			names = append(names, name)
		}
		sort.Strings(names)
		changed := false
		for _, name := range names {
			if _, exists := selected[name]; exists {
				continue
			}
			schema, exists := all[name]
			if !exists {
				return nil, fmt.Errorf("selected operation references missing schema %q", name)
			}
			selected[name] = schema
			before := len(wanted)
			if err := collectSchemaRefs(schema, wanted); err != nil {
				return nil, err
			}
			changed = changed || len(wanted) > before
		}
		if !changed && len(selected) == len(wanted) {
			return selected, nil
		}
	}
}

func collectSchemaRefs(value any, destination map[string]bool) error {
	return collectComponentRefs(value, "#/components/schemas/", destination)
}

func collectComponentRefs(value any, prefix string, destination map[string]bool) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if key == "$ref" {
				reference, ok := nested.(string)
				if !ok {
					return fmt.Errorf("selected contract contains unsupported reference %v", nested)
				}
				if strings.HasPrefix(reference, prefix) {
					destination[strings.TrimPrefix(reference, prefix)] = true
				}
				continue
			}
			if err := collectComponentRefs(nested, prefix, destination); err != nil {
				return err
			}
		}
	case []any:
		for _, nested := range typed {
			if err := collectComponentRefs(nested, prefix, destination); err != nil {
				return err
			}
		}
	}
	return nil
}

func object(parent map[string]any, key string) (map[string]any, error) {
	value, ok := parent[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("pinned specification lacks %s object", key)
	}
	return value, nil
}
