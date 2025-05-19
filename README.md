# godepbeat

A Go tool to check the health of your Go module dependencies.

## Features

- Analyzes `go.mod` files to extract dependency information
- Checks if dependencies are maintained or archived
- Uses concurrent requests (10 at a time) to efficiently check multiple dependencies
- Provides both text and JSON output formats

## Usage

```
godepbeat [options] <path-to-go-project>

Options:
  -json      Output in JSON format
  -quiet     Suppress progress output
```

## Example

```
$ godepbeat /path/to/your/project
Analyzing Go project at: /path/to/your/project
Found 25 dependencies in go.mod
Module: github.com/your/project
Go Version: 1.18
Analyzing: github.com/some/dependency                      [Active (Last published: Jan 25, 2023)]
Analyzing: github.com/archived/dependency                  [ARCHIVED (Last published: Mar 5, 2020)]
...

Summary:
- Total Dependencies: 25
- Direct Dependencies: 18
- Unmaintained Dependencies: 3
```

## License

GNU General Public License v3.0