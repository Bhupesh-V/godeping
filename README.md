# godeping

_Pronounced: "Go-dep-ping"_

> **A Go tool to check whether you Go project dependencies are maintained or not.**

## Installation

```
go install github.com/Bhupesh-V/godeping@latest
```

## Use-cases

1. **Tech Debt, Refactoring & Cleanup**
   - When inheriting an unfamiliar Go project, get a sense of technical debt in terms of unmaintained dependencies.
2. **Security Audits & Compliance**
   - Flag deps that might not receive security patches anymore.
3. **Open-Source Spirit**
   - Help (sponsor) maintainers to keep the unmaintained libraries alive.
   - Fork the ones that cannont be helped. Take charge on giving back to the community.

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
```

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
