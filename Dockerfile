# Etapa de build
FROM golang:1.22 AS builder

# Definir o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copiar os arquivos go.mod e go.sum
COPY go.mod go.sum ./

# Baixar as dependências
RUN go mod download

# Copiar o código-fonte para o contêiner
COPY . .

# Compilar a aplicação com CGO desabilitado para uma compilação estática
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Etapa final
FROM alpine:latest

# Definir o diretório de trabalho dentro do contêiner
WORKDIR /root/

# Instalar o pacote ca-certificates se sua aplicação fizer requisições HTTPS
RUN apk --no-cache add ca-certificates

# Copiar o binário compilado da etapa de build
COPY --from=builder /app/main .

# Expor a porta que a aplicação utiliza (ajuste conforme necessário)
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
