package simba

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func getFunctionComment(handler any) string {
	handlerValue := reflect.ValueOf(handler)
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
