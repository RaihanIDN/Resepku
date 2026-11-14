# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Taruh di /tmp agar folder /app tetap kosong saat git clone dijalankan
ADD https://api.github.com/repos/RaihanIDN/Resepku/git/refs/heads/main /tmp/version.json

# Ambil semua kode & folder static langsung dari GitHub terbaru
RUN git clone https://github.com/RaihanIDN/Resepku.git .

# Bersihkan karakter spasi aneh dan pastikan port berjalan di 7860
RUN sed -i 's/\xc2\xa0/ /g' main.go
RUN sed -i 's/8080/7860/g' main.go

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Salin hasil build beserta folder static dan html
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

ENV PORT=7860
EXPOSE 7860

CMD ["./main"]
