#!/bin/bash

# 🚀 Script de démarrage rapide pour Forum Zukuk Auth (Go)

echo "📦 Forum Zukuk - Système d'Authentification"
echo "==========================================="
echo ""

# Vérifier Go
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé. Visite https://golang.org/"
    exit 1
fi

echo "✅ Go détecté: $(go version)"
echo ""

# Vérifier go.mod
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod non trouvé"
    exit 1
fi

# Télécharger les dépendances
echo "📥 Téléchargement des dépendances Go..."
go mod tidy
echo ""

# Vérifier .env
if [ ! -f ".env" ]; then
    echo "⚠️  .env non trouvé!"
    echo ""
    echo "📝 Créer un fichier .env avec:"
    echo "---"
    cat .env.example
    echo "---"
    echo ""
    echo "Puis relancer ce script."
    exit 1
fi

echo "✅ Configuration trouvée (.env)"
echo ""

# Démarrer le serveur
echo "🚀 Démarrage du serveur sur http://localhost:3000"
echo ""
echo "📋 Pour tester:"
echo "   - Ouvrir: http://localhost:3000/auth.html"
echo "   - Ouvrir: http://localhost:3000/sign.html"
echo ""
echo "💡 Note: Les fichiers HTML doivent être servis via HTTP"
echo "   (pas en file:// pour que fetch fonctionne)"
echo ""

go run main.go
