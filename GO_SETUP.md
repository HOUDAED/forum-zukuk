# Installation et utilisation (Go)

## Prérequis
- Go 1.21 ou supérieur
- Gmail ou un service SMTP configuré

## Installation

1. **Cloner les dépendances Go:**
```bash
go mod tidy
```

2. **Configurer les variables d'environnement:**
```bash
cp .env.example .env
# Éditer le fichier .env avec vos configurations
```

3. **Configuration Email (Gmail):**
   - Créer un mot de passe d'application Google: https://myaccount.google.com/apppasswords
   - Ajouter EMAIL_USER et EMAIL_PASSWORD dans .env

## Démarrage

**Mode développement:**
```bash
go run main.go
```

**Mode production (compiler d'abord):**
```bash
go build -o forum-zukuk
./forum-zukuk
```

Le serveur démarrera sur `http://localhost:3000`

## Architecture

```
.
├── main.go                 # Point d'entrée du serveur
├── models/                 # Structures de données
│   └── user.go
├── handlers/               # Handlers HTTP
│   └── auth.go
├── middleware/             # Middleware (rate limiting, etc.)
│   └── ratelimit.go
├── utils/                  # Utilitaires
│   ├── validation.go
│   ├── email.go
│   └── jwt.go
├── auth/                   # Frontend pages
├── sign/
├── go.mod                  # Dépendances Go
└── .env.example            # Variables d'environnement
```

## Points terminaux API

### Authentification
- `POST /api/auth/login` - Connexion avec OTP
- `POST /api/auth/verify-otp` - Vérification du code OTP
- `POST /api/auth/register` - Inscription nouvel utilisateur
- `GET /api/auth/google` - Infos Google OAuth
- `POST /api/auth/google-callback` - Callback Google OAuth
- `GET /api/auth/school` - Infos Portail Scolaire
- `POST /api/auth/school-callback` - Callback Portail Scolaire

### Santé
- `GET /api/health` - Vérifier l'état du serveur

## Dépendances Go

- **gin**: Framework web performant
- **golang-jwt**: Gestion des tokens JWT
- **golang.org/x/crypto**: Hachage bcrypt des mots de passe
- **godotenv**: Chargement des fichiers .env
- **gomail**: Envoi d'emails SMTP

## Notes

- Les utilisateurs et OTPs sont stockés en mémoire (À remplacer par une base de données pour la production)
- Les codes OTP expirent après 5 minutes pour la connexion et 10 minutes pour l'inscription
- Les mots de passe sont hachés avec bcrypt (coût 10)
- Les tokens JWT sont valides pendant 7 jours
