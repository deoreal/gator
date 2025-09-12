# Gator 🐊

Gator is a command-line RSS feed aggregator written in Go. It allows users to manage RSS feeds, follow feeds, and browse aggregated posts from their subscribed feeds.

## Features

- **User Management**: Register users and manage login sessions
- **Feed Management**: Add RSS feeds and manage subscriptions
- **Feed Following**: Follow/unfollow RSS feeds
- **Content Aggregation**: Automatically scrape and aggregate posts from followed feeds
- **Post Browsing**: Browse posts from your followed feeds with customizable limits
- **Database Storage**: Persistent storage using PostgreSQL

## Prerequisites

- Go 1.25.0 or later
- PostgreSQL database
- Database connection configured in `.gatorconfig.json`

## Installation

### Option 1: Install directly with Go

```bash
go install github.com/deoreal/gator@latest
```

This will install the `gator` binary to your `$GOPATH/bin` directory (or `$HOME/go/bin` if `GOPATH` is not set). Make sure this directory is in your `PATH`.

### Option 2: Build from source

1. Clone the repository:
```bash
git clone https://github.com/deoreal/gator.git
cd gator
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o gator
```

## Configuration

Create a `.gatorconfig.json` file in your home directory with your database connection string:

```json
{
  "db_url": "postgres://username:password@localhost/gator?sslmode=disable"
}
```

## Usage

### User Management

**Register a new user:**
```bash
./gator register <username>
```

**Login as an existing user:**
```bash
./gator login <username>
```

**List all users:**
```bash
./gator users
```

### Feed Management

**Add a new RSS feed:**
```bash
./gator addfeed <feed_name> <feed_url>
```

**List all feeds:**
```bash
./gator feeds
```

**Follow an existing feed:**
```bash
./gator follow <feed_url>
```

**Unfollow a feed:**
```bash
./gator unfollow <feed_url>
```

**List feeds you're following:**
```bash
./gator following
```

### Content Aggregation

**Start continuous feed aggregation:**
```bash
./gator agg [interval]
```
- `interval`: Time between scraping cycles (default: "10s")
- Examples: "30s", "5m", "1h"

**Perform one-time scrape:**
```bash
./gator scrape
```

### Browsing Posts

**Browse posts from followed feeds:**
```bash
./gator browse [limit]
```
- `limit`: Number of posts to display (default: 2)

### Database Management

**Reset the database (caution: removes all data):**
```bash
./gator reset
```

## Examples

```bash
# Register and login
./gator register alice
./gator login alice

# Add and follow a feed
./gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"
./gator follow "https://blog.boot.dev/index.xml"

# Start aggregating feeds every 30 seconds
./gator agg 30s

# Browse latest 5 posts
./gator browse 5
```

## Architecture

- **Database Layer**: Uses SQLC for type-safe SQL queries with PostgreSQL
- **Configuration**: JSON-based configuration management
- **RSS Parsing**: Native XML parsing for RSS feed content
- **CLI Interface**: Command-based interface with middleware for authentication
- **Concurrent Scraping**: Ticker-based feed scraping with configurable intervals

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/lib/pq` - PostgreSQL driver
- Built-in Go libraries for HTTP, XML parsing, and database operations

## License

This project is part of the Boot.dev Go course curriculum.