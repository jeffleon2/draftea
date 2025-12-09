# ============================================
# BUILD STAGE (alpine + static binary)
# ============================================
FROM golang:1.25.3-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copiar go.mod antes para cache
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

# ============================================
# RUNTIME STAGE (scratch — ultra minimal)
# ============================================
FROM scratch

# Certificados para HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

# Copiar el binario estático
COPY --from=builder /app/main .

# Exponer puerto si deseas
EXPOSE 8080

ENTRYPOINT ["/app/main"]
