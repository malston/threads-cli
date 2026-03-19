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

### Option A: OAuth (full setup)

1. Create a Threads app at https://developers.facebook.com
2. Add `https://localhost:8443/callback` as a valid redirect URI
3. Set environment variables:
   ```bash
   export THREADS_APP_ID=your_app_id
   export THREADS_APP_SECRET=your_app_secret
   ```
4. Authenticate:
   ```bash
   threads auth login --port 8443
   ```

### Option B: Manual token

If you already have an access token (e.g., from the Graph API Explorer):

```bash
# Pipe to avoid shell history exposure
echo "$TOKEN" | threads auth save-token

# Or pass directly (visible in shell history)
threads auth save-token YOUR_TOKEN
```

### Option C: Environment variable

```bash
export THREADS_ACCESS_TOKEN=your_token
```

## Usage

### Authentication

```bash
threads auth login --port 8443     # OAuth flow (opens browser)
threads auth save-token             # Save token from stdin
threads auth token                  # Show token status and expiry
threads auth token --refresh        # Refresh if expiring soon
```

### Posts

```bash
threads post create "Hello world"                                    # Text post
threads post create "Check this out" --image https://example.com/a.jpg   # Image post
threads post create "Watch this" --video https://example.com/v.mp4   # Video post
threads post create "Photos" --carousel https://a.jpg,https://b.jpg  # Carousel
threads post get 12345                                               # Get post by ID
threads post list                                                    # List your posts
threads post list --limit 10 --after CURSOR                          # With pagination
threads post delete 12345                                            # Delete a post
threads post repost 12345                                            # Repost
```

### Search

```bash
threads search "trending topic"                          # Keyword search
threads search "aithreads" --mode tag                    # Tag search
threads search "topic" --author someuser --sort recent   # By author, most recent
threads search "photo" --media-type image                # Filter by media type
threads search "event" --since 1700000000 --until 1700086400  # Date range
```

Search flags:

| Flag           | Values                   | Description              |
| -------------- | ------------------------ | ------------------------ |
| `--author`     | username                 | Filter by author         |
| `--sort`       | `top`, `recent`          | Sort order               |
| `--mode`       | `keyword`, `tag`         | Search mode              |
| `--media-type` | `text`, `image`, `video` | Filter by media type     |
| `--since`      | unix timestamp           | Results after this time  |
| `--until`      | unix timestamp           | Results before this time |

### Replies

```bash
threads reply list 12345                    # List replies to a post
threads reply create 12345 "Great post!"    # Reply to a post
threads reply hide 67890                    # Hide a reply
threads reply unhide 67890                  # Unhide a reply
```

### Profile

```bash
threads profile              # Your profile
threads profile username     # Look up a public profile
```

### Insights

```bash
threads insights post 12345                              # Post metrics
threads insights post 12345 --metrics views,likes        # Specific metrics
threads insights user                                    # User metrics
threads insights user --since 1700000000 --until 1700086400  # Time range
```

### Rate Limits

```bash
threads limits    # Show current rate limit usage
```

## Global Flags

| Flag       | Default | Description                               |
| ---------- | ------- | ----------------------------------------- |
| `--format` | `text`  | Output format: `json`, `text`, or `table` |
| `--token`  |         | Override access token                     |

## Token Resolution

The CLI resolves the access token in this order:

1. `--token` flag
2. `THREADS_ACCESS_TOKEN` environment variable
3. Stored credentials at `~/.threads-cli/credentials.json`

## Configuration Files

| Path                              | Purpose                                                     |
| --------------------------------- | ----------------------------------------------------------- |
| `~/.threads-cli/credentials.json` | OAuth tokens (created by `auth login` or `auth save-token`) |
| `~/.threads-cli/config.json`      | App config (optional, env vars preferred)                   |

## Development

```bash
make build            # Build binary
make test             # Run unit tests
make test-integration # Run integration tests (requires THREADS_ACCESS_TOKEN)
make lint             # Run linter
make install          # Install to GOPATH/bin
make clean            # Remove binary
```

## Privacy

See [Privacy Policy](https://www.markalston.net/threads-cli/).
