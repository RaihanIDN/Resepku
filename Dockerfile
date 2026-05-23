# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

# Install curl dan unzip untuk mengambil aset dari GitHub publik
RUN apk add --no-cache curl unzip

WORKDIR /app

# Copy berkas go.mod terlebih dahulu
COPY go.mod ./

# Ambil folder static & file HTML segar langsung dari GitHub publik agar sinkron
RUN curl -L -o repo.zip https://github.com/RaihanIDN/Resepku/archive/refs/heads/main.zip \
    && unzip repo.zip \
    && mv Resepku-main/static ./static \
    && cp Resepku-main/*.html ./ \
    && rm -rf repo.zip Resepku-main

# Copy source code utama (main.go, dll)
COPY . .

# Trik Jitu: Ambil library Supabase & .env sekaligus membuat berkas go.sum otomatis bray
RUN go mod tidy

# Build binary aplikasi Go menjadi berkas mandiri bernama 'main'
RUN go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Ambil file binary utama yang sudah bersih
COPY --from=builder /app/main .

# Ambil folder static dan semua file HTML template hasil ekstrak tadi
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

# Port standar Hugging Face Spaces
EXPOSE 7860

# Jalankan aplikasinya bray
CMD ["./main"]
