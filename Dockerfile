# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Salin file konfigurasi module Go terlebih dahulu untuk cache dependensi
COPY go.mod ./
# Jika kamu sudah punya go.sum di repositori, aktifkan baris di bawah ini:
# COPY go.sum ./

# Rapihkan dan unduh dependensi yang dibutuhkan
RUN go mod tidy

# Salin seluruh source code, folder static, dan file HTML ke dalam container
COPY . .

# Build binary aplikasi Go (CGO dinonaktifkan agar stabil di Alpine)
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

# Ekspos port 8080 sesuai konfigurasi Back4App dan kode main.go
EXPOSE 8080

# Eksekusi aplikasi
CMD ["./main"]
