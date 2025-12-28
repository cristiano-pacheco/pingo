package tracegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// RemovalConfig holds the configuration for trace removal
type RemovalConfig struct {
	// Paths to scan for Go files (can be directories or specific files)
	Paths []string
	// DryRun if true, will not write changes to files
	DryRun bool
	// Verbose enables verbose output
	Verbose bool
}

// Remover handles the removal of trace instrumentation
type Remover struct {
	config RemovalConfig
	fset   *token.FileSet
}

// NewRemover creates a new trace remover
func NewRemover(config RemovalConfig) *Remover {
	return &Remover{
		config: config,
		fset:   token.NewFileSet(),
	}
}

// Remove processes all files in the configured paths and removes tracing code
func (r *Remover) Remove() error {
	for _, path := range r.config.Paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if info.IsDir() {
			if err := r.processDirectory(path); err != nil {
				return err
			}
		} else {
			if err := r.processFile(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// processDirectory recursively processes all Go files in a directory
func (r *Remover) processDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if err := r.processFile(path); err != nil {
				return fmt.Errorf("error processing %s: %w", path, err)
			}
		}
		return nil
	})
}

// processFile processes a single Go file
func (r *Remover) processFile(filename string) error {
	node, err := parser.ParseFile(r.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing file: %w", err)
	}

	modified := false

	// Walk through the AST and remove tracing code
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if r.removeTracingCode(funcDecl) {
			modified = true
		}

		return true
	})

	if !modified {
		return nil
	}

	// Write the modified file
	if r.config.DryRun {
		if r.config.Verbose {
			fmt.Printf("[DRY RUN] Would modify: %s\n", filename)
		}
		return nil
	}

	return r.writeFile(filename, node)
}

// removeTracingCode removes tracing code from a function
func (r *Remover) removeTracingCode(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil || len(funcDecl.Body.List) < 2 {
		return false
	}

	modified := false

	// Check if first two statements are trace.StartSpan and defer span.End()
	if r.isStartSpanStatement(funcDecl.Body.List[0]) {
		if r.config.Verbose {
			spanName := r.getSpanName(funcDecl)
			fmt.Printf("Removing trace from: %s\n", spanName)
		}

		// Remove the first statement (ctx, span := trace.StartSpan(...))
		funcDecl.Body.List = funcDecl.Body.List[1:]
		modified = true

		// Check if the new first statement is defer span.End()
		if len(funcDecl.Body.List) > 0 && r.isDeferSpanEndStatement(funcDecl.Body.List[0]) {
			funcDecl.Body.List = funcDecl.Body.List[1:]
		}
	}

	return modified
}

// isStartSpanStatement checks if a statement is ctx, span := trace.StartSpan(...)
func (r *Remover) isStartSpanStatement(stmt ast.Stmt) bool {
	assignStmt, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}

	if len(assignStmt.Lhs) != 2 || len(assignStmt.Rhs) != 1 {
		return false
	}

	// Check if left side is ctx, span
	ctx, ok1 := assignStmt.Lhs[0].(*ast.Ident)
	span, ok2 := assignStmt.Lhs[1].(*ast.Ident)
	if !ok1 || !ok2 || ctx.Name != "ctx" || span.Name != "span" {
		return false
	}

	// Check if right side is trace.StartSpan(...)
	callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "trace" && selector.Sel.Name == "StartSpan"
}

// isDeferSpanEndStatement checks if a statement is defer span.End()
func (r *Remover) isDeferSpanEndStatement(stmt ast.Stmt) bool {
	deferStmt, ok := stmt.(*ast.DeferStmt)
	if !ok {
		return false
	}

	selector, ok := deferStmt.Call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "span" && selector.Sel.Name == "End"
}

// getSpanName generates the span name for a function (for logging)
func (r *Remover) getSpanName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return funcDecl.Name.Name
	}

	receiverType := funcDecl.Recv.List[0].Type
	var typeName string

	switch t := receiverType.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			typeName = ident.Name
		}
	case *ast.Ident:
		typeName = t.Name
	}

	if typeName != "" {
		return fmt.Sprintf("%s.%s", typeName, funcDecl.Name.Name)
	}

	return funcDecl.Name.Name
}

// writeFile writes the modified AST back to the file
func (r *Remover) writeFile(filename string, node *ast.File) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, r.fset, node); err != nil {
		return fmt.Errorf("error formatting node: %w", err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	if r.config.Verbose {
		fmt.Printf("Modified: %s\n", filename)
	}

	return nil
}
