#!/bin/bash

# Script para rodar migrations no EC2
set -e

echo "🗄️ Executando migrations..."

# Configurações do banco (ajuste conforme necessário)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-planning}

DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# Instalar Goose se não existir
if ! command -v goose &> /dev/null; then
    echo "Instalando Goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Executar migrations
goose -dir db/migrations postgres "$DATABASE_URL" up

echo "✅ Migrations executadas com sucesso!"