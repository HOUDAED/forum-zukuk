#!/bin/bash
# Test du système d'authentification - Exemples curl

# Configuration
API="http://localhost:3000"
EMAIL="test@example.com"
PASSWORD="TestPassword123"
PSEUDO="testuser"

echo "🧪 Tests du système d'authentification Forum Zukuk"
echo "=================================================="
echo ""

# Couleurs
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Health check
echo -e "${BLUE}1. Vérifier que le serveur est actif${NC}"
curl -X GET "$API/api/health" -w "\n"
echo ""

# Test 2: Inscription (Register)
echo -e "${BLUE}2. Inscription d'un nouvel utilisateur${NC}"
curl -X POST "$API/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"pseudo\": \"$PSEUDO\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"confirmPassword\": \"$PASSWORD\"
  }" \
  -w "\n"
echo ""

# Test 3: Inscription avec email déjà existant (devrait échouer)
echo -e "${BLUE}3. Essayer d'enregistrer avec le même email (devrait échouer)${NC}"
curl -X POST "$API/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"pseudo\": \"anotheruser\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"confirmPassword\": \"$PASSWORD\"
  }" \
  -w "\n"
echo ""

# Test 4: Login
echo -e "${BLUE}4. Connexion avec email/password${NC}"
echo "⚠️  Cela génère et envoie un code OTP par email"
curl -X POST "$API/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }" \
  -w "\n"
echo ""

# Test 5: Vérifier OTP (tu dois obtenir le code du email)
echo -e "${BLUE}5. Vérifier le code OTP${NC}"
echo "ℹ️  Remplace 123456 par le vrai code reçu par email"
curl -X POST "$API/api/auth/verify-otp" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"code\": \"123456\"
  }" \
  -w "\n"
echo ""

# Test 6: Login avec mauvais mot de passe
echo -e "${BLUE}6. Login avec mauvais password (devrait échouer)${NC}"
curl -X POST "$API/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"wrongpassword\"
  }" \
  -w "\n"
echo ""

# Test 7: Validation - password trop court
echo -e "${BLUE}7. Enregistrement avec password trop court (devrait échouer)${NC}"
curl -X POST "$API/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"pseudo\": \"shortpwd\",
    \"email\": \"short@example.com\",
    \"password\": \"short\",
    \"confirmPassword\": \"short\"
  }" \
  -w "\n"
echo ""

# Test 8: Validation - email invalide
echo -e "${BLUE}8. Enregistrement avec email invalide (devrait échouer)${NC}"
curl -X POST "$API/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"pseudo\": \"bademail\",
    \"email\": \"notanemail\",
    \"password\": \"ValidPassword123\",
    \"confirmPassword\": \"ValidPassword123\"
  }" \
  -w "\n"
echo ""

echo -e "${GREEN}✅ Tests terminés!${NC}"
echo ""
echo "📝 Notes:"
echo "- Remplace EMAIL, PASSWORD, PSEUDO avec tes valeurs"
echo "- Le code OTP est envoyé par email (vérifier spam)"
echo "- Les codes OTP expirent après 5-10 minutes"
echo "- Vérifier les logs du serveur pour plus d'infos"
