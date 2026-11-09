# ================= STEP 1: BUILDER =================
FROM golang:1.21-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Ambil kode TERBARU langsung dari GitHub (termasuk folder static dan port 7860)
# Cache Buster: v2
RUN git clone https://github.com/RaihanIDN/Resepku.git .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# ================= STEP 2: RUNNER =================
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./

EXPOSE 7860

CMD ["./main"]
