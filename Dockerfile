# ---------- Frontend (web UI — Vite/React) ----------
FROM node:20-alpine AS web
# pnpm 8 (lockfileVersion 6.0 do projeto) — o pnpm 11 exigiria Node >= 22
RUN corepack enable && corepack prepare pnpm@8 --activate
WORKDIR /src/app
COPY app/wails.json wails.json
COPY app/frontend/package.json app/frontend/pnpm-lock.yaml frontend/
WORKDIR /src/app/frontend
RUN pnpm install --frozen-lockfile
COPY app/frontend/ ./
RUN pnpm build

# ---------- Backend (GUI server — Go) ----------
FROM golang:1.21-alpine AS builder
WORKDIR /src

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/palworld-ds-gui-server .

# ---------- Runtime ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /data
COPY --from=builder /out/palworld-ds-gui-server /usr/local/bin/palworld-ds-gui-server
COPY --from=web /src/app/frontend/dist /usr/local/share/palworld-ds-gui-web
EXPOSE 21577
# O binário roda a partir de /data para que logs/settings/steamcmd fiquem no volume persistente.
# /data/server pré-existente evita que o GUI baixe o servidor dedicado (o jogo já roda em container no swarm).
ENTRYPOINT ["/bin/sh", "-c", "mkdir -p /data/server && cp /usr/local/bin/palworld-ds-gui-server /data/palworld-ds-gui-server && exec /data/palworld-ds-gui-server"]
