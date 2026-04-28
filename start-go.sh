#!/bin/bash

# Script de démarrage pour le serveur Go

# Vérifier si Go est installé
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé. Veuillez installer Go 1.21 ou supérieur."
    exit 1
fi

# Vérifier la version de Go
GO_VERSION=$(go version | grep -oP 'go\K[0-9.]+')
echo "✅ Go trouvé: version $GO_VERSION"

# Télécharger les dépendances
echo "📦 Téléchargement des dépendances..."
go mod tidy

# Vérifier si le fichier .env existe
if [ ! -f .env ]; then
    echo "⚠️  Fichier .env non trouvé."
    echo "   Créez un fichier .env en copiant .env.example:"
    echo "   cp .env.example .env"
    echo "   puis configurez vos variables d'environnement."
fi

# Démarrer le serveur
echo "🚀 Démarrage du serveur..."
go run main.go
