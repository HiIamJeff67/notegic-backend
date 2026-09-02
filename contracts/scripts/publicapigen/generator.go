package main

import "fmt"

type ArtifactGenerator struct {
	root   string
	parser *ContractParser
}

func NewArtifactGenerator(root string, parser *ContractParser) *ArtifactGenerator {
	return &ArtifactGenerator{root: root, parser: parser}
}

func (g *ArtifactGenerator) Generate(endpoints []endpoint) {
	if g.parser == nil {
		panic("public API generator requires a contract parser")
	}
	publicEndpoints := filterPublicEndpoints(endpoints)
	writeGatewayArtifacts(g.root, publicEndpoints)
	writeRealtimeArtifacts(g.root)
	writeDecisionRecord(g.root, publicEndpoints)
	fmt.Printf("generated %d APIGateway operations and RealtimeGateway public HTTP/WebSocket contracts\n", len(publicEndpoints))
}

func filterPublicEndpoints(endpoints []endpoint) []endpoint {
	publicEndpoints := make([]endpoint, 0, len(endpoints))
	allowedTags := map[string]bool{
		"root-shelves":              true,
		"sub-shelves":               true,
		"materials":                 true,
		"block-packs":               true,
		"blocks":                    true,
		"stations":                  true,
		"routines":                  true,
		"routine-tasks":             true,
		"routine-task-dependencies": true,
		"routine-tags":              true,
	}
	for _, endpoint := range endpoints {
		// Only stable resource domains are public in the first API release.
		// Auth, user/account, notification, realtime, GraphQL, and static
		// surfaces remain ClientGateway-only until explicitly reviewed.
		if !allowedTags[endpoint.Tag] {
			continue
		}
		publicEndpoints = append(publicEndpoints, endpoint)
	}
	return publicEndpoints
}
