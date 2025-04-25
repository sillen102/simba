package simbaOpenapi

import (
	"go/ast"
	"sync"
)

// fileCache stores parsed files and their functions
type fileCache struct {
	files map[string]*ast.File       // Cache of parsed files
	funcs map[string]map[string]bool // Map of filename to function names
	mutex sync.RWMutex               // Mutex for thread safety
}

// newFileCache creates a new file cache
func newFileCache() *fileCache {
	return &fileCache{
		files: make(map[string]*ast.File),
		funcs: make(map[string]map[string]bool),
	}
}

// add adds a file and its functions to the cache
func (c *fileCache) add(filename string, file *ast.File) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.files[filename] = file
	c.funcs[filename] = make(map[string]bool)

	// Index all functions in the file
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			c.funcs[filename][fd.Name.Name] = true
		}
		return true
	})
}

// findFunction looks for a function in the cache
func (c *fileCache) findFunction(functionName string) *ast.File {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for filename, funcs := range c.funcs {
		if funcs[functionName] {
			return c.files[filename]
		}
	}
	return nil
}

// hasFunction checks if a function exists in any cached file
func (c *fileCache) hasFunction(functionName string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, funcs := range c.funcs {
		if funcs[functionName] {
			return true
		}
	}
	return false
}
