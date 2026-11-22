# syntax=docker/dockerfile:1
# Atalaia: one static binary with the web app inside, in a minimal non-root image.
# Build from the repository root:  docker build -f deploy/Dockerfile -t atalaia .

FROM node:24-alpine@sha256:116b800c3754b440ee0aa3ed2f20f5bf3d99a554bfd9adeea09ea8017f6e85f5 AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npx vite build --outDir /out --emptyOutDir

FROM golang:1-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /out internal/web/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/EttoreCSenatore/atalaia/internal/version.Version=${VERSION}" -o /atalaia ./cmd/atalaia \
 && mkdir /data

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /atalaia /atalaia
COPY --from=build --chown=65532:65532 /data /data
ENV ATALAIA_DB=/data/atalaia.db
VOLUME /data
USER 65532:65532
# Needs the host network to see the LAN (see docker-compose.yml).
ENTRYPOINT ["/atalaia"]
CMD ["serve", "--listen", "127.0.0.1:8080"]
