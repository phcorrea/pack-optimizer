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

- `GET /api/health`
- `GET /api/pack-sizes`
- `PUT /api/pack-sizes`
- `POST /api/optimize`
```json
{"items_ordered":12001,"pack_sizes":[250,500,1000,2000,5000]}
```

`GET /api/pack-sizes` response example:

```json
{"pack_sizes":[5000,2000,1000,500,250]}
```

Update pack sizes:

```bash
curl -X PUT http://localhost:8080/api/pack-sizes \
  -H "Content-Type: application/json" \
  -d '{"pack_sizes":[250,500,1000,2000,5000]}'
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
