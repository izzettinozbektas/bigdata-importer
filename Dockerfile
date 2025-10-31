# ---------- STAGE 1: Build ----------
    FROM golang:1.21-alpine AS builder

    WORKDIR /app
    
    COPY . .
    
    RUN [ ! -f go.mod ] && go mod init bigdataimporter || echo "mod var"
    
    RUN go mod tidy
    
    RUN go build -o bigdataimporter ./cmd/main.go
    
    
    # ---------- STAGE 2: Runtime ----------
    FROM alpine:latest
    
    WORKDIR /app
    
    # Binary'yi al
    COPY --from=builder /app/bigdataimporter .
    
    # Config dosyasını da kopyala
    COPY config.yaml /app/config.yaml
    
    EXPOSE 8080
    
    CMD ["./bigdataimporter"]
    