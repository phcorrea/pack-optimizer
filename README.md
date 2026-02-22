# Pack Optimizer

## What this does

- Optimizes pack combinations using these priorities:
1. ship only whole packs
2. ship the least number of items that still satisfies the order
3. if tied on shipped items, ship the fewest packs
- Exposes the logic via HTTP API.
- Serves a small HTML UI to interact with the API.
- Uses per-request pack sizes.

## Run

```bash
go run ./cmd/server
```

Environment variables:
- `PORT` (default: `8080`)

## API

- `GET /api/health`
- `POST /api/optimize`
```json
{"items_ordered":12001,"pack_sizes":[250,500,1000,2000,5000]}
```

Example:

```bash
curl -X POST http://localhost:8080/api/optimize \
  -H "Content-Type: application/json" \
  -d '{"items_ordered":12001,"pack_sizes":[250,500,1000,2000,5000]}'
```

## Tests

```bash
go test ./...
```

## Docker

Start backend (API + HTML UI):

```bash
docker compose up --build backend
```

Then open `http://localhost:8080`.

Run tests in Docker:

```bash
docker compose up --build test
```
