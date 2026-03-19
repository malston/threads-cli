---
name: saved-threads
description: Use the `threads` CLI to interact with the Meta Threads API -- post, search, manage replies, view insights
---

# saved-threads CLI

A CLI wrapper for the Meta Threads API. Binary name: `threads`.

## Setup

### Prerequisites

1. Create a Threads app at https://developers.facebook.com
2. Set environment variables:
   - `THREADS_APP_ID` - Your Threads app ID
   - `THREADS_APP_SECRET` - Your Threads app secret
3. Run `threads auth login` to authenticate via OAuth

### Build

```bash
cd /Users/markalston/code/saved-threads
make build    # produces ./threads binary
make install  # installs to GOPATH/bin
```

## Commands

### Authentication

```bash
threads auth login              # Start OAuth flow (opens browser URL)
threads auth token              # Show token status and expiry
threads auth token --refresh    # Refresh token if expiring soon
```

### Posts

```bash
threads post create "Hello world"                          # Text post
threads post create "Check this out" --image https://...   # Image post
threads post create "My photos" --carousel https://a.jpg,https://b.jpg  # Carousel
threads post get 12345                                     # Get post by ID
threads post list                                          # List your posts
threads post list --limit 10 --after CURSOR                # With pagination
threads post delete 12345                                  # Delete a post
threads post repost 12345                                  # Repost
```

### Replies

```bash
threads reply list 12345                    # List replies to post
threads reply create 12345 "Great post!"    # Reply to a post
threads reply hide 67890                    # Hide a reply
threads reply unhide 67890                  # Unhide a reply
```

### Search

```bash
threads search "trending topic"                     # Search posts
threads search "query" --limit 50 --after CURSOR    # With pagination
```

### Profile

```bash
threads profile              # Show your profile
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

- `--format json|text|table` - Output format (default: text)
- `--token TOKEN` - Override access token (also reads `THREADS_ACCESS_TOKEN` env var)

## Token Resolution Order

1. `--token` flag
2. `THREADS_ACCESS_TOKEN` environment variable
3. Stored credentials from `~/.saved-threads/credentials.json`

## Config Files

- `~/.saved-threads/credentials.json` - Stored OAuth tokens
- `~/.saved-threads/config.json` - App configuration (optional, env vars preferred)
