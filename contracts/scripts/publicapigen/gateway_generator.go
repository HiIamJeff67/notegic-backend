package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
)

func writeGatewayArtifacts(root string, endpoints []endpoint) {
	base := filepath.Join(root, "contracts", "api-gateway", "v1", "public")
	components := map[string]any{
		"schemas": map[string]any{
			"Exception": map[string]any{
				"type":     "object",
				"required": []string{"reason", "domain", "operation", "message", "retryable"},
				"properties": map[string]any{
					"reason": map[string]any{"type": "string"}, "domain": map[string]any{"type": "string"},
					"operation": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"},
					"retryable": map[string]any{"type": "boolean"},
				},
			},
			"ErrorResponse": map[string]any{
				"type": "object", "required": []string{"success", "data", "exception"},
				"properties": map[string]any{
					"success":   map[string]any{"type": "boolean", "const": false},
					"data":      map[string]any{"type": "null"},
					"exception": map[string]any{"$ref": "#/components/schemas/Exception"},
				},
			},
		},
		"securitySchemes": map[string]any{
			"apiKey": map[string]any{"type": "apiKey", "in": "header", "name": "X-API-Key", "description": "User-owned API key. The secret is shown only once when created."},
		},
	}
	schemas := components["schemas"].(map[string]any)
	paths := map[string]any{}
	tags := map[string]bool{}
	for _, route := range endpoints {
		tags[route.Tag] = true
		pathItem, _ := paths[route.Path].(map[string]any)
		if pathItem == nil {
			pathItem = map[string]any{}
			paths[route.Path] = pathItem
		}
		operation := map[string]any{
			"operationId":       route.OperationID,
			"summary":           words(route.OperationID),
			"tags":              []string{route.Tag},
			"x-go-request-dto":  route.RequestType,
			"x-go-response-dto": route.ResponseType,
		}
		if route.Tag == "graphql" {
			operation["description"] = "The complete GraphQL SDL is published at public/graphql/schema.graphql; executable search operations are under examples/graphql."
		}
		header, body, param, query := requestParts(route.RequestType)
		query = routeQuery(route.Path, param, query)
		parameters := []any{}
		for _, name := range pathNames(route.Path) {
			schema := map[string]any{"type": "string", "format": "uuid"}
			if route.Tag == "static" && name == "id" {
				schema = map[string]any{"type": "integer", "minimum": 1}
			}
			for fieldName, property := range properties(param) {
				if matchesPathParameter(fieldName, name) {
					if propertySchema, ok := property.(map[string]any); ok {
						schema = propertySchema
					}
					break
				}
			}
			parameters = append(parameters, map[string]any{"name": name, "in": "path", "required": true, "schema": schema, "example": schemaExample(schema, name)})
		}
		for name, rawSchema := range properties(query) {
			schema := rawSchema.(map[string]any)
			parameters = append(parameters, map[string]any{"name": name, "in": "query", "required": isRequired(query, name), "schema": schema, "example": schemaExample(schema, name)})
		}
		if _, ok := properties(header)["userAgent"]; ok {
			parameters = append(parameters, map[string]any{"name": "User-Agent", "in": "header", "required": true, "schema": map[string]any{"type": "string"}, "example": "NotegicIntegration/1.0"})
		}
		if len(parameters) > 0 {
			sort.Slice(parameters, func(i, j int) bool {
				return parameters[i].(map[string]any)["name"].(string) < parameters[j].(map[string]any)["name"].(string)
			})
			operation["parameters"] = parameters
		}
		if route.Tag == "graphql" && route.Method == "POST" {
			body = map[string]any{"type": "object", "required": []string{"query"}, "properties": map[string]any{
				"query": map[string]any{"type": "string"}, "operationName": map[string]any{"type": "string"}, "variables": map[string]any{"type": "object"},
			}}
		}
		if len(properties(body)) > 0 || (route.Tag == "graphql" && route.Method == "POST") {
			requestSchemaName := upperFirst(route.OperationID) + "RequestBody"
			schemas[requestSchemaName] = body
			operation["requestBody"] = map[string]any{
				"required": true,
				"content": map[string]any{"application/json": map[string]any{
					"schema":  map[string]any{"$ref": "#/components/schemas/" + requestSchemaName},
					"example": operationBodyExample(route.OperationID, body, false),
				}},
			}
		}
		if route.Tag == "static" {
			operation["responses"] = map[string]any{
				"200": map[string]any{"description": "PNG avatar; unknown IDs fall back to the default avatar.", "content": map[string]any{"image/png": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}},
				"429": map[string]any{"description": "Rate limit exceeded", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}},
			}
			operation["security"] = []any{}
			pathItem[strings.ToLower(route.Method)] = operation
			continue
		}
		responseData := map[string]any{}
		if definition, ok := findType(route.ResponseType); ok {
			responseData = schemaFor(definition.expr, false, map[string]bool{route.ResponseType: true})
		}
		removePrivateTokenFields(responseData)
		responseName := upperFirst(route.OperationID) + "ResponseData"
		schemas[responseName] = responseData
		successName := upperFirst(route.OperationID) + "SuccessResponse"
		schemas[successName] = map[string]any{
			"type": "object", "required": []string{"success", "data", "exception"},
			"properties": map[string]any{
				"success":           map[string]any{"type": "boolean", "const": true},
				"data":              map[string]any{"$ref": "#/components/schemas/" + responseName},
				"exception":         map[string]any{"type": "null"},
				"embedded":          map[string]any{"type": "object", "properties": map[string]any{"publicId": map[string]any{"type": "string", "format": "uuid"}}},
				"refreshableTokens": map[string]any{"type": "object", "properties": map[string]any{"newCSRFToken": map[string]any{"type": "string"}}},
			},
		}
		operation["responses"] = gatewayResponses(successName, route.OperationID)
		if authenticated(route.OperationID) {
			operation["security"] = []any{map[string]any{"apiKey": []any{}}}
		} else {
			operation["security"] = []any{}
		}
		pathItem[strings.ToLower(route.Method)] = operation
	}
	tagList := []any{}
	for tag := range tags {
		tagList = append(tagList, map[string]any{"name": tag})
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i].(map[string]any)["name"].(string) < tagList[j].(map[string]any)["name"].(string)
	})
	openAPI := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title": "Notegic APIGateway API", "version": cgateway.DevelopmentCompleteVersion,
			"description": "Complete machine-readable contract for routes exposed by APIGateway v1. Authenticated requests use user-owned API keys.",
		},
		"servers": []any{
			map[string]any{"url": "http://localhost/api/development/v1", "description": "Local development"},
			map[string]any{"url": "https://api.notegic.app/api/development/v1", "description": "Hosted Beta"},
		},
		"tags": tagList, "paths": paths, "components": components,
	}
	writeJSON(filepath.Join(base, "openapi", "openapi.json"), openAPI)
	writeJSON(filepath.Join(base, "postman", "notegic-api-gateway-v1.postman_collection.json"), gatewayPostman(endpoints))
	writeJSON(filepath.Join(base, "postman", "notegic-api-gateway-v1.postman_environment.example.json"), postmanEnvironment())
	writeText(filepath.Join(base, "examples", "curl", "all-endpoints.sh"), curlExamples(endpoints))
	writeText(filepath.Join(base, "examples", "http", "all-endpoints.http"), httpExamples(endpoints))
	writeText(filepath.Join(base, "reference", "endpoints.md"), endpointReference(endpoints))
	writeGatewayRules(base, len(endpoints))
}

func removePrivateTokenFields(schema map[string]any) {
	props := properties(schema)
	delete(props, "accessToken")
	delete(props, "refreshToken")
	required, ok := schema["required"].([]string)
	if !ok {
		return
	}
	filtered := required[:0]
	for _, name := range required {
		if name != "accessToken" && name != "refreshToken" {
			filtered = append(filtered, name)
		}
	}
	schema["required"] = filtered
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func pathNames(path string) []string {
	pattern := regexp.MustCompile(`\{([^}]+)\}`)
	result := []string{}
	for _, match := range pattern.FindAllStringSubmatch(path, -1) {
		result = append(result, match[1])
	}
	return result
}

func isRequired(schema map[string]any, name string) bool {
	for _, raw := range toAnySlice(schema["required"]) {
		if raw == name {
			return true
		}
	}
	return false
}

func toAnySlice(value any) []any {
	switch values := value.(type) {
	case []any:
		return values
	case []string:
		result := make([]any, len(values))
		for i, item := range values {
			result[i] = item
		}
		return result
	default:
		return nil
	}
}

func gatewayResponses(successName, operationID string) map[string]any {
	response := func(description, schema string) map[string]any {
		return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": schema}}}}
	}
	successStatus := "200"
	if strings.HasPrefix(operationID, "create") {
		successStatus = "201"
	}
	return map[string]any{
		successStatus: response("Successful operation", "#/components/schemas/"+successName),
		"400":         response("Invalid request", "#/components/schemas/ErrorResponse"),
		"401":         response("Authentication or CSRF failed", "#/components/schemas/ErrorResponse"),
		"403":         response("Permission denied", "#/components/schemas/ErrorResponse"),
		"404":         response("Resource not found", "#/components/schemas/ErrorResponse"),
		"409":         response("State conflict", "#/components/schemas/ErrorResponse"),
		"429":         response("Rate limit exceeded", "#/components/schemas/ErrorResponse"),
		"500":         response("Unexpected server error", "#/components/schemas/ErrorResponse"),
		"503":         response("Service unavailable", "#/components/schemas/ErrorResponse"),
	}
}

func operationBodyExample(operationID string, schema map[string]any, variables bool) any {
	if variables {
		switch operationID {
		case "login":
			return map[string]any{"account": "{{account}}", "password": "{{password}}"}
		case "register":
			return map[string]any{"name": "{{account}}", "email": "{{email}}", "password": "{{password}}"}
		}
	}
	return schemaExample(schema, "body")
}

func queryPairs(schema map[string]any) [][2]string {
	names := make([]string, 0, len(properties(schema)))
	for name := range properties(schema) {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([][2]string, 0, len(names))
	for _, name := range names {
		property := properties(schema)[name].(map[string]any)
		example := schemaExample(property, name)
		if values, ok := example.([]any); ok {
			if len(values) == 0 {
				continue
			}
			example = values[0]
		}
		var value string
		switch typed := example.(type) {
		case string:
			value = typed
		default:
			encoded, _ := json.Marshal(typed)
			value = string(encoded)
		}
		result = append(result, [2]string{name, value})
	}
	return result
}

func appendQuery(raw string, query map[string]any) string {
	pairs := queryPairs(query)
	for index, pair := range pairs {
		separator := "&"
		if index == 0 {
			separator = "?"
		}
		raw += separator + url.QueryEscape(pair[0]) + "=" + url.QueryEscape(pair[1])
	}
	return raw
}

func postmanURL(base, path string, query map[string]any) map[string]any {
	for _, name := range pathNames(path) {
		path = strings.ReplaceAll(path, "{"+name+"}", "{{"+parameterVariableName(name)+"}}")
	}
	raw := appendQuery("{{"+base+"}}"+path, query)
	queryItems := []any{}
	for _, pair := range queryPairs(query) {
		queryItems = append(queryItems, map[string]any{"key": pair[0], "value": pair[1]})
	}
	result := map[string]any{"raw": raw, "host": []string{"{{" + base + "}}"}}
	if len(queryItems) > 0 {
		result["query"] = queryItems
	}
	return result
}

func gatewayPostman(endpoints []endpoint) map[string]any {
	byTag := map[string][]any{}
	for _, route := range endpoints {
		_, body, param, query := requestParts(route.RequestType)
		query = routeQuery(route.Path, param, query)
		headers := []any{map[string]any{"key": "User-Agent", "value": "{{userAgent}}", "type": "text"}}
		if authenticated(route.OperationID) {
			headers = append(headers, map[string]any{"key": "X-API-Key", "value": "{{apiKey}}", "type": "text"})
		}
		if route.Method != "GET" {
			headers = append(headers, map[string]any{"key": "Content-Type", "value": "application/json", "type": "text"})
		}
		request := map[string]any{
			"method": route.Method, "header": headers,
			"url":         postmanURL("apiGatewayBaseUrl", route.Path, query),
			"description": fmt.Sprintf("%s. Go DTO: `%s`; response DTO: `%s`.", words(route.OperationID), route.RequestType, route.ResponseType),
		}
		if len(properties(body)) > 0 || (route.Tag == "graphql" && route.Method == "POST") {
			example := operationBodyExample(route.OperationID, body, true)
			if route.Tag == "graphql" {
				example = map[string]any{"query": "query ContractCheck { __typename }", "variables": map[string]any{}}
			}
			raw, _ := json.MarshalIndent(example, "", "  ")
			request["body"] = map[string]any{"mode": "raw", "raw": string(raw), "options": map[string]any{"raw": map[string]any{"language": "json"}}}
		}
		tests := []string{
			"pm.test('HTTP response is below 500', function () { pm.expect(pm.response.code).to.be.below(500); });",
		}
		name := kebabCase(route.OperationID)
		if route.Method == "DELETE" || strings.Contains(strings.ToLower(route.OperationID), "reset") {
			name = "[DESTRUCTIVE] " + name
		}
		byTag[route.Tag] = append(byTag[route.Tag], map[string]any{
			"name":    name,
			"request": request,
			"event":   []any{map[string]any{"listen": "test", "script": map[string]any{"type": "text/javascript", "exec": tests}}},
		})
	}
	folders := []any{}
	for tag, items := range byTag {
		folders = append(folders, map[string]any{"name": tag, "item": items})
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].(map[string]any)["name"].(string) < folders[j].(map[string]any)["name"].(string)
	})
	return map[string]any{
		"info": map[string]any{
			"_postman_id": "a8f3fa62-f2cd-4cf0-97ab-a801b42ee101", "name": "Notegic APIGateway v1",
			"description": "Generated from the APIGateway v1 route and Go DTO contracts. Authenticated requests use X-API-Key.",
			"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item": folders,
	}
}

func postmanEnvironment() map[string]any {
	values := []any{
		map[string]any{"key": "apiGatewayBaseUrl", "value": "http://localhost/api/development/v1", "enabled": true},
		map[string]any{"key": "userAgent", "value": "Postman/Notegic-v1", "enabled": true},
		map[string]any{"key": "apiKey", "value": "", "enabled": true, "type": "secret"},
	}
	for _, name := range []string{"id", "userPublicId", "stationId", "routineId", "routineTagId", "routineTaskId", "rootShelfId", "subShelfId", "prevSubShelfId", "parentSubShelfId", "materialId", "blockPackId", "blockId", "itemId"} {
		value := "00000000-0000-4000-8000-000000000001"
		if name == "id" {
			value = "1"
		}
		values = append(values, map[string]any{"key": name, "value": value, "enabled": true})
	}
	return map[string]any{
		"id": "8b8bf8a7-bbc0-4af7-9a6a-068180a25fa0", "name": "Notegic v1 example (no credentials)",
		"values": values, "_postman_variable_scope": "environment", "_postman_exported_using": "Notegic generator",
	}
}

func shellPath(path string) string {
	for _, name := range pathNames(path) {
		path = strings.ReplaceAll(path, "{"+name+"}", "${"+parameterVariableName(name)+"}")
	}
	return path
}

func curlExamples(endpoints []endpoint) string {
	var output bytes.Buffer
	output.WriteString("#!/usr/bin/env bash\nset -euo pipefail\n\n")
	output.WriteString("api_gateway_base_url=\"${API_GATEWAY_BASE_URL:-http://localhost/api/development/v1}\"\napi_key=\"${API_KEY:-}\"\n")
	output.WriteString("user_agent=\"${USER_AGENT:-NotegicCurlExample/1.0}\"\n")
	output.WriteString("id=\"${AVATAR_ID:-1}\"\n")
	for _, id := range []string{"userPublicId", "stationId", "routineId", "routineTagId", "routineTaskId", "rootShelfId", "subShelfId", "prevSubShelfId", "parentSubShelfId", "materialId", "blockPackId", "blockId", "itemId"} {
		fmt.Fprintf(&output, "%s=\"${%s:-00000000-0000-4000-8000-000000000001}\"\n", id, strings.ToUpper(id))
	}
	output.WriteString("\n# Run individual functions deliberately. DELETE/reset functions are not invoked automatically.\n")
	for _, route := range endpoints {
		_, body, param, query := requestParts(route.RequestType)
		query = routeQuery(route.Path, param, query)
		fmt.Fprintf(&output, "\n%s() {\n", route.OperationID)
		fmt.Fprintf(&output, "  curl --fail-with-body --silent --show-error -X %s \\\n    -H \"User-Agent: $user_agent\"", route.Method)
		if route.Method != "GET" {
			output.WriteString(" \\\n    -H \"Content-Type: application/json\"")
		}
		if authenticated(route.OperationID) {
			output.WriteString(" \\\n    -H \"X-API-Key: $api_key\"")
		}
		if len(properties(body)) > 0 || (route.Tag == "graphql" && route.Method == "POST") {
			if route.Tag == "graphql" {
				output.WriteString(" \\\n    --data '{\"query\":\"query ContractCheck { __typename }\",\"variables\":{}}'")
			} else {
				example, _ := json.Marshal(operationBodyExample(route.OperationID, body, false))
				fmt.Fprintf(&output, " \\\n    --data '%s'", strings.ReplaceAll(string(example), "'", "'\\''"))
			}
		}
		fmt.Fprintf(&output, " \\\n    \"$api_gateway_base_url%s\"\n}\n", appendQuery(shellPath(route.Path), query))
	}
	return output.String()
}

func endpointReference(endpoints []endpoint) string {
	var output bytes.Buffer
	output.WriteString("# APIGateway v1 endpoint reference\n\n")
	output.WriteString("This catalog is generated from the APIGateway public route allowlist. Request and response property definitions live in `../openapi/openapi.json`.\n\n")
	output.WriteString("| Method | Path | Operation | Request DTO | Response DTO |\n| --- | --- | --- | --- | --- |\n")
	for _, route := range endpoints {
		fmt.Fprintf(&output, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n", route.Method, route.Path, route.OperationID, route.RequestType, route.ResponseType)
	}
	return output.String()
}

func httpExamples(endpoints []endpoint) string {
	var output bytes.Buffer
	output.WriteString("@apiGatewayBaseUrl = http://localhost/api/development/v1\n@apiKey = replace-with-your-api-key\n@userAgent = NotegicHttpFile/1.0\n")
	output.WriteString("@id = 1\n")
	for _, id := range []string{"userPublicId", "stationId", "routineId", "routineTagId", "routineTaskId", "rootShelfId", "subShelfId", "prevSubShelfId", "parentSubShelfId", "materialId", "blockPackId", "blockId", "itemId"} {
		fmt.Fprintf(&output, "@%s = 00000000-0000-4000-8000-000000000001\n", id)
	}
	for _, route := range endpoints {
		_, body, param, query := requestParts(route.RequestType)
		query = routeQuery(route.Path, param, query)
		path := route.Path
		for _, name := range pathNames(path) {
			path = strings.ReplaceAll(path, "{"+name+"}", "{{"+parameterVariableName(name)+"}}")
		}
		fmt.Fprintf(&output, "\n### %s %s\n%s {{apiGatewayBaseUrl}}%s\nUser-Agent: {{userAgent}}\n", route.Method, words(route.OperationID), route.Method, appendQuery(path, query))
		if route.Method != "GET" {
			output.WriteString("Content-Type: application/json\n")
		}
		if authenticated(route.OperationID) {
			output.WriteString("X-API-Key: {{apiKey}}\n")
		}
		if len(properties(body)) > 0 || (route.Tag == "graphql" && route.Method == "POST") {
			example := operationBodyExample(route.OperationID, body, false)
			if route.Tag == "graphql" {
				example = map[string]any{"query": "query ContractCheck { __typename }", "variables": map[string]any{}}
			}
			raw, _ := json.MarshalIndent(example, "", "  ")
			fmt.Fprintf(&output, "\n%s\n", raw)
		}
	}
	return output.String()
}

func writeGatewayRules(base string, endpointCount int) {
	writeText(filepath.Join(base, "README.md"), fmt.Sprintf(`# Notegic APIGateway v1 public API

This directory contains the machine-readable and human-readable contract for all %d versioned routes currently exposed by APIGateway v1.

The published domains are RootShelf, SubShelf, Material, BlockPack, Block, Station, Routine, RoutineTask, RoutineTaskDependency, and RoutineTag. RoutineRecord is available through ClientGateway GraphQL search only. Client-only auth, user/account, notification, realtime, GraphQL, and static routes are intentionally excluded.

- **Canonical contract:** `+"`openapi/openapi.json`"+` (OpenAPI 3.1)
- **Rules:** `+"`rules/`"+`
- **Endpoint catalog:** `+"`reference/endpoints.md`"+`
- **Runnable examples:** `+"`examples/curl/all-endpoints.sh`"+` and `+"`examples/http/all-endpoints.http`"+`
- **API key example:** send `+"`X-API-Key`"+` using the value in your private environment.
- **Postman:** import both JSON files in `+"`postman/`"+`
- **Version records:** `+"`versions/dev-log.md`"+` and `+"`versions/comparison.md`"+`

The generated artifacts are refreshed from routes and Go DTOs with:

`+"```bash"+`
make -C contracts public-api-gen
`+"```"+`

No real account, password, cookie, CSRF token, or API secret belongs in this directory.`, endpointCount))
	writeText(filepath.Join(base, "rules", "authentication.md"), `# Authentication and credential rules

APIGateway v1 integrations authenticate with a user-owned `+"`X-API-Key`"+` header. Create the key through the authenticated ClientGateway flow, record the returned secret once, and send it on subsequent API requests.

- The full secret is returned only once. Store it in a secret manager or local environment and never commit it.
- The server persists only a SHA-256 digest and a short display prefix.
- Keys can be expired or revoked; revoked keys fail immediately even if a cache entry exists.
- Do not log request bodies containing credentials, `+"`X-API-Key`"+`, Cookie, Set-Cookie, or CSRF values.
- Unauthorized rate limits are primarily keyed by client IP; API key ID is auxiliary only.`)
	writeText(filepath.Join(base, "rules", "http-contract.md"), `# HTTP contract rules

- Base path: `+"`/api/development/v1`"+` for the current Beta namespace.
- Request and response media type: `+"`application/json`"+`.
- Path resource identifiers are UUID strings unless the operation schema says otherwise.
- Times use RFC 3339 date-time strings.
- Public success envelope: `+"`{ \"success\": true, \"data\": ..., \"exception\": null }`"+`.
- Public failure envelope: `+"`{ \"success\": false, \"data\": null, \"exception\": ... }`"+`.
- `+"`exception.retryable`"+` is the server's explicit retry signal. A client must not infer retryability only from the message.
- Optional `+"`embedded.publicId`"+` identifies the authenticated actor.
- Optional `+"`refreshableTokens.newCSRFToken`"+` replaces the previously stored CSRF value.
- Unknown request fields should not be used for forward compatibility. Only documented properties form the contract.
- Batch requests are not atomic unless the operation description or future version explicitly promises atomicity.
- DELETE may be soft delete or permanent delete; permanent endpoints include `+"`permanently`"+` in their path.

The OpenAPI operation extensions `+"`x-go-request-dto`"+` and `+"`x-go-response-dto`"+` identify the source contracts used to generate each schema.`)
	writeText(filepath.Join(base, "rules", "rate-limits-and-retries.md"), `# Rate limits and retry rules

APIGateway emits these headers on rate-limited routes:

- `+"`X-RateLimit-Limit`"+`: request allowance for the active window.
- `+"`X-RateLimit-Remaining`"+`: remaining allowance.
- `+"`X-RateLimit-Reset`"+`: Unix timestamp for the next reset estimate.
- `+"`X-RateLimit-Window`"+`: configured window duration.
- `+"`X-RateLimit-Policy`"+`: currently `+"`hybrid-token-bucket`"+`.

Current APIGateway v1 routes use an IP/fingerprint limit of 1,000 requests per minute with a 100 requests/second token bucket and burst 10. An authenticated-user limiter exists internally but is not a documented allowance for these routes. These are service limits, not permanent entitlements, and may be lowered during Beta.

On HTTP 429, wait until the reset time and add randomized backoff. Retry only idempotent reads or writes carrying an application-level idempotency guarantee. The current public mutations do not generally expose an idempotency key, so a client must reconcile state before retrying a timed-out mutation.`)
	writeText(filepath.Join(base, "rules", "origins-and-security.md"), `# Origin and security rules

- Send the API key in `+"`X-API-Key`"+`; never put it in a URL, query string, or request body.
- Cross-origin browser access works only for origins configured in the APIGateway allowlist.
- CLI and server clients should omit Origin and Referer rather than forge an allowed browser origin.
- HTTPS is mandatory outside local development.
- Never commit Postman environments after adding credentials.
- Never place account passwords, API keys, or authorization codes in URLs.
- Treat all IDs and permissions returned by the client as untrusted; server authorization remains authoritative.

Public documentation does not itself grant a third-party origin permission to call the service.`)
	writeText(filepath.Join(base, "rules", "versioning.md"), `# Version and change rules

The URL version (`+"`v1`"+`) is the compatibility boundary for request and response contracts. Repository release tags use Semantic Versioning independently.

- Backward-compatible fields and endpoints may be added within v1.
- Clients must ignore unknown response fields.
- Removing or changing the meaning/type of a field requires a new API version or a documented migration window.
- Deprecated operations remain in the OpenAPI document with a removal date before deletion.
- The development namespace indicates Beta stability; it does not remove the v1 compatibility obligation for documented behavior.
- OpenAPI, Postman, examples, rules, routes, and Go DTO changes must ship together.

Compare generated OpenAPI artifacts between releases to produce the public API change log.`)
	writeText(filepath.Join(base, "versions", "dev-log.md"), fmt.Sprintf(`# APIGateway v1 development log

## Current contract baseline

- Published surface: %d APIGateway operations across nine enabled resource domains.
- Contract format: OpenAPI 3.1.
- Authentication: user-owned `+"`X-API-Key`"+` header; key creation remains on ClientGateway.
- Tooling: Postman 2.1 collection/environment, curl functions, and an HTTP client file.
- Stability: Beta namespace with v1 compatibility rules.

Future entries must identify added, changed, deprecated, and removed operations and link to their migration notes. Never record a breaking change only in application release notes.`, endpointCount))
	writeText(filepath.Join(base, "versions", "comparison.md"), `# APIGateway API version comparison

Only APIGateway v1 is currently published, so there is no second public contract to compare yet.

| Capability | v1 Beta |
| --- | --- |
| HTTP contract | OpenAPI 3.1 |
| Public domains | RootShelf, SubShelf, Material, BlockPack, Block, Station, Routine, RoutineTask, RoutineTaskDependency, RoutineTag |
| Authentication | User-owned `+"`X-API-Key`"+` header |
| API keys | Required for every published operation |
| Importable client | Postman Collection 2.1 |

When another API version is introduced, this file must compare base paths, authentication, renamed or removed operations, schema changes, error behavior, rate limits, and migration deadlines.`)
}
