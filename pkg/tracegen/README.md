# Tracegen - Automatic Tracing Instrumentation

`tracegen` is an AST-based code generation tool that automatically injects OpenTelemetry tracing instrumentation into your Go code.

## Features

- **Automatic Trace Injection**: Automatically adds `trace.Span` and `defer span.End()` to functions
- **Smart Detection**: Only instruments functions with `context.Context` as the first parameter
- **Pattern Matching**: Filter functions by name pattern (e.g., only trace `Execute` methods)
- **Safe & Reversible**: Can remove previously added traces with `--remove` flag
- **Dry Run Mode**: Preview changes before applying them
- **Directory Support**: Process entire directories recursively

## Installation

The tool is already integrated into your project. No additional installation needed.

## Usage

### Basic Commands

```bash
# Generate traces for all files in a directory
make tracegen

# Preview what would be changed (dry run)
make tracegen-dry

# Remove existing traces
make tracegen-remove
```

### CLI Usage

```bash
# Inject traces into specific files
go run ./main.go tracegen --path ./internal/modules/identity/usecase

# Inject traces with function name pattern matching
go run ./main.go tracegen --path ./internal --pattern Execute

# Dry run to see what would be changed
go run ./main.go tracegen --path ./internal --dry-run --verbose

# Remove existing traces
go run ./main.go tracegen --path ./internal --remove

# Process multiple paths
go run ./main.go tracegen --path ./internal/modules/identity --path ./internal/modules/monitor
```

### Flags

- `--path, -p`: Path(s) to scan for Go files (required, can specify multiple)
- `--pattern`: Function name pattern to match (e.g., 'Execute')
- `--import`: Custom import path for trace package (default: "github.com/cristiano-pacheco/go-otel/trace")
- `--dry-run`: Show what would be changed without modifying files
- `--verbose, -v`: Enable verbose output
- `--remove`: Remove existing trace instrumentation

## How It Works

The tool uses Go's `ast` (Abstract Syntax Tree) package to:

1. Parse Go source files
2. Find functions that match the criteria:
   - Have `context.Context` as the first parameter
   - Match the specified pattern (if provided)
   - Don't already have tracing code
3. Inject the following code at the beginning of each matching function:
   ```go
   ctx, span := trace.Span(ctx, "TypeName.MethodName")
   defer span.End()
   ```
4. Add the trace import if not already present
5. Format and write the modified code back

### Example

**Before:**
```go
func (uc *AuthGenerateTokenUseCase) Execute(
    ctx context.Context,
    input GenerateTokenInput,
) (GenerateTokenOutput, error) {
    output := GenerateTokenOutput{}
    // ... rest of the function
}
```

**After:**
```go
func (uc *AuthGenerateTokenUseCase) Execute(
    ctx context.Context,
    input GenerateTokenInput,
) (GenerateTokenOutput, error) {
    ctx, span := trace.Span(ctx, "AuthGenerateTokenUseCase.Execute")
    defer span.End()
    
    output := GenerateTokenOutput{}
    // ... rest of the function
}
```

## Integration with Build Process

You can integrate tracegen into your build process:

### Option 1: Pre-build Step

Add to your Makefile:
```makefile
build: tracegen
	go build -o bin/myapp ./main.go
```

### Option 2: Go Generate

Add `//go:generate` comments to your files:
```go
//go:generate go run ../../cmd/tracegen/main.go --path ./usecase
package mypackage
```

Then run:
```bash
go generate ./...
```

## Best Practices

1. **Use Version Control**: Always commit your code before running tracegen so you can review and revert changes if needed
2. **Start with Dry Run**: Use `--dry-run` first to see what will be modified
3. **Be Specific**: Use `--pattern` to target specific functions instead of adding traces everywhere
4. **Review Changes**: Always review the generated code to ensure it meets your needs
5. **Exclude Test Files**: The tool automatically skips `*_test.go` files

## Limitations

- Only instruments functions with `context.Context` as the first parameter
- Assumes the trace package follows the OpenTelemetry pattern
- Does not handle complex AST modifications (e.g., if the function signature is split across multiple lines in unusual ways)

## Architecture

The tool consists of three main components:

- **`pkg/tracegen/tracegen.go`**: Core generator that injects traces
- **`pkg/tracegen/remover.go`**: Removes existing trace instrumentation
- **`cmd/tracegen.go`**: CLI interface using Cobra

## Examples

### Instrument All Use Cases
```bash
go run ./main.go tracegen --path ./internal/modules/identity/usecase --verbose
```

### Instrument Only Execute Methods
```bash
go run ./main.go tracegen --path ./internal/modules --pattern Execute --verbose
```

### Remove All Traces from a Module
```bash
go run ./main.go tracegen --path ./internal/modules/identity --remove --verbose
```

### Preview Changes for Multiple Paths
```bash
go run ./main.go tracegen \
  --path ./internal/modules/identity \
  --path ./internal/modules/monitor \
  --dry-run \
  --verbose
```

## Troubleshooting

### "Import not used" Error
If you see this error, it means the trace import was added but no traces were injected. This can happen if:
- The function already has tracing code
- The function doesn't match the pattern
- The function doesn't have `context.Context` as the first parameter

Solution: Remove the unused import manually or use `--remove` flag.

### Changes Not Applied
Make sure:
- You're not using `--dry-run` flag
- You have write permissions to the files
- The files are valid Go source files

### Syntax Errors After Generation
This usually means the original file had syntax issues. Make sure your code compiles before running tracegen:
```bash
go build ./...
```

## Contributing

To add new features or fix bugs:

1. Modify the code in `pkg/tracegen/`
2. Test your changes with `--dry-run` first
3. Run on real code and verify the output
4. Add documentation for new features

## License

This tool is part of the Pingo project and follows the same license.
