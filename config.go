package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfiguration(file string) ([]ResolverMapping, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var rawResolvers []json.RawMessage // Use RawMessage to defer parsing
	if err := json.Unmarshal(data, &rawResolvers); err != nil {
		return nil, err
	}

	var resolvers []ResolverMapping
	for _, raw := range rawResolvers {
		var basicMapping struct {
			Domain   string          `json:"domain"`
			Network  string          `json:"network"`
			Resolver json.RawMessage `json:"resolver"`
		}
		if err := json.Unmarshal(raw, &basicMapping); err != nil {
			return nil, err
		}
		resolver, err := unmarshalResolver(basicMapping.Resolver)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, ResolverMapping{
			Resolver: resolver,
			Domain:   basicMapping.Domain,
			Network:  basicMapping.Network,
		})
	}

	return resolvers, nil
}

func unmarshalResolver(data []byte) (Resolver, error) {
	var resolverType struct {
		Type string `json:"type"`
	}

	// First, unmarshal to get the type
	if err := json.Unmarshal(data, &resolverType); err != nil {
		return nil, err
	}

	var resolver Resolver
	switch resolverType.Type {
	case "basic":
		var basic BasicResolver
		if err := json.Unmarshal(data, &basic); err != nil {
			return nil, err
		}
		resolver = basic
	case "dot":
		var basic BasicResolver
		if err := json.Unmarshal(data, &basic); err != nil {
			return nil, err
		}
		resolver = MakeDoTResolver(basic.Server)
	case "static":
		var static StaticResolver
		if err := json.Unmarshal(data, &static); err != nil {
			return nil, err
		}
		resolver = static
	case "suffix":
		var suffix SuffixResolver
		if err := json.Unmarshal(data, &suffix); err != nil {
			return nil, err
		}
		resolver = suffix
	case "merge":
		var merge struct {
			Resolvers []json.RawMessage `json:"resolvers"`
		}
		if err := json.Unmarshal(data, &merge); err != nil {
			return nil, err
		}
		var mergedResolvers []Resolver
		for _, r := range merge.Resolvers {
			mergedResolver, err := unmarshalResolver(r)
			if err != nil {
				return nil, err
			}
			mergedResolvers = append(mergedResolvers, mergedResolver)
		}
		resolver = MergeResolver{Resolvers: mergedResolvers}
	default:
		return nil, fmt.Errorf("unknown resolver type: %s", resolverType.Type)
	}

	return resolver, nil
}
