# --- Estágio 1: Compilação ---
FROM golang:alpine AS builder
RUN apk add --no-cache alpine-sdk
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o odonto-flow-api ./cmd/api/main.go

# --- Estágio 2: Execução Segura ---
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/odonto-flow-api .
COPY private.pem public.pem ./

# Segurança: Executa como usuário não-root
RUN adduser -D -u 10001 odontouser
USER odontouser

EXPOSE 8080
CMD ["./odonto-flow-api"]