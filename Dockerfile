# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Salin seluruh source code (termasuk go.mod dan main.go) TERLEBIH DAHULU
COPY . .

# Sekarang jalankan tidy agar bisa membaca main.go dan mengunduh package (godotenv & pq)
RUN go mod tidy

# Build binary aplikasi Go
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest

# Tambahkan sertifikat keamanan standar
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Salin file binary utama yang sudah bersih dari tahap builder
COPY --from=builder /app/main .

# Salin folder static dan file HTML template
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

# Ekspos port 8080
EXPOSE 8080

# Eksekusi aplikasi
CMD ["./main"]
