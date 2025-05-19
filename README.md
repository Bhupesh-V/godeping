# godeping

A Go tool to check the health of your Go module dependencies.

## Usage

```
godeping [options] <path-to-go-project>

Options:
  -quiet
        Suppress progress output
```

## Example

```
$ godeping -quiet /path/to/your/project
Analyzing Go project at: /path/to/your/project
Found 90 dependencies in go.mod
Module: /path/to/your/project
Go Version: 1.24.2
Direct Dependencies: 30

Archived (Dead) Direct Dependencies:
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