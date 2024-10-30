# Gator

Gator is a CLI blog aggregator.

## Installation

This CLI requires Postgres and Go installed to run the program.

Make sure you have the latest [Go toolchain](https://golang.org/dl/) installed as well as a local Postgres database. You can then install `gator` with:

```bash
go install ...
```
## Config

Create a `.gatorconfig.json` file in your home directory with the following structure:

```json
{
  "db_url": "postgres://username:@localhost:5432/database?sslmode=disable"
}
```

Replace the values with your database connection string.

## Usage

Create a new user:

```bash
gator register <name>
```

Add a feed:
```bash
gator add feed <feed_name> <url>
```
Start the aggregator:

```bash
gator agg 30s
```

View the posts:

```bash
gator browse [limit]
```
There are a few other commands you'll need as well:

- `gator login <name>` - Log in as a user that already exists
- `gator users` - List all users
- `gator feeds` - List all feeds
- `gator follow <url>` - Follow a feed that already exists in the database
- `gator unfollow <url>` - Unfollow a feed that already exists in the database

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.