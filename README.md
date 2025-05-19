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
```

## Example

```
$ godepbeat /path/to/your/project
Analyzing Go project at: /path/to/your/project
Found 90 dependencies in go.mod
Module: /path/to/your/project
Go Version: 1.24.2
Direct Dependencies: 30

Archived (Dead) Go Dependencies:
github.com/avast/retry-go
          Last Published: Oct 13, 2020
github.com/golang/mock
          Last Published: Jun 11, 2021
github.com/pkg/errors
          Last Published: Jan 14, 2020
github.com/opentracing/opentracing-go
          Last Published: Jul 1, 2020
github.com/patrickmn/go-cache
          Last Published: Jul 22, 2017

Summary:
- Total Dependencies: 90
- Direct Dependencies: 30
- Unmaintained Dependencies: 5
```

## License

GNU General Public License v3.0