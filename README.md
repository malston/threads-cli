# threads-cli

CLI wrapper for the Meta Threads API.

## Installation

```bash
go install github.com/malston/threads-cli@latest
```

Or build from source:

```bash
git clone https://github.com/malston/threads-cli.git
cd threads-cli
make install
```

## Setup

1. Create a Threads app at https://developers.facebook.com
2. Set environment variables:
   ```bash
   export THREADS_APP_ID=your_app_id
   export THREADS_APP_SECRET=your_app_secret
   ```
3. Authenticate:
   ```bash
   threads auth login
   ```

## Usage

```bash
# Posts
threads post create "Hello world"
threads post create "Check this out" --image https://example.com/photo.jpg
threads post list --limit 10
threads post get 12345
threads post delete 12345
threads post repost 12345

# Replies
threads reply list 12345
threads reply create 12345 "Great post!"
threads reply hide 67890
threads reply unhide 67890

# Search
threads search "trending topic"

# Profile
threads profile            # Your profile
threads profile username   # Public profile lookup

# Insights
threads insights post 12345
threads insights user --since 1700000000 --until 1700086400

# Rate limits
threads limits
```

## Global Flags

| Flag       | Default | Description                               |
| ---------- | ------- | ----------------------------------------- |
| `--format` | `text`  | Output format: `json`, `text`, or `table` |
| `--token`  |         | Override access token                     |

Token resolution: `--token` flag > `THREADS_ACCESS_TOKEN` env var > stored credentials.

## Development

```bash
make build            # Build binary
make test             # Run unit tests
make test-integration # Run integration tests (requires THREADS_ACCESS_TOKEN)
make lint             # Run linter
```
