package simba

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	simbaHttp "github.com/sillen102/simba/http"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type commentInfo struct {
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

type security struct {
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

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	handlerName := getFunctionName(handler)

	// Parse function comments
	comment := getFunctionComment(handlerValue, handlerType)
	info := parseHandlerComment(comment)

	operationContext.SetIsDeprecated(info.deprecated)

	// Add ID
	if info.id != "" {
		operationContext.SetID(info.id)
	} else {
		operationContext.SetID(strcase.ToKebab(handlerName))
	}

	// Add tags
	if len(info.tags) > 0 {
		operationContext.SetTags(info.tags...)
	} else {
		operationContext.SetTags(strcase.ToCamel(getPackageName(handler)))
	}

	// Add summary
	if info.summary != "" {
		operationContext.SetSummary(info.summary)
	} else {
		operationContext.SetSummary(camelToSpaced(strcase.ToCamel(handlerName)))
	}

	// Add description
	if info.description != "" {
		operationContext.SetDescription(info.description)
	} else {
		operationContext.SetDescription(getCommentStrippedFromTags(comment))
	}

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
		if routeInfo.respBody == (NoBody{}) {
			info.statusCode = http.StatusNoContent
		} else {
			info.statusCode = getHandlerResponseStatus(handlerValue, handlerType)
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
		secComment := getFunctionComment(authFuncValue, authFuncType)
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

func getFunctionComment(handlerValue reflect.Value, handlerType reflect.Type) string {
	// Skip comment extraction for non-function types
	if handlerType.Kind() != reflect.Func {
		return ""
	}

	// Get the function pointer
	var pc uintptr
	if handlerValue.IsValid() && !handlerValue.IsNil() {
		pc = handlerValue.Pointer()
	} else {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	fileName, _ := fn.FileLine(pc)
	if fileName == "" {
		return ""
	}

	// Parse the source file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return ""
	}

	// Get the complete function name
	funcName := fn.Name()
	// Extract just the function name from the full path
	if idx := strings.LastIndex(funcName, "."); idx != -1 {
		funcName = funcName[idx+1:]
	}

	// Find the function declaration and its comments
	var comment string
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == funcName {
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

func parseHandlerComment(comment string) commentInfo {
	lines := strings.Split(strings.TrimSpace(comment), "\n")

	info := commentInfo{
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

func parseAuthFuncComment(comment string) security {
	lines := strings.Split(strings.TrimSpace(comment), "\n")
	sec := security{}

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

	// Get the function pointer
	var pc uintptr
	if handlerValue.IsValid() && !handlerValue.IsNil() {
		pc = handlerValue.Pointer()
	} else {
		return 0
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return 0
	}

	fileName, _ := fn.FileLine(pc)
	if fileName == "" {
		return 0
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return 0
	}

	// Extract just the function name from the full name.
	funcName := fn.Name()
	if idx := strings.LastIndex(funcName, "."); idx != -1 {
		funcName = funcName[idx+1:]
	}

	var status int
	ast.Inspect(node, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok || fd.Name.Name != funcName {
			return true
		}

		// Iterate over statements in the function body.
		for _, stmt := range fd.Body.List {
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				continue
			}

			// Assume the returned composite literal is wrapped with a pointer operator.
			unary, ok := ret.Results[0].(*ast.UnaryExpr)
			if !ok {
				continue
			}

			cl, ok := unary.X.(*ast.CompositeLit)
			if !ok {
				continue
			}

			// Look for the Status field in the composite literal.
			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				ident, ok := kv.Key.(*ast.Ident)
				if !ok || ident.Name != "Status" {
					continue
				}

				// Handle basic integer literal.
				if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
					s, err := strconv.Atoi(basicLit.Value)
					if err == nil {
						status = s
						return false // Stop inspecting further once found.
					}
				}

				// Handle SelectorExpr for constants (e.g., http.StatusCreated).
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

		return true
	})

	if status == 0 {
		return http.StatusOK // Default status code
	}

	return status
}

func getPackageName(handler any) string {
	pc := reflect.ValueOf(handler).Pointer()
	fn := runtime.FuncForPC(pc)
	fullPath := fn.Name()

	// Split the full path into parts
	parts := strings.Split(fullPath, "/")
	// Get the last part which contains package.function
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
	// Extract just the function name
	if idx := strings.LastIndex(fullName, "."); idx != -1 {
		return fullName[idx+1:]
	}
	return fullName
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
