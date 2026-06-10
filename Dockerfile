# ---- Stage 1: build the Vue frontend ----
# Cached unless frontend/ changes: package*.json first (deps layer), then sources.
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- Stage 2: build the Go backend (CGO required for go-sqlite3) ----
# Cached unless Go code changes: go.mod/go.sum first (deps layer), then ONLY the Go
# source trees — so frontend-only edits never trigger a Go rebuild, and vice versa.
FROM golang:1.22-bookworm AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
ENV CGO_ENABLED=1
RUN go build -o /out/chklst ./cmd/chklst

# ---- Stage 3: minimal runtime image ----
FROM debian:bookworm-slim
WORKDIR /app
# ca-certificates: outbound HTTPS to Jira / SMTP / Ollama. tzdata: correct local
# time for the daily-summary scheduler. curl: container healthcheck.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend /out/chklst /app/chklst
COPY --from=frontend /app/frontend/dist /app/frontend/dist

# Persisted at runtime via named volumes (see docker-compose.yml)
RUN mkdir -p /app/backups /app/data

EXPOSE 8000
ENV PORT=8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8000/health || exit 1
ENTRYPOINT ["/app/chklst"]
