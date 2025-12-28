package tracegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the configuration for trace generation
type Config struct {
	// Paths to scan for Go files (can be directories or specific files)
	Paths []string
	// FunctionPattern is a pattern to match function/method names (e.g., "Execute")
	// If empty, all functions with context.Context as first param will be traced
	FunctionPattern string
	// TraceImportPath is the import path for the trace package
	TraceImportPath string
	// DryRun if true, will not write changes to files
	DryRun bool
	// Verbose enables verbose output
	Verbose bool
}

// Generator handles the AST manipulation for trace injection
type Generator struct {
	config Config
	fset   *token.FileSet
	logger *slog.Logger
}

// NewGenerator creates a new trace generator
func NewGenerator(config Config) *Generator {
	if config.TraceImportPath == "" {
		config.TraceImportPath = "github.com/cristiano-pacheco/go-otel/trace"
	}
	return &Generator{
		config: config,
		fset:   token.NewFileSet(),
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

// Generate processes all files in the configured paths
func (g *Generator) Generate() error {
	for _, path := range g.config.Paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if info.IsDir() {
			if processErr := g.processDirectory(path); processErr != nil {
				return processErr
			}
		} else {
			if processErr := g.processFile(path); processErr != nil {
				return processErr
			}
		}
	}
	return nil
}

// processDirectory recursively processes all Go files in a directory
func (g *Generator) processDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if processErr := g.processFile(path); processErr != nil {
				return fmt.Errorf("error processing %s: %w", path, processErr)
			}
		}
		return nil
	})
}

// processFile processes a single Go file
func (g *Generator) processFile(filename string) error {
	node, err := parser.ParseFile(g.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing file: %w", err)
	}

	modified, hasTraceImport := g.processAST(node)
	if !modified {
		return nil
	}

	// Add trace import if not present
	if hasTraceImport && !g.hasImport(node, g.config.TraceImportPath) {
		g.addImport(node, g.config.TraceImportPath)
	}

	// Write the modified file
	if g.config.DryRun {
		if g.config.Verbose {
			g.logger.Info("[DRY RUN] Would modify file", "file", filename)
		}
		return nil
	}

	return g.writeFile(filename, node)
}

// processAST walks the AST and injects tracing code
func (g *Generator) processAST(node *ast.File) (bool, bool) {
	modified := false
	hasTraceImport := false

	// Check if trace package is already imported
	for _, imp := range node.Imports {
		if imp.Path.Value == fmt.Sprintf(`"%s"`, g.config.TraceImportPath) {
			hasTraceImport = true
			break
		}
	}

	// Walk through the AST and inject tracing code
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if function should be traced
		if !g.shouldTraceFunction(funcDecl) {
			return true
		}

		// Check if tracing code already exists
		if g.hasTracingCode(funcDecl) {
			return true
		}

		// Inject tracing code
		if g.injectTracingCode(funcDecl) {
			modified = true
			hasTraceImport = true
		}

		return true
	})

	return modified, hasTraceImport
}

// shouldTraceFunction determines if a function should have tracing injected
func (g *Generator) shouldTraceFunction(funcDecl *ast.FuncDecl) bool {
	// Check if function has parameters
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return false
	}

	// Check if first parameter is context.Context
	firstParam := funcDecl.Type.Params.List[0]
	if !g.isContextType(firstParam.Type) {
		return false
	}

	// If pattern is specified, check function name
	if g.config.FunctionPattern != "" {
		return strings.Contains(funcDecl.Name.Name, g.config.FunctionPattern)
	}

	return true
}

// isContextType checks if a type is context.Context
func (g *Generator) isContextType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "context" && selector.Sel.Name == "Context"
}

// hasTracingCode checks if function already has tracing code
func (g *Generator) hasTracingCode(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
		return false
	}

	// Check first statement for trace.Span
	firstStmt := funcDecl.Body.List[0]
	assignStmt, ok := firstStmt.(*ast.AssignStmt)
	if !ok {
		return false
	}

	if len(assignStmt.Rhs) == 0 {
		return false
	}

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

	return ident.Name == "trace" && selector.Sel.Name == "Span"
}

// injectTracingCode injects tracing code at the beginning of a function
func (g *Generator) injectTracingCode(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil {
		return false
	}

	// Get function name with receiver if it's a method
	spanName := g.getSpanName(funcDecl)

	if g.config.Verbose {
		g.logger.Info("Injecting trace", "function", spanName)
	}

	// Create: ctx, span := trace.Span(ctx, "SpanName")
	startSpanStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "ctx"},
			&ast.Ident{Name: "span"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "trace"},
					Sel: &ast.Ident{Name: "Span"},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: "ctx"},
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf(`"%s"`, spanName),
					},
				},
			},
		},
	}

	// Create: defer span.End()
	deferStmt := &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "span"},
				Sel: &ast.Ident{Name: "End"},
			},
		},
	}

	// Insert at the beginning of the function body
	funcDecl.Body.List = append(
		[]ast.Stmt{startSpanStmt, deferStmt},
		funcDecl.Body.List...,
	)

	return true
}

// getSpanName generates the span name for a function
func (g *Generator) getSpanName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return funcDecl.Name.Name
	}

	// Get receiver type name
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

// hasImport checks if a package is already imported
func (g *Generator) hasImport(node *ast.File, importPath string) bool {
	for _, imp := range node.Imports {
		if imp.Path.Value == fmt.Sprintf(`"%s"`, importPath) {
			return true
		}
	}
	return false
}

// addImport adds an import to the file
func (g *Generator) addImport(node *ast.File, importPath string) {
	// Find the import declaration or create one
	var importDecl *ast.GenDecl
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf(`"%s"`, importPath),
		},
	}

	if importDecl == nil {
		// Create a new import declaration
		importDecl = &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: `"context"`,
					},
				},
			},
		}
		node.Decls = append([]ast.Decl{importDecl}, node.Decls...)
	}

	importDecl.Specs = append(importDecl.Specs, newImport)
}

// writeFile writes the modified AST back to the file
func (g *Generator) writeFile(filename string, node *ast.File) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, g.fset, node); err != nil {
		return fmt.Errorf("error formatting node: %w", err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	if g.config.Verbose {
		g.logger.Info("Modified file", "file", filename)
	}

	return nil
}
