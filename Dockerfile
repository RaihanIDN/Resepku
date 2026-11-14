# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Ambil langsung file terbaru yang sudah dikirim otomatis oleh GitHub Actions dari GitHub
COPY . .

# Otomatis pastikan port berjalan di 7860 milik Hugging Face
RUN sed -i 's/8080/7860/g' main.go

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

ENV PORT=7860
EXPOSE 7860

CMD ["./main"]
