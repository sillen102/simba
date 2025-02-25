package simba

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type commentInfo struct {
	Description string
	Errors      []struct {
		Code    int
		Message string
	}
}

func generateRouteDocumentation(reflector *openapi31.Reflector, routeInfo *routeInfo, handler any) {
	op, err := reflector.NewOperationContext(routeInfo.method, routeInfo.path)
	if err != nil {
		panic(fmt.Errorf("failed to create operation context: %w", err))
	}

	// Parse function comments
	comment := getFunctionComment(handler)
	info := parseFunctionComment(comment)

	// Add request body if it exists
	if routeInfo.reqBody != nil {
		op.AddReqStructure(routeInfo.reqBody, func(cu *openapi.ContentUnit) {
			cu.ContentType = routeInfo.accepts
			cu.Description = info.Description
		})
	}

	// Add params if they exist
	if routeInfo.params != nil {
		op.AddReqStructure(routeInfo.params)
	}

	// Add response with 200 status code
	op.AddRespStructure(routeInfo.respBody, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusOK
		cu.ContentType = routeInfo.produces
	})

	// Add default error responses
	op.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusBadRequest
		cu.Description = "Request body contains invalid data"
	})
	op.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusUnprocessableEntity
		cu.Description = "Request body could not be processed"
	})
	op.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusInternalServerError
		cu.Description = "Unexpected error"
	})

	// Add custom error responses
	for _, e := range info.Errors {
		op.AddRespStructure(ErrorResponse{}, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = e.Code
			cu.Description = e.Message
		})
	}

	err = reflector.AddOperation(op)
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

	// Handle both direct functions and method values
	var pc uintptr
	if handlerValue.Kind() == reflect.Func {
		if handlerValue.IsValid() && !handlerValue.IsNil() {
			pc = handlerValue.Pointer()
		} else {
			return ""
		}
	} else {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil || strings.Contains(fn.Name(), ".func") {
		return "" // Skip anonymous functions
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

	// Get the function name
	funcName := filepath.Base(fn.Name())
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

func parseFunctionComment(comment string) commentInfo {
	lines := strings.Split(strings.TrimSpace(comment), "\n")

	info := commentInfo{
		Errors: make([]struct {
			Code    int
			Message string
		}, 0),
	}

	var descLines []string
	insideDesc := false

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "@Description"):
			insideDesc = true
			text := strings.TrimSpace(strings.TrimPrefix(line, "@Description"))
			if text != "" {
				descLines = append(descLines, text)
			}
		case strings.HasPrefix(line, "@Error"):
			insideDesc = false
			errorLine := strings.TrimSpace(strings.TrimPrefix(line, "@Error"))
			// Then split on @Error
			parts := strings.SplitN(errorLine, " ", 2)
			if len(parts) >= 2 {
				code, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				info.Errors = append(info.Errors, struct {
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

	info.Description = strings.Join(descLines, "\n")

	return info
}
