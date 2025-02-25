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

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type commentInfo struct {
	description string
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
	descriptionTag = "@Description"
	errorTag       = "@Error"
	basicAuthTag   = "@BasicAuth"
	apiKeyAuthTag  = "@APIKeyAuth"
	bearerAuthTag  = "@BearerAuth"
)

func generateRouteDocumentation(reflector *openapi31.Reflector, routeInfo *routeInfo, handler Handler) {
	operationContext, err := reflector.NewOperationContext(routeInfo.method, routeInfo.path)
	if err != nil {
		panic(fmt.Errorf("failed to create operation context: %w", err))
	}

	// Parse function comments
	comment := getFunctionComment(handler)
	info := parseHandlerComment(comment)

	// Add request body if it exists
	if routeInfo.reqBody != nil {
		operationContext.AddReqStructure(routeInfo.reqBody, func(cu *openapi.ContentUnit) {
			cu.ContentType = routeInfo.accepts
			cu.Description = info.description
		})
	}

	// Add params if they exist
	if routeInfo.params != nil {
		operationContext.AddReqStructure(routeInfo.params)
	}

	// Add response with 200 status code
	operationContext.AddRespStructure(routeInfo.respBody, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusOK
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
		secComment := getFunctionComment(routeInfo.authFunc)
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
	}

	err = reflector.AddOperation(operationContext)
	if err != nil {
		panic(fmt.Errorf("failed to add operation to openapi reflector: %w", err))
	}
}

func getFunctionComment(function any) string {
	handlerValue := reflect.ValueOf(function)
	handlerType := handlerValue.Type()

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
		errors: make([]struct {
			Code    int
			Message string
		}, 0),
	}

	var descLines []string
	insideDesc := false

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, descriptionTag):
			insideDesc = true
			text := strings.TrimSpace(strings.TrimPrefix(line, descriptionTag))
			if text != "" {
				descLines = append(descLines, text)
			}
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
