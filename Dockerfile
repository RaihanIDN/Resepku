FROM golang:1.21-alpine AS builder

RUN apk add --no-cache curl unzip

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN curl -L -o repo.zip https://github.com/RaihanIDN/Resepku/archive/refs/heads/main.zip \
    && unzip repo.zip \
    && mv Resepku-main/static ./static \
    && rm -rf repo.zip Resepku-main

RUN go build -o main .

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/static ./static
COPY --from=builder /app/*.html ./


EXPOSE 7860

CMD ["./main"]
