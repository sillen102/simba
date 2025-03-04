package simba

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	simbaHttp "github.com/sillen102/simba/http"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

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

type securityHandlerInfo struct {
	securityScheme      authType
	securityName        string
	securityDescription string
	format              string
	fieldName           string
	in                  openapi.In
}

type authType int

// Auth types
const (
	none = iota
	basicAuth
	apiKey
	bearerAuth
)

// Tags for parsing comments
const (
	idTag          = "@ID"
	tagTag         = "@Tag"
	summaryTag     = "@Summary"
	descriptionTag = "@Description"
	statusCodeTag  = "@StatusCode"
	errorTag       = "@Error"
	deprecatedTag  = "@Deprecated"
	basicAuthTag   = "@BasicAuth"
	apiKeyAuthTag  = "@APIKeyAuth"
	bearerAuthTag  = "@BearerAuth"
)

func generateRouteDocumentation(reflector *openapi31.Reflector, routeInfo *routeInfo, handler any) {
	operationContext, err := reflector.NewOperationContext(routeInfo.method, routeInfo.path)
	if err != nil {
		panic(fmt.Errorf("failed to create operation context: %w", err))
	}

	// Get the handler value and type using reflection
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	info := getHandlerInfo(handler)

	operationContext.SetIsDeprecated(info.deprecated)
	operationContext.SetID(info.id)
	operationContext.SetTags(info.tags...)
	operationContext.SetSummary(info.summary)
	operationContext.SetDescription(info.description)

	// Add request body if it exists
	if routeInfo.reqBody != nil {
		operationContext.AddReqStructure(routeInfo.reqBody, func(cu *openapi.ContentUnit) {
			cu.ContentType = routeInfo.accepts
		})
	}

	// Add params if they exist
	if routeInfo.params != nil {
		operationContext.AddReqStructure(routeInfo.params)
	}

	// Get response status code
	if info.statusCode == 0 {
		code := getHandlerResponseStatus(handlerValue, handlerType)
		if code == 0 {
			if routeInfo.respBody == (NoBody{}) {
				info.statusCode = http.StatusNoContent
			} else {
				info.statusCode = http.StatusOK
			}
		} else {
			info.statusCode = code
		}
	}

	// Add response with the status code
	operationContext.AddRespStructure(routeInfo.respBody, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = info.statusCode
		cu.ContentType = routeInfo.produces
	})

	// Add default error responses
	operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusBadRequest
		cu.Description = "Request body contains invalid data"
	})
	operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnprocessableEntity
		cu.Description = "Request body could not be processed"
	})
	operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusInternalServerError
		cu.Description = "Unexpected error"
	})

	// Add custom error responses
	for _, e := range info.errors {
		operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = e.Code
			cu.Description = e.Message
		})
	}

	// Add security if authenticated route
	if routeInfo.authFunc != nil {
		authFuncType := reflect.TypeOf(routeInfo.authFunc)
		authFuncValue := reflect.ValueOf(routeInfo.authFunc)
		secComment := getHandlerComment(authFuncValue, authFuncType)
		sec := parseAuthFuncComment(secComment)

		switch sec.securityScheme {
		case none:
			// Do nothing
		case basicAuth:
			reflector.SpecEns().SetHTTPBasicSecurity(sec.securityName, sec.securityDescription)
		case apiKey:
			reflector.SpecEns().SetAPIKeySecurity(
				sec.securityName,
				sec.fieldName,
				sec.in,
				sec.securityDescription,
			)
		case bearerAuth:
			reflector.SpecEns().SetHTTPBearerTokenSecurity(
				sec.securityName,
				sec.format,
				sec.securityDescription,
			)
		}

		operationContext.AddSecurity(sec.securityName)

		operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = http.StatusUnauthorized
			cu.Description = "Authorization failed"
		})
		operationContext.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = http.StatusForbidden
			cu.Description = "Access denied"
		})
	}

	err = reflector.AddOperation(operationContext)
	if err != nil {
		panic(fmt.Errorf("failed to add operation to openapi reflector: %w", err))
	}
}

func getHandlerInfo(handler any) handlerInfo {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	comment := getHandlerComment(handlerValue, handlerType)
	info := parseHandlerComment(comment)

	if info.id == "" {
		info.id = strcase.ToKebab(strcase.ToKebab(getFunctionName(handlerValue.Interface())))
	}

	if len(info.tags) == 0 {
		info.tags = []string{strcase.ToCamel(getPackageName(handlerValue.Interface()))}
	}

	if info.summary == "" {
		info.summary = camelToSpaced(strcase.ToCamel(getFunctionName(handlerValue.Interface())))
	}

	if info.description == "" {
		info.description = getCommentStrippedFromTags(comment)
	}

	return info
}

// getSourceCode reads a file and returns its source code content
func getSourceCode(fileName string) ([]byte, error) {
	if fileName == "" || strings.Contains(fileName, "<autogenerated>") {
		return nil, fmt.Errorf("invalid file name")
	}
	return os.ReadFile(fileName)
}

// parseSourceFile parses source code into an AST
func parseSourceFile(fileName string, src []byte) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, fileName, src, parser.ParseComments)
}

func getHandlerComment(handlerValue reflect.Value, handlerType reflect.Type) string {
	// Skip comment extraction for non-function types
	if handlerType.Kind() != reflect.Func {
		return ""
	}

	// Get the function pointer
	pc := handlerValue.Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	fullName := fn.Name()
	fileName, _ := fn.FileLine(pc)
	var methodName string

	// Check if the method has a receiver
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		methodName = fullName
	} else {
		methodName = strings.Replace(fullName[lastDot+1:], "-fm", "", 1)
	}

	// For test functions or auto-generated code, scan project files
	if fileName == "" || strings.Contains(fileName, "<autogenerated>") {
		return findCommentByScanning(methodName)
	}

	// Parse the source file directly
	content, err := os.ReadFile(fileName)
	if err != nil {
		return findCommentByScanning(methodName) // Fallback to scanning
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, content, parser.ParseComments)
	if err != nil {
		return findCommentByScanning(methodName) // Fallback to scanning
	}

	return extractCommentForFunction(node, methodName)
}

func findCommentByScanning(methodName string) string {
	// Get current directory to start the search
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Find all Go files recursively up to 2 levels deep
	var files []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Only consider .go files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, path)
		}

		// Limit directory depth to avoid excessive scanning
		if info.IsDir() && path != currentDir {
			components := strings.Split(path, string(os.PathSeparator))
			depth := 0
			for i := len(components) - 1; i >= 0; i-- {
				if components[i] == "." {
					continue
				}
				depth++
				if depth > 2 { // Max 2 levels deep
					return filepath.SkipDir
				}
			}
		}

		return nil
	}

	_ = filepath.Walk(currentDir, walkFn)

	// Search each file for the method
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, content, parser.ParseComments)
		if err != nil {
			continue
		}

		comment := extractCommentForFunction(node, methodName)
		if comment != "" {
			return comment
		}
	}

	return ""
}

// Helper function to extract comment for a specific function
func extractCommentForFunction(node *ast.File, methodName string) string {
	var comment string
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == methodName {
				if funcDecl.Doc != nil {
					comment = funcDecl.Doc.Text()
				}
				return false
			}
		}
		return true
	})
	return strings.TrimSpace(comment)
}

func parseHandlerComment(comment string) handlerInfo {
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

func parseAuthFuncComment(comment string) securityHandlerInfo {
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	sec := securityHandlerInfo{}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, basicAuthTag):
			basicLine := strings.TrimSpace(strings.TrimPrefix(line, basicAuthTag))
			parts := strings.SplitN(basicLine, " ", 2)

			if len(parts) >= 2 {
				sec.securityScheme = basicAuth
				sec.securityName = strings.Replace(parts[0], "\"", "", -1)
				sec.securityDescription = strings.Replace(parts[1], "\"", "", -1)
			}
		case strings.HasPrefix(line, apiKeyAuthTag):
			apiKeyLine := strings.TrimSpace(strings.TrimPrefix(line, apiKeyAuthTag))
			parts := strings.SplitN(apiKeyLine, " ", 4)

			if len(parts) >= 4 {
				sec.securityScheme = apiKey
				sec.securityName = strings.Replace(parts[0], "\"", "", -1)
				sec.fieldName = strings.Replace(parts[1], "\"", "", -1)
				sec.in = openapi.In(strings.Replace(parts[2], "\"", "", -1))
				sec.securityDescription = strings.Replace(parts[3], "\"", "", -1)
			}
		case strings.HasPrefix(line, bearerAuthTag):
			bearerLine := strings.TrimSpace(strings.TrimPrefix(line, bearerAuthTag))
			parts := strings.SplitN(bearerLine, " ", 3)

			if len(parts) >= 3 {
				sec.securityScheme = bearerAuth
				sec.securityName = strings.Replace(parts[0], "\"", "", -1)
				sec.format = strings.Replace(parts[1], "\"", "", -1)
				sec.securityDescription = strings.Replace(parts[2], "\"", "", -1)
			}
		}
	}

	return sec
}

func getHandlerResponseStatus(handlerValue reflect.Value, handlerType reflect.Type) int {
	if handlerType.Kind() != reflect.Func {
		return 0
	}

	pc := handlerValue.Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return 0
	}

	fullName := fn.Name()
	fileName, _ := fn.FileLine(pc)

	// Extract method name
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		return 0
	}
	methodName := strings.Replace(fullName[lastDot+1:], "-fm", "", 1)

	// Try the original file first
	content, err := getSourceCode(fileName)
	if err == nil {
		node, err := parseSourceFile(fileName, content)
		if err == nil {
			status := findStatusInAST(node, methodName)
			if status != 0 {
				return status
			}
		}
	}

	// Fallback to scanning all Go files in the current directory and subdirectories
	return findStatusByScanning(methodName)
}

func findStatusByScanning(methodName string) int {
	// Get current directory to start the search
	currentDir, err := os.Getwd()
	if err != nil {
		return 0
	}

	// Find all Go files recursively up to 2 levels deep
	var files []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Only consider .go files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, path)
		}

		// Limit directory depth
		if info.IsDir() && path != currentDir {
			components := strings.Split(path, string(os.PathSeparator))
			depth := 0
			for i := len(components) - 1; i >= 0; i-- {
				if components[i] == "." {
					continue
				}
				depth++
				if depth > 2 { // Max 2 levels deep
					return filepath.SkipDir
				}
			}
		}
		return nil
	}

	_ = filepath.Walk(currentDir, walkFn)

	// Search each file for the method
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, content, parser.ParseComments)
		if err != nil {
			continue
		}

		status := findStatusInAST(node, methodName)
		if status != 0 {
			return status
		}
	}

	return 0
}

func findStatusInAST(node *ast.File, methodName string) int {
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
									// Handle basic integer literal
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

func getPackageName(handler any) string {
	pc := reflect.ValueOf(handler).Pointer()
	fn := runtime.FuncForPC(pc)
	fullPath := fn.Name()

	// Split the full path into parts
	parts := strings.Split(fullPath, "/")
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

func getFunctionName(i any) string {
	// Get the function value
	function := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	// Get the full function name (includes package path)

	fullName := function.Name()
	var methodName string

	// Check if the method has a receiver
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 {
		methodName = fullName
	} else {
		methodName = strings.Replace(fullName[lastDot+1:], "-fm", "", 1)
	}

	// Extract just the function name
	if idx := strings.LastIndex(methodName, "."); idx != -1 {
		return methodName[idx+1:]
	}
	return methodName
}

func getCommentStrippedFromTags(comment string) string {
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	result := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "@") {
			continue
		}
		result += line + "\n"
	}

	return strings.TrimSpace(result)
}

func camelToSpaced(s string) string {
	words := strcase.ToDelimited(s, ' ')
	words = strings.ToLower(words)
	return strings.ToUpper(words[:1]) + words[1:]
}
