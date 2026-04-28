# Forum Zukuk - Système d'Authentification Amélioré

## 🚀 Installation et Lancement

### Prérequis
- Node.js >= 14
- npm ou yarn

### Étapes

1. **Installer les dépendances**
   ```bash
   npm install
   ```

2. **Configurer les variables d'environnement**
   ```bash
   cp .env.example .env
   ```
   Éditez `.env` et remplissez:
   - `EMAIL_USER` et `EMAIL_PASSWORD` (Gmail avec [mot de passe d'app](https://myaccount.google.com/apppasswords))
   - `JWT_SECRET` (une clé forte aléatoire)

3. **Lancer le serveur**
   ```bash
   npm start
   ```
   Ou en développement avec rechargement auto:
   ```bash
   npm run dev
   ```

Le serveur écoute sur `http://localhost:3000`

---

## 📝 Architecture

### Backend (Node.js + Express)
- **auth.js** : Serveur principal avec toutes les routes

#### Routes API

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/auth/login` | Connexion avec email/password → envoie OTP |
| POST | `/api/auth/verify-otp` | Vérifie le code OTP → retourne JWT |
| POST | `/api/auth/register` | Inscription avec validation → envoie OTP |
| GET | `/api/health` | Vérifier que le serveur fonctionne |

### Frontend
- **auth.html** : Page de connexion
- **auth-client.js** : Logique login/OTP
- **sign.html** : Page d'inscription
- **sign-client.js** : Logique registration

---

## ✨ Améliorations implémentées

### Sécurité
✅ Hachage des mots de passe avec bcrypt  
✅ JWT tokens (durée 7 jours)  
✅ Rate limiting (5 tentatives login/15min, 3 inscriptions/heure)  
✅ Validation email stricte (regex)  
✅ Validation mot de passe (min 8 caractères)  
✅ CORS configuré  

### Validation
✅ **Côté serveur** : Tous les inputs validés
✅ **Côté client** : Validation en temps réel  

### UX/UI
✅ Messages d'erreur clairs  
✅ Loading states avec spinner  
✅ Animations lisses  
✅ Responsive design (mobile-first)  
✅ Gradient moderne  

### Email
✅ Nodemailer intégré  
✅ Templates HTML professionnels  
✅ Codes OTP avec expiration (5-10 min)  

---

## 🔄 Flow d'authentification

### Login
```
1. Email + Password → POST /api/auth/login
2. Validation + Hash check
3. Générer OTP → Envoyer par email
4. Afficher formulaire OTP
5. Code OTP → POST /api/auth/verify-otp
6. Retourner JWT token
7. Sauvegarder en localStorage
```

### Signup
```
1. Pseudo + Email + Password → POST /api/auth/register
2. Validation stricte
3. Hash password
4. Créer user (isVerified: false)
5. Générer OTP → Envoyer par email
6. Rediriger vers login
```

---

## 📦 Dépendances principales

- **express** : Framework Web
- **bcrypt** : Hash des mots de passe
- **jsonwebtoken** : Génération JWT
- **nodemailer** : Envoi emails
- **express-rate-limit** : Protection brute-force
- **cors** : Gestion des origines croisées

---

## 🔐 Configuration Google Mail

Pour utiliser Gmail comme service d'email:

1. Activer [2FA sur ton compte Google](https://myaccount.google.com/security)
2. Générer un [mot de passe d'application](https://myaccount.google.com/apppasswords)
3. Copier ce mot de passe dans `.env` sous `EMAIL_PASSWORD`

---

## ⚠️ À faire après

1. **Intégrer une vraie base de données** (MongoDB, PostgreSQL, MySQL)
2. **Ajouter OAuth** (Google, GitHub, portail écoles)
3. **Implémenter la vérification d'email** (click link dans email)
4. **Ajouter 2FA** (TOTP/Authy)
5. **Refresh tokens** (JWT refresh strategy)
6. **Migrations** pour la structure DB
7. **Tests unitaires** et d'intégration

---

## 🛠️ Troubleshooting

### Erreur: "Cannot find module"
```bash
npm install
```

### Emails ne s'envoient pas
- Vérifier `.env` (EMAIL_USER, EMAIL_PASSWORD)
- Activer les "applications moins sécurisées" (Gmail legacy) ou utiliser mot de passe d'app
- Vérifier les logs serveur

### CORS error au frontend
- Vérifier que `FRONTEND_URL` dans `.env` correspond à l'origin du navigateur
- Tester avec `curl` depuis terminal

---

## 📄 Licence

ISC
