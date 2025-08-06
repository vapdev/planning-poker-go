# Planning Poker Go

Um sistema de Planning Poker desenvolvido em Go com interface web em Vue.js/Nuxt.js para auxiliar equipes ágeis na estimativa de tarefas.

## 🏗️ Arquitetura

- **Backend**: Go com WebSockets para comunicação em tempo real
- **Frontend**: Nuxt.js/Vue.js (projeto separado)
- **Banco de Dados**: PostgreSQL
- **Migrações**: Goose

## 📋 Pré-requisitos

- Go 1.22+
- PostgreSQL 13+
- Docker & Docker Compose (opcional)
- Goose (para migrações)

## 🚀 Deploy Options

### 1. Deploy Local (Desenvolvimento)

#### Passo 1: Clonar e configurar
```bash
git clone <seu-repositorio>
cd planning-poker-go

# Criar arquivo de configuração
cp .env.example .env
```

#### Passo 2: Configurar PostgreSQL

**Opção A: PostgreSQL Local**
```bash
# Instalar PostgreSQL (macOS)
brew install postgresql
brew services start postgresql

# Criar banco
createdb planning-poker
```

**Opção B: PostgreSQL com Docker**
```bash
docker-compose up -d db
```

#### Passo 3: Instalar Goose e executar migrações
```bash
# Instalar Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Adicionar ao PATH (adicione ao seu ~/.zshrc ou ~/.bashrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Executar migrações
goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5432/planning" up
```

#### Passo 4: Executar aplicação
```bash
# Instalar dependências
go mod download

# Executar
go run .
```

A aplicação estará disponível em `http://localhost:8080`

### 2. Deploy com Docker

#### Passo 1: Build completo
```bash
# Subir toda a stack
docker-compose up -d

# Verificar se os containers estão rodando
docker-compose ps
```

#### Passo 2: Executar migrações
```bash
# Aguardar o banco estar pronto e executar migrações
sleep 10
goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5432/planning" up
```

#### Comandos úteis Docker
```bash
# Ver logs
docker-compose logs -f app
docker-compose logs -f db

# Parar containers
docker-compose down

# Rebuild da aplicação
docker-compose up -d --build app
```

### 3. Deploy em Cloud (AWS/DigitalOcean/Linode)

#### Usando Docker em qualquer VPS
```bash
# Em qualquer servidor com Docker
git clone <seu-repositorio>
cd planning-poker-go

# Configurar variáveis para produção
cp .env.example .env
# Editar com suas configurações de produção

# Deploy
docker-compose up -d

# Executar migrações
sleep 10
goose -dir db/migrations postgres $DATABASE_URL up
```

#### AWS EC2 / DigitalOcean Droplet
```bash
# 1. Criar instância (Ubuntu 20.04+)
# 2. Conectar via SSH
# 3. Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker $USER

# 4. Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 5. Deploy do projeto
git clone <seu-repositorio>
cd planning-poker-go
docker-compose up -d
```

### 4. Deploy em VPS/Servidor

#### Passo 1: Preparar servidor
```bash
# Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### Passo 2: Deploy
```bash
# Clonar repositório
git clone <seu-repositorio>
cd planning-poker-go

# Configurar variáveis de ambiente para produção
cp .env.example .env
# Editar .env com configurações de produção

# Subir aplicação
docker-compose up -d

# Executar migrações
docker exec -it planning-poker-go-app-1 /bin/sh
# Dentro do container, executar migrações se necessário
```

#### Passo 3: Configurar Nginx (opcional)
```nginx
server {
    listen 80;
    server_name seu-dominio.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 Configurações

### Variáveis de Ambiente

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `DATABASE_URL` | Connection string do PostgreSQL | `postgres://user:pass@host:5432/dbname` |
| `PORT` | Porta da aplicação | `8080` |

### Configuração do Banco

O arquivo `docker-compose.yml` já inclui as configurações do PostgreSQL:
- **Usuário**: `postgres`
- **Senha**: `postgres`
- **Banco**: `planning`
- **Porta**: `5432`

## 🗃️ Migrações

### Comandos úteis do Goose

```bash
# Ver status das migrações
goose -dir db/migrations postgres $DATABASE_URL status

# Executar migrações
goose -dir db/migrations postgres $DATABASE_URL up

# Reverter última migração
goose -dir db/migrations postgres $DATABASE_URL down

# Criar nova migração
goose -dir db/migrations create nome_da_migracao sql
```

### Estrutura do Banco

As migrações criam as seguintes tabelas:
- `users` - Usuários do sistema
- `rooms` - Salas de planning poker
- `room_users` - Relacionamento usuários/salas
- `issues` - Issues para votação
- `votes` - Votos dos usuários

## 🔍 Troubleshooting

### Problemas comuns

**1. Goose não encontrado**
```bash
# Adicionar ao PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

**2. Erro de conexão com banco**
```bash
# Verificar se PostgreSQL está rodando
docker-compose ps
# ou
brew services list | grep postgresql
```

**3. Porta já em uso**
```bash
# Verificar o que está usando a porta
lsof -i :8080
# Matar processo se necessário
kill -9 <PID>
```

## 📝 Scripts Úteis

### Script de deploy automático
```bash
#!/bin/bash
# deploy.sh

echo "🚀 Iniciando deploy..."

# Parar containers existentes
docker-compose down

# Build e restart
docker-compose up -d --build

# Aguardar banco ficar pronto
sleep 15

# Executar migrações
goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5432/planning" up

echo "✅ Deploy concluído! Aplicação em http://localhost:8080"
```

Para usar:
```bash
# Dar permissão de execução
chmod +x deploy.sh

# Executar deploy
./deploy.sh
```

### Health check
```bash
#!/bin/bash
# health-check.sh

curl -f http://localhost:8080/health || exit 1
```

## 🤝 Contribuição

1. Faça fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob licença MIT. Veja o arquivo `LICENSE` para mais detalhes.
