package simbaOpenapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	simbaHttp "github.com/sillen102/simba/http"
	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type OpenAPIGenerator struct {
}

type handlerInfo struct {
	id          string
	tags        []string
	summary     string
	description string
	statusCode  int
	deprecated  bool
	errors      []struct {
		Code    int
		Message string
	}
}

// Tags for parsing comments
const (
	idTag          = "@ID"
	tagTag         = "@Tag"
	summaryTag     = "@Summary"
	descriptionTag = "@Description"
	statusCodeTag  = "@StatusCode"
	errorTag       = "@Error"
	deprecatedTag  = "@Deprecated"
)

func NewOpenAPIGenerator() *OpenAPIGenerator {
	return &OpenAPIGenerator{}
}

// GenerateDocumentation generates OpenAPI documentation for all routes
func (g *OpenAPIGenerator) GenerateDocumentation(routeInfos []openapiModels.RouteInfo, mimetype string) ([]byte, error) {
	reflector, err := GetReflector()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI reflector: %w", err)
	}

	for _, routeInfo := range routeInfos {
		err = generateRouteDocumentation(reflector, &routeInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate documentation for route: %w", err)
		}
	}

	var schema []byte
	if mimetype == mimetypes.ApplicationJSON {
		schema, err = reflector.Spec.MarshalJSON()
	} else {
		schema, err = reflector.Spec.MarshalYAML()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAPI schema: %w", err)
	}

	return schema, nil
}

// generateRouteDocumentation generates OpenAPI documentation for a route
func generateRouteDocumentation(reflector *openapi31.Reflector, routeInfo *openapiModels.RouteInfo) error {
	operationContext, err := reflector.NewOperationContext(routeInfo.Method, routeInfo.Path)
	if err != nil {
		return err
	}

	info := getHandlerInfo(routeInfo.Handler)

	operationContext.SetIsDeprecated(info.deprecated)
	operationContext.SetID(info.id)
	operationContext.SetTags(info.tags...)
	operationContext.SetSummary(info.summary)
	operationContext.SetDescription(info.description)

	// Add request body if it exists
	if routeInfo.ReqBody != nil {
		operationContext.AddReqStructure(routeInfo.ReqBody, func(cu *openapi.ContentUnit) {
			cu.ContentType = routeInfo.Accepts
		})
	}

	// Add params if they exist
	if routeInfo.Params != nil {
		operationContext.AddReqStructure(routeInfo.Params)
	}

	// Get response status code
	if info.statusCode == 0 {
		if routeInfo.RespBody == (simbaModels.NoBody{}) {
			info.statusCode = http.StatusNoContent // Default for no response body
		} else {
			info.statusCode = http.StatusOK // Default for response body
		}
	}

	// Add response with the status code
	operationContext.AddRespStructure(routeInfo.RespBody, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = info.statusCode
		cu.ContentType = routeInfo.Produces
	})

	// Add default error responses
	operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusBadRequest
		cu.Description = "Request body contains invalid data"
	})
	operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnprocessableEntity
		cu.Description = "Request body could not be processed"
	})
	operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusInternalServerError
		cu.Description = "Unexpected error"
	})

	// Add custom error responses
	for _, e := range info.errors {
		operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = e.Code
			cu.Description = e.Message
		})
	}

	// Add security if authenticated route
	if routeInfo.AuthHandler != nil {
		authHandler, ok := routeInfo.AuthHandler.(interface {
			GetType() openapiModels.AuthType
			GetName() string
			GetFieldName() string
			GetFormat() string
			GetDescription() string
			GetIn() openapi.In
		})

		if ok {
			switch authHandler.GetType() {
			case openapiModels.AuthTypeBasic:
				ah, _ := routeInfo.AuthHandler.(interface {
					GetName() string
					GetDescription() string
				})
				reflector.SpecEns().SetHTTPBasicSecurity(ah.GetName(), ah.GetDescription())
			case openapiModels.AuthTypeAPIKey:
				reflector.SpecEns().SetAPIKeySecurity(
					authHandler.GetName(),
					authHandler.GetFieldName(),
					authHandler.GetIn(),
					authHandler.GetDescription(),
				)
			case openapiModels.AuthTypeBearer:
				reflector.SpecEns().SetHTTPBearerTokenSecurity(
					authHandler.GetName(),
					authHandler.GetFormat(),
					authHandler.GetDescription(),
				)
			}

			operationContext.AddSecurity(authHandler.GetName())

			operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
				cu.HTTPStatus = http.StatusUnauthorized
				cu.Description = "Authorization failed"
			})
			operationContext.AddRespStructure(simbaErrors.ErrorResponse{}, func(cu *openapi.ContentUnit) {
				cu.HTTPStatus = http.StatusForbidden
				cu.Description = "Access denied"
			})
		}
	}

	err = reflector.AddOperation(operationContext)
	if err != nil {
		return err
	}

	return nil
}

// getHandlerInfo extracts the handler information from the handler function
func getHandlerInfo(handler any) handlerInfo {
	functionPointer := getFunctionPointer(handler)
	runTimeFunc := getFuncRuntime(functionPointer)
	functionFullName := getFunctionFullName(runTimeFunc)
	functionPackagePath := extractPackagePath(functionFullName)
	functionFile := getFunctionASTFile(functionPackagePath)
	methodName := extractMethodNameWithoutReceiver(functionFullName)
	functionComment := extractCommentForFunction(functionFile, methodName)

	info := parseHandlerCommentTags(functionComment)

	if info.id == "" {
		info.id = strcase.ToKebab(methodName)
	}

	if len(info.tags) == 0 {
		info.tags = []string{strcase.ToCamel(getPackageName(functionFullName))}
	}

	if info.summary == "" {
		info.summary = camelToSpaced(strcase.ToCamel(methodName))
	}

	if info.description == "" {
		info.description = getCommentStrippedFromTags(functionComment, methodName)
	}

	if info.statusCode == 0 {
		info.statusCode = findStatusInAST(functionFile, methodName)
	}

	return info
}

// getFunctionPointer gets the function pointer for a handler
func getFunctionPointer(handler any) uintptr {
	return reflect.ValueOf(handler).Pointer()
}

// getFuncRuntime gets the runtime function for a handler
func getFuncRuntime(handlerPointer uintptr) *runtime.Func {
	return runtime.FuncForPC(handlerPointer)
}

// getFunctionFullName gets the full name of a function using its pointer
func getFunctionFullName(fn *runtime.Func) string {
	return fn.Name()
}

// getFunctionASTFile finds the Go source file containing a handler function
func getFunctionASTFile(packagePath string) *ast.File {
	// For receiver methods, search in the correct package directory
	if packagePath != "" {
		// Get the physical path on disk for the package
		pkgDir := findPackageDir(packagePath)
		if pkgDir != "" {
			// Search all Go files in this directory
			files, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
			if err == nil {
				for _, file := range files {
					node, err := parseFile(file)
					if err != nil {
						continue
					}

					return node
				}
			}
		}
	}

	return nil
}

// parseFile parses a file and returns its AST
func parseFile(fileName string) (*ast.File, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// extractPackagePath gets the package path from a full function name
func extractPackagePath(fullName string) string {
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		return ""
	}

	// For receiver methods, we need to find the second-to-last dot
	// or the dot before the opening parenthesis
	if idx := strings.Index(fullName, "("); idx != -1 && idx < lastDot {
		return fullName[:idx-1] // -1 to remove the dot
	}

	// For regular functions, package path is everything before the last dot
	parts := strings.Split(fullName, ".")
	if len(parts) < 2 {
		return ""
	}

	return strings.Join(parts[:len(parts)-1], ".")
}

// findPackageDir converts a Go import path to a filesystem path
func findPackageDir(importPath string) string {
	// Try to use GOPATH first
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		pkgDir := filepath.Join(gopath, "src", importPath)
		if _, err := os.Stat(pkgDir); err == nil {
			return pkgDir
		}
	}

	// Try to use Go modules
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", importPath)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return ""
}

// extractMethodNameWithoutReceiver gets just the method name from a full function name
func extractMethodNameWithoutReceiver(fullName string) string {
	// Handle methods with receivers (with potential "-fm" suffix)
	// e.g., "github.com/package.(*Type).Method-fm" -> "Method"
	if idx := strings.LastIndex(fullName, "."); idx != -1 {
		name := fullName[idx+1:]
		// Remove "-fm" suffix
		name = strings.Replace(name, "-fm", "", 1)
		return name
	}
	return fullName
}

// getSimpleMethodName extracts just the method name without any package or receiver info
func getSimpleMethodName(fullName string) string {
	// Get the part after the last dot, which should be the method name
	if idx := strings.LastIndex(fullName, "."); idx >= 0 && idx < len(fullName)-1 {
		name := fullName[idx+1:]
		// Remove any "-fm" suffix that Go adds to method function values (e.g., "Method-fm") for methods with receivers
		return strings.Replace(name, "-fm", "", 1)
	}
	return fullName
}

// extractCommentForFunction extracts comment for a specific function
func extractCommentForFunction(node *ast.File, methodName string) string {
	if node == nil {
		return ""
	}

	var comment string

	// Clean the method name to get just the base name
	simpleName := getSimpleMethodName(methodName)

	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == simpleName {
				if funcDecl.Doc != nil {
					comment = funcDecl.Doc.Text()
					return false
				}
			}
		}
		return true
	})

	return strings.TrimSpace(comment)
}

// parseHandlerCommentTags parses the comment for a handler function and extracts information from comment tags
func parseHandlerCommentTags(comment string) handlerInfo {
	lines := strings.Split(strings.TrimSpace(comment), "\n")

	info := handlerInfo{
		tags: make([]string, 0),
		errors: make([]struct {
			Code    int
			Message string
		}, 0),
	}

	var descLines []string
	insideDesc := false

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, idTag):
			info.id = strings.TrimSpace(strings.TrimPrefix(line, idTag))
		case strings.HasPrefix(line, tagTag):
			tag := strings.TrimSpace(strings.TrimPrefix(line, tagTag))
			info.tags = append(info.tags, tag)
		case strings.HasPrefix(line, summaryTag):
			info.summary = strings.TrimSpace(strings.TrimPrefix(line, summaryTag))
		case strings.HasPrefix(line, descriptionTag):
			insideDesc = true
			text := strings.TrimSpace(strings.TrimPrefix(line, descriptionTag))
			if text != "" {
				descLines = append(descLines, text)
			}
		case strings.HasPrefix(line, statusCodeTag):
			code, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, statusCodeTag)))
			if err != nil {
				continue
			}
			info.statusCode = code
		case strings.HasPrefix(line, deprecatedTag):
			info.deprecated = true
		case strings.HasPrefix(line, errorTag):
			insideDesc = false
			errorLine := strings.TrimSpace(strings.TrimPrefix(line, errorTag))
			// Then split on @Error
			parts := strings.SplitN(errorLine, " ", 2)
			if len(parts) >= 2 {
				code, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				info.errors = append(info.errors, struct {
					Code    int
					Message string
				}{Code: code, Message: parts[1]})
			}
		case insideDesc:
			descLines = append(descLines, line)
		case strings.HasPrefix(line, "@"):
			insideDesc = false
		}
	}

	info.description = strings.Join(descLines, "\n")

	return info
}

// findStatusInAST looks for status codes in the AST
func findStatusInAST(node *ast.File, methodName string) int {
	if node == nil {
		return 0
	}

	var status int

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != methodName {
			return true
		}

		// We found the function, now look for return statements
		ast.Inspect(funcDecl, func(n ast.Node) bool {
			// Look for return statements
			ret, ok := n.(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				return true
			}

			// Check if we're returning a response object
			for _, result := range ret.Results {
				// Try to find Status field in composite literals
				if unary, ok := result.(*ast.UnaryExpr); ok {
					if cl, ok := unary.X.(*ast.CompositeLit); ok {
						for _, elt := range cl.Elts {
							if kv, ok := elt.(*ast.KeyValueExpr); ok {
								if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Status" {
									if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
										if s, err := strconv.Atoi(basicLit.Value); err == nil {
											status = s
											return false
										}
									}

									// Handle HTTP constant (e.g., http.StatusOK)
									if selExpr, ok := kv.Value.(*ast.SelectorExpr); ok {
										if pkgIdent, ok := selExpr.X.(*ast.Ident); ok && pkgIdent.Name == "http" {
											if code, ok := simbaHttp.HTTPStatusMapping[selExpr.Sel.Name]; ok {
												status = code
												return false
											}
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})
		return false // Stop searching after finding the function
	})

	return status
}

// getPackageName extracts the package name for a handler function given its full name
func getPackageName(fullName string) string {
	// Split the full path into parts
	parts := strings.Split(fullName, "/")
	// Get the last part which contains package.function
	if len(parts) == 0 {
		return ""
	}
	lastPart := parts[len(parts)-1]
	// Split package.function and take the package name
	pkgAndFunc := strings.Split(lastPart, ".")
	if len(pkgAndFunc) > 1 {
		return pkgAndFunc[0]
	}
	return lastPart
}

// getCommentStrippedFromTags removes tags from a comment so that only the description remains
func getCommentStrippedFromTags(comment string, methodName string) string {
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	result := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "@") {
			continue
		}
		result += line + "\n"
	}

	comment = strings.TrimSpace(result)
	if strings.HasPrefix(comment, methodName) {
		comment = strings.TrimSpace(strings.TrimPrefix(comment, methodName))
	}

	return comment
}

// camelToSpaced converts a camel case string to a spaced string
func camelToSpaced(s string) string {
	words := strcase.ToDelimited(s, ' ')
	words = strings.ToLower(words)
	return strings.ToUpper(words[:1]) + words[1:]
}
