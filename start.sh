#!/bin/bash

# 🚀 Script de démarrage rapide pour Forum Zukuk Auth

echo "📦 Forum Zukuk - Système d'Authentification"
echo "==========================================="
echo ""

# Vérifier Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js n'est pas installé. Visite https://nodejs.org/"
    exit 1
fi

echo "✅ Node.js détecté: $(node --version)"
echo ""

# Vérifier package.json
if [ ! -f "package.json" ]; then
    echo "❌ package.json non trouvé"
    exit 1
fi

# Installer les dépendances si nécessaire
if [ ! -d "node_modules" ]; then
    echo "📥 Installation des dépendances..."
    npm install
    echo ""
fi

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
echo "⚠️  Pour servir les fichiers HTML:"
echo "   Terminal 1: npm start"
echo "   Terminal 2: npx http-server -p 8080"
echo "   Puis ouvrir: http://localhost:8080/auth.html"
echo ""

npm start
