COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

govet:
	go vet ./...

security-check:
	govulncheck ./...

test:
	go test -race ./... -coverpkg=./... -coverprofile=$(COVERAGE_FILE)

coverage:
	go tool cover -html $(COVERAGE_FILE) -o $(COVERAGE_HTML)

get-coverage: ## Get overall project test coverage
	go run scripts/get_overall_coverage.go