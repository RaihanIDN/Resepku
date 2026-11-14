# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Paksa Docker selalu cek commit terbaru di GitHub (Anti-Cache)
ADD https://api.github.com/repos/RaihanIDN/Resepku/git/refs/heads/main version.json

# Ambil semua kode & folder static langsung dari GitHub terbaru
RUN git clone https://github.com/RaihanIDN/Resepku.git .

# Pastikan port di main.go jalan di 7860 dan bersih dari karakter spasi aneh
RUN sed -i 's/\xc2\xa0/ /g' main.go
RUN sed -i 's/8080/7860/g' main.go

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Salin semua hasil build beserta folder static dan html sekaligus
COPY --from=builder /app /app

ENV PORT=7860
EXPOSE 7860

CMD ["./main"]
