FROM golang:1.26.0-alpine AS base
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

#####
FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

#####
FROM base AS test
CMD ["go", "test", "-v", "-count=1", "./..."]

#####
FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app

COPY --from=build /out/server /app/server

ENV PORT=8080
EXPOSE ${PORT}

ENTRYPOINT ["/app/server"]
