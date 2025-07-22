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
    
    COPY --from=builder /app/bigdataimporter .
    
    EXPOSE 8080
    
    CMD ["./bigdataimporter"]
    