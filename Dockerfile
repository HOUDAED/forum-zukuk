FROM golang:1.25-alpine

# Installation des dépendances requises pour compiler avec CGO (nécessaire pour go-sqlite3)
RUN apk add --no-cache gcc musl-dev bash

# Définition du répertoire de travail dans le conteneur
WORKDIR /app

# Copie des fichiers de configuration des modules Go (s'ils existent)
COPY go.mod go.sum* ./
RUN go mod download

# Copie de tout le reste du code source
COPY . .

# Activation de CGO et compilation des trois binaires
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -o seed_app seed.go && \
    go build -o api_server api/server.go && \
    go build -o frontend_server frontend/server.go

# Création du script de démarrage qui enchaîne les 3 étapes
RUN echo '#!/bin/bash' > /app/start.sh && \
    echo 'set -e' >> /app/start.sh && \
    echo 'echo "🌱 Initialisation de la base de données..."' >> /app/start.sh && \
    echo 'cd /app && ./seed_app' >> /app/start.sh && \
    echo 'echo "🚀 Démarrage de l'\''API en arrière-plan..."' >> /app/start.sh && \
    echo 'cd /app/api && ../api_server &' >> /app/start.sh && \
    echo 'sleep 2' >> /app/start.sh && \
    echo 'echo "🎨 Démarrage du Frontend..."' >> /app/start.sh && \
    echo 'cd /app && ./frontend_server' >> /app/start.sh && \
    chmod +x /app/start.sh

# Exposition des ports du Frontend (3000) et de l'API (8081)
EXPOSE 3000 8081

# Lancement de l'application via le script
CMD ["/app/start.sh"]
