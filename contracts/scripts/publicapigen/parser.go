package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var (
	typesByName          = map[string][]typeDefinition{}
	enumValues           = map[string][]string{}
	bindRequests         = map[string]string{}
	responsesByRequest   = map[string]string{}
	pathParameterPattern = regexp.MustCompile(`:([A-Za-z][A-Za-z0-9-]*)`)
)

// ContractParser owns source discovery and AST-to-contract parsing.
type ContractParser struct{ root string }

func NewContractParser(root string) *ContractParser { return &ContractParser{root: root} }

func (p *ContractParser) Parse() []endpoint {
	resetParserState()
	loadContractTypes(p.root)
	loadBinderContracts(p.root)
	loadControllerContracts(p.root)
	return loadGatewayEndpoints(p.root)
}

func (p *ContractParser) FindType(name string) (typeDefinition, bool) { return findType(name) }

func resetParserState() {
	typesByName = map[string][]typeDefinition{}
	enumValues = map[string][]string{}
	bindRequests = map[string]string{}
	responsesByRequest = map[string]string{}
}

func loadContractTypes(root string) {
	contractRoots := []string{
		filepath.Join(root, "contracts", "core", "v1", "api"),
		filepath.Join(root, "contracts", "core", "v1", "types"),
		filepath.Join(root, "contracts", "notification", "v1", "api"),
		filepath.Join(root, "contracts", "notification", "v1", "types"),
		filepath.Join(root, "contracts", "types"),
	}
	for _, contractRoot := range contractRoots {
		must(filepath.WalkDir(contractRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return err
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if parseErr != nil {
				return parseErr
			}
			for _, declaration := range file.Decls {
				general, ok := declaration.(*ast.GenDecl)
				if !ok {
					continue
				}
				switch general.Tok {
				case token.TYPE:
					for _, specification := range general.Specs {
						typeSpec := specification.(*ast.TypeSpec)
						typesByName[typeSpec.Name.Name] = append(typesByName[typeSpec.Name.Name], typeDefinition{
							directory: filepath.Dir(path),
							expr:      typeSpec.Type,
						})
					}
				case token.CONST:
					for _, specification := range general.Specs {
						valueSpec := specification.(*ast.ValueSpec)
						typeName := expressionName(valueSpec.Type)
						if typeName == "" {
							continue
						}
						for _, value := range valueSpec.Values {
							if literal, ok := value.(*ast.BasicLit); ok && literal.Kind == token.STRING {
								decoded, decodeErr := strconv.Unquote(literal.Value)
								if decodeErr == nil {
									enumValues[typeName] = append(enumValues[typeName], decoded)
								}
							}
						}
					}
				}
			}
			return nil
		}))
	}
}

func expressionName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return value.Sel.Name
	default:
		return ""
	}
}

func loadBinderContracts(root string) {
	binderRoot := filepath.Join(root, "runtimes", "clientgateway", "transports", "api", "binders")
	pattern := regexp.MustCompile(`(?s)func \(b \*\w+Binder\) Bind(\w+)\s*\(\s*controllerFunc controllers\.Func\[\*(?:\w+\.)?(\w+RequestDto)\]`)
	entries, err := os.ReadDir(binderRoot)
	must(err)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(binderRoot, entry.Name()))
		must(readErr)
		for _, match := range pattern.FindAllStringSubmatch(string(content), -1) {
			bindRequests["Bind"+match[1]] = match[2]
		}
	}
}

func loadControllerContracts(root string) {
	controllerRoot := filepath.Join(root, "runtimes", "clientgateway", "transports", "api", "controllers")
	pattern := regexp.MustCompile(`(?s)Call(?:Securly)?\[\s*(?:\w+\.)?(\w+RequestDto)\s*,\s*(?:\w+\.)?(\w+ResponseDto)\s*,?\s*\]`)
	entries, err := os.ReadDir(controllerRoot)
	must(err)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(controllerRoot, entry.Name()))
		must(readErr)
		for _, match := range pattern.FindAllStringSubmatch(string(content), -1) {
			responsesByRequest[match[1]] = match[2]
		}
	}
}

type routeFileConfig struct {
	tag      string
	prefixes map[string]string
}

var routeConfigs = map[string]routeFileConfig{
	"auth_route.go":                {"auth", map[string]string{"authRoutes": "/auth"}},
	"user_route.go":                {"users", map[string]string{"userRoutes": "/users"}},
	"user_info_route.go":           {"user-info", map[string]string{"userInfoRoutes": "/me/info"}},
	"user_setting_route.go":        {"user-settings", map[string]string{"userSettingRoutes": "/me/settings"}},
	"user_account_route.go":        {"user-account", map[string]string{"userAccountRoutes": "/me/account"}},
	"station_route.go":             {"stations", map[string]string{"stationRoutes": "/stations", "visualizationRoutes": "/stations/visualizations"}},
	"routine_route.go":             {"routines", map[string]string{"routineRoutes": "/routines", "visualizationRoutes": "/routines/visualizations"}},
	"routine_tag_route.go":         {"routine-tags", map[string]string{"routineTagRoutes": "/routine-tags"}},
	"routine_task_route.go":        {"routine-tasks", map[string]string{"routineTaskRoutes": "/routine-tasks", "visualizationRoutes": "/routine-tasks/visualizations"}},
	"routine_task_record_route.go": {"routine-task-records", map[string]string{"routineTaskRecordRouterGroup": "/routine-task-records", "visualizationRoutes": "/routine-task-records/visualizations"}},
	"root_shelf_route.go":          {"root-shelves", map[string]string{"rootShelfRoutes": "/root-shelves"}},
	"sub_shelf_route.go":           {"sub-shelves", map[string]string{"subShelfRoutes": "/sub-shelves"}},
	"material_route.go":            {"materials", map[string]string{"materialRoutes": "/materials"}},
	"block_pack_route.go":          {"block-packs", map[string]string{"blockPackRoutes": "/block-packs"}},
	"block_route.go":               {"blocks", map[string]string{"blockRoutes": "/blocks"}},
	"notification_route.go":        {"notifications", map[string]string{"notificationRoutes": "/notifications"}},
	"realtime_route.go":            {"realtime", map[string]string{"connectionRouterGroup": "/realtime/connection", "channelRouterGroup": "/realtime/channel"}},
	"graphql_route.go":             {"graphql", map[string]string{"graphqlRoutes": "/graphql"}},
	"static_routes.go":             {"static", map[string]string{"globalImagesGroup": "/static/global-images"}},
}

func loadGatewayEndpoints(root string) []endpoint {
	routeRoot := filepath.Join(root, "runtimes", "clientgateway", "transports", "api", "routes", "developmentroutes")
	result := []endpoint{}
	for fileName, config := range routeConfigs {
		path := filepath.Join(routeRoot, fileName)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		must(err)
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !isHTTPMethod(selector.Sel.Name) {
				return true
			}
			router, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			prefix, ok := config.prefixes[router.Name]
			if !ok {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			suffix, _ := strconv.Unquote(literal.Value)
			bindName := ""
			ast.Inspect(call, func(inner ast.Node) bool {
				innerSelector, selectorOK := inner.(*ast.SelectorExpr)
				if selectorOK && strings.HasPrefix(innerSelector.Sel.Name, "Bind") {
					if owner, ownerOK := innerSelector.X.(*ast.Ident); ownerOK && strings.HasSuffix(owner.Name, "Binder") {
						bindName = innerSelector.Sel.Name
					}
				}
				return true
			})
			operationID := strings.TrimPrefix(bindName, "Bind")
			requestType := bindRequests[bindName]
			if operationID == "" {
				if fileName == "static_routes.go" {
					operationID = "GetGlobalAvatar"
				} else {
					operationID = "GraphQL" + strings.Title(strings.ToLower(selector.Sel.Name))
				}
			}
			result = append(result, endpoint{
				Method:       strings.ToUpper(selector.Sel.Name),
				Path:         normalizePath(prefix + suffix),
				Tag:          config.tag,
				OperationID:  lowerFirst(operationID),
				RequestType:  requestType,
				ResponseType: responsesByRequest[requestType],
			})
			return true
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Method < result[j].Method
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func isHTTPMethod(value string) bool {
	switch value {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func normalizePath(path string) string {
	path = strings.ReplaceAll(path, "//", "/")
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return pathParameterPattern.ReplaceAllStringFunc(path, func(match string) string {
		return "{" + kebabCase(strings.TrimPrefix(match, ":")) + "}"
	})
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func words(value string) string {
	var output []rune
	for index, current := range value {
		if index > 0 && unicode.IsUpper(current) && (unicode.IsLower(rune(value[index-1])) || unicode.IsDigit(rune(value[index-1]))) {
			output = append(output, ' ')
		}
		output = append(output, current)
	}
	if len(output) == 0 {
		return value
	}
	output[0] = unicode.ToUpper(output[0])
	return string(output)
}

func findType(name string) (typeDefinition, bool) {
	definitions := typesByName[name]
	if len(definitions) == 0 {
		return typeDefinition{}, false
	}
	return definitions[0], true
}

func requestParts(name string) (map[string]any, map[string]any, map[string]any, map[string]any) {
	definition, ok := findType(name)
	if !ok {
		return objectSchema(), objectSchema(), objectSchema(), objectSchema()
	}
	structure, ok := definition.expr.(*ast.StructType)
	if !ok {
		return objectSchema(), objectSchema(), objectSchema(), objectSchema()
	}
	for _, field := range structure.Fields.List {
		indices, ok := field.Type.(*ast.IndexListExpr)
		if !ok || len(indices.Indices) != 4 {
			continue
		}
		return schemaFor(indices.Indices[0], true, map[string]bool{}),
			schemaFor(indices.Indices[1], true, map[string]bool{}),
			schemaFor(indices.Indices[2], true, map[string]bool{}),
			schemaFor(indices.Indices[3], true, map[string]bool{})
	}
	return objectSchema(), objectSchema(), objectSchema(), objectSchema()
}

// Some existing DTOs place URL-bound values in Param even when the route does
// not contain a matching path placeholder. The binders read those values from
// the query string, so expose every non-path Param field as a query parameter.
func routeQuery(path string, param, query map[string]any) map[string]any {
	result := objectSchema()
	resultProperties := properties(result)
	required := map[string]bool{}
	pathParameters := map[string]bool{}
	for _, name := range pathNames(path) {
		pathParameters[name] = true
	}
	copyProperties := func(source map[string]any, skipPath bool) {
		for name, schema := range properties(source) {
			if skipPath && (pathParameters[name] || pathParameters[kebabCase(name)]) {
				continue
			}
			resultProperties[name] = schema
			if isRequired(source, name) {
				required[name] = true
			}
		}
	}
	copyProperties(param, true)
	copyProperties(query, false)
	if len(required) > 0 {
		names := make([]string, 0, len(required))
		for name := range required {
			names = append(names, name)
		}
		sort.Strings(names)
		result["required"] = names
	}
	return result
}
