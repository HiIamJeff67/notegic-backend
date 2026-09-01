package main

import "testing"

func TestFilterPublicEndpointsUsesResourceAllowlist(t *testing.T) {
	endpoints := []endpoint{
		{Tag: "root-shelves"},
		{Tag: "sub-shelves"},
		{Tag: "materials"},
		{Tag: "block-packs"},
		{Tag: "blocks"},
		{Tag: "stations"},
		{Tag: "routines"},
		{Tag: "routine-tasks"},
		{Tag: "routine-tags"},
		{Tag: "auth"},
		{Tag: "users"},
		{Tag: "notifications"},
		{Tag: "realtime"},
		{Tag: "graphql"},
		{Tag: "static"},
	}

	filtered := filterPublicEndpoints(endpoints)
	if len(filtered) != 9 {
		t.Fatalf("filterPublicEndpoints() returned %d endpoints, want 9", len(filtered))
	}
	for _, endpoint := range filtered {
		if endpoint.Tag == "auth" || endpoint.Tag == "users" || endpoint.Tag == "notifications" || endpoint.Tag == "realtime" || endpoint.Tag == "graphql" || endpoint.Tag == "static" || endpoint.Tag == "routine-records" {
			t.Fatalf("filterPublicEndpoints() retained client-only tag %q", endpoint.Tag)
		}
	}
}
