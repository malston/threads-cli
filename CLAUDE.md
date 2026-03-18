# saved-threads

CLI wrapper for the Meta Threads API.

## Build and Test

```bash
make build    # Build binary
make test     # Run unit tests
make lint     # Run linter
make install  # Install to GOPATH
make test-integration  # Run integration tests (requires THREADS_ACCESS_TOKEN)
```

## Architecture

- `cmd/` - Cobra CLI commands
- `internal/api/` - HTTP client, error handling, pagination
- `internal/auth/` - OAuth flow and token management
- `internal/config/` - App config and credential storage
- `internal/threads/` - Domain operations (posts, replies, search, profile, insights)
- `internal/output/` - Output formatting (JSON, text, table)

## Environment Variables

- `THREADS_APP_ID` - Threads app ID (required for auth)
- `THREADS_APP_SECRET` - Threads app secret (required for auth)
- `THREADS_ACCESS_TOKEN` - Access token (overrides stored credentials)
