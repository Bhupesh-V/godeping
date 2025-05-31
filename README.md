# godeping

_Pronounced: "Go deping". Technically, "Go dep ping"_

> **A Go tool to check whether you Go project dependencies are maintained or not.**

[![Go Build & Test](https://github.com/Bhupesh-V/godeping/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/Bhupesh-V/godeping/actions/workflows/main.yml)

## Installation

```
go install github.com/Bhupesh-V/godeping@latest
```

## Use-cases

1. **Tech Debt, Refactoring & Cleanup**
   - When inheriting an unfamiliar Go project, get a sense of technical debt in terms of dependencies.
2. **Security Audits & Compliance**
   - Flag deps that might not receive security patches anymore.
3. **Open-Source Spirit**
   - Help (sponsor) maintainers to keep the unmaintained libraries alive.
   - Fork the ones that cannot be helped. Take charge on giving back to the community 🏃🏼‍♂️.

## Judgement Criteria

`godeping` relies on the Go Infrastructure to determine whether a dependency is archived or not. Namely `pkg.go.dev` which itself is powered by [`index.golang.org`](https://index.golang.org/).

- As of today an API for [`pkg.go.dev` is still not available](https://github.com/golang/go/issues/36785).
- By default, `godeping` considers a module unmaintained if it hasn't been updated in 2 years. This threshold can be customized using the `-since` flag.

## Usage

Quick start (assuming you are in the root of your Go project):

```bash
godeping -quiet .
```

All options:

```bash
godeping [options] <path-to-go-project>

Options:
  -json
        Output in JSON format
  -quiet
        Suppress progress output
  -since string
        Consider dependencies as unmaintained if not updated since this duration (e.g. 1y, 6m, 2y3m) (default "2y")
```

### Duration Format for -since

The `-since` flag accepts durations in several formats:

- Years: `2y` (2 years)
- Months: `6m` (6 months)
- Years and months: `1y6m` (1 year and 6 months)
- Standard Go duration format: `720h` (30 days)

## Example

### Text (Default) Mode

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

### Using Custom Duration

```
$ godeping -quiet -since 6m /path/to/your/project
```

This checks for dependencies that haven't been updated in the last 6 months.

### JSON Mode

Use the `-json` flag to output the results in JSON format. JSON mode enables `-quiet` by default.

```
godeping -json /path/to/your/project
```

```json
{
  "module": "your/main/module",
  "goVersion": "1.24.2",
  "totalDependencies": 90,
  "directDependencies": 30,
  "deadDirectDependencies": [
    {
      "module_path": "github.com/golang/mock",
      "last_published": "2021-06-11T00:00:00Z"
    },
    {
      "module_path": "github.com/avast/retry-go",
      "last_published": "2020-10-13T00:00:00Z"
    },
    {
      "module_path": "github.com/patrickmn/go-cache",
      "last_published": "2017-07-22T00:00:00Z"
    },
    {
      "module_path": "github.com/pkg/errors",
      "last_published": "2020-01-14T00:00:00Z"
    },
    {
      "module_path": "github.com/opentracing/opentracing-go",
      "last_published": "2020-07-01T00:00:00Z"
    }
  ]
}
```

## Alternatives

If you fancy freedom.

```bash
go mod edit -json \
  | jq -r '.Require[].Path' \
  | grep github.com \
  | while read -r path; do
    # Strip github.com/
    repo=$(echo "$path" | sed -E 's|^github.com/||')
    # Remove /v2, /v3, etc. at the end
    clean_repo=$(echo "$repo" | sed -E 's|/v[0-9]+$||')
    echo -n "$clean_repo: "
    gh repo view "$clean_repo" --json isArchived --jq '.isArchived' 2>/dev/null || echo "not found"
done
```

## License

This project is licensed under the GNU General Public License v3.0 License - see the [LICENSE](LICENSE) file for details.
