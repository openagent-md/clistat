## Commands

```bash
make test          # Run tests
make test-race     # Tests with race detection
make lint          # golangci-lint + shellcheck
make fmt           # gofumpt
```

Single test: `go test -v -run TestName ./...`

## Conventions

- Use `afero.Fs` for filesystem operations (enables test mocking)

## Test Environment Variables

```bash
CLISTAT_IS_CONTAINERIZED=yes
CLISTAT_HAS_MEMORY_LIMIT=yes
CLISTAT_HAS_CPU_LIMIT=yes
CLISTAT_IS_CGROUPV2=yes
```
