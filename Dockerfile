# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Salin seluruh file dari repositori
COPY . .

# Rapihkan dependensi
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Salin file binary
COPY --from=builder /app/main .

# Salin folder static dan HTML
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

# Ekspos port wajib Hugging Face
EXPOSE 7860

CMD ["./main"]
