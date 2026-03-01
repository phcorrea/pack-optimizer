# Pack Optimizer

## What this does

- Optimizes pack combinations using these priorities:
1. ship only whole packs
2. ship the least number of items that still satisfies the order
3. if tied on shipped items, ship the fewest packs
- Exposes the logic via HTTP API.
- Serves a small HTML UI to interact with the API.
- Supports reading/updating default pack sizes via API.

## Run

```bash
go run ./cmd/server
```

Environment variables:
- `PORT` (default: `8080`)

## API

### `POST /api/optimize`

Response example:
```json
{"items_ordered":12001}
```

Example:

```bash
curl -X POST http://localhost:8080/api/optimize \
  -H "Content-Type: application/json" \
  -d '{"items_ordered":12001}'
```

### `GET /api/pack-sizes`

Response example:

```json
{"pack_sizes":[5000,2000,1000,500,250]}
```

### `PUT /api/pack-sizes`

Response example:

```json
{"pack_sizes":[5000,2000,1000,500,250]}
```

Example:
```bash
curl -X PUT http://localhost:8080/api/pack-sizes \
  -H "Content-Type: application/json" \
  -d '{"pack_sizes":[250,500,1000,2000,5000]}'
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
