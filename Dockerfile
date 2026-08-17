# --- build  --------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api


RUN mkdir -p /out/data/media /out/data/tmp && chown -R 65532:65532 /out/data

# --- runtime -----------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/api /api
COPY --from=builder --chown=65532:65532 /out/data /data

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/api"]
