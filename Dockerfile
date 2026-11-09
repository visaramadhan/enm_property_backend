# syntax=docker/dockerfile:1
# =============================================================================
# ENMProperty Backend (Go + Gin)
#
# Build multi-tahap:
#   1. builder  -> kompilasi biner statis (CGO dinonaktifkan)
#   2. runtime  -> image Alpine ramping yang hanya memuat biner + CA + tzdata
#
# Berkas migrasi SQL di-embed ke biner (go:embed di package migrations),
# sehingga tidak perlu disalin terpisah ke image akhir.
# =============================================================================

# ---------------------------------------------------------------- Tahap build
FROM golang:1.27-alpine AS builder

WORKDIR /src

# Salin manifes dependensi lebih dulu agar lapisan unduhan dapat di-cache
# selama go.mod/go.sum tidak berubah.
COPY go.mod go.sum ./
RUN go mod download

# Salin sisa kode (termasuk migrations/*.sql) lalu bangun biner.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/enmproperty-api .

# ------------------------------------------------------------- Tahap runtime
FROM alpine:3.20

# ca-certificates : verifikasi TLS ke Firebase/Midtrans/Google Cloud.
# tzdata          : memenuhi APP_TIMEZONE=Asia/Jakarta (time.LoadLocation).
# Busybox sudah menyediakan wget untuk HEALTHCHECK.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/enmproperty-api ./enmproperty-api

# /app/storage : media lokal (STORAGE_LOCAL_DIR), diisi volume pada compose.
# /app/secrets : service account Firebase (di-mount read-only), opsional.
RUN mkdir -p /app/storage /app/secrets && chown -R app:app /app

# Jalankan sebagai pengguna non-root.
USER app

ENV HTTP_PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -O- "http://127.0.0.1:${HTTP_PORT}/api/v1/health" >/dev/null 2>&1 || exit 1

ENTRYPOINT ["./enmproperty-api"]
