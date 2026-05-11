# API Endpoints - Google Maps Integration

## 📍 Endpoints disponibles

### GET `/api/activities`
Récupère toutes les activités

**Response:**
```json
[
    {
        "id": 1,
        "title": "Yoga du matin",
        "city": "Paris",
        "address": "123 Rue de Paris, 75001 Paris",
        "category": "relax",
        "mood": "relaxant",
        "lat": 48.8566,
        "lng": 2.3522,
        "description": "Session de yoga détente pour débuter la journée",
        "distance": "1.5",
        "rating": "4.8",
        "people": "12",
        "icon": "🧘",
        "color": "#8b5cf6",
        "photo": "https://example.com/photo.jpg"
    }
]
```

**Status Codes:**
- `200 OK` - Activités retournées
- `500 Internal Server Error` - Erreur serveur

---

### POST `/api/activities`
Crée une nouvelle activité

**Request Body:**
```json
{
    "title": "Yoga du matin",
    "city": "Paris",
    "category": "relax",
    "mood": "relaxant",
    "lat": 48.8566,
    "lng": 2.3522,
    "description": "Session de yoga détente pour débuter la journée",
    "maxPlaces": 10,
    "photo": "data:image/png;base64,..." // Optionnel
}
```

**Response:**
```json
{
    "id": 1,
    "title": "Yoga du matin",
    "city": "Paris",
    "address": "Paris, France",
    "category": "relax",
    "mood": "relaxant",
    "lat": 48.8566,
    "lng": 2.3522,
    "description": "Session de yoga détente pour débuter la journée",
    "distance": "0.0",
    "rating": "0.0",
    "people": "10",
    "icon": "🧘",
    "color": "#8b5cf6",
    "photo": "..."
}
```

**Status Codes:**
- `200 OK` - Activité créée
- `400 Bad Request` - Données invalides
- `500 Internal Server Error` - Erreur serveur

**Validation:**
- `title`: Requis, string
- `city`: Requis, string
- `category`: Requis, enum (sport|relax|talk|creative|social|nature)
- `mood`: Requis, enum (relaxant|social|actif|faible-énergie)
- `lat`: Requis, number (-90 à 90)
- `lng`: Requis, number (-180 à 180)
- `description`: Optionnel, string
- `maxPlaces`: Optionnel, number (> 0)
- `photo`: Optionnel, string (base64 ou URL)

---

### GET `/api/activities/:id`
Récupère une activité spécifique

**Response:** Même structure que GET /api/activities

**Status Codes:**
- `200 OK` - Activité retournée
- `404 Not Found` - Activité non trouvée
- `500 Internal Server Error` - Erreur serveur

---

### PUT `/api/activities/:id`
Modifie une activité existante

**Request Body:** Identique à POST /api/activities

**Response:** Activité modifiée

**Status Codes:**
- `200 OK` - Activité modifiée
- `400 Bad Request` - Données invalides
- `404 Not Found` - Activité non trouvée
- `500 Internal Server Error` - Erreur serveur

---

### DELETE `/api/activities/:id`
Supprime une activité

**Response:**
```json
{
    "message": "Activité supprimée"
}
```

**Status Codes:**
- `200 OK` - Activité supprimée
- `404 Not Found` - Activité non trouvée
- `500 Internal Server Error` - Erreur serveur

---

## 🔐 Authentification

### POST `/api/auth/register`
Crée un nouveau compte utilisateur

### POST `/api/auth/login`
Authentifie un utilisateur

### POST `/api/auth/logout`
Déconnecte l'utilisateur

---

## 📝 Exemple de requête JavaScript

### Charger les activités
```javascript
fetch('/api/activities')
    .then(response => response.json())
    .then(data => {
        console.log('Activités:', data);
        places = data;
        showPlaces();
    })
    .catch(error => console.error('Erreur:', error));
```

### Créer une activité
```javascript
const activity = {
    title: "Yoga du matin",
    city: "Paris",
    category: "relax",
    mood: "relaxant",
    lat: 48.8566,
    lng: 2.3522,
    description: "Session de yoga détente",
    maxPlaces: 10,
    photo: canvasDataUrl // Optionnel
};

fetch('/api/activities', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(activity)
})
    .then(response => response.json())
    .then(data => {
        console.log('Activité créée:', data);
        places.unshift(data);
        showPlaces();
    })
    .catch(error => console.error('Erreur:', error));
```

### Modifier une activité
```javascript
const activity = {
    title: "Yoga du soir",
    // ... autres champs
};

fetch('/api/activities/1', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(activity)
})
    .then(response => response.json())
    .then(data => console.log('Activité modifiée:', data))
    .catch(error => console.error('Erreur:', error));
```

### Supprimer une activité
```javascript
fetch('/api/activities/1', {
    method: 'DELETE'
})
    .then(response => response.json())
    .then(data => console.log('Activité supprimée:', data))
    .catch(error => console.error('Erreur:', error));
```

---

## 🌍 Autres endpoints

### GET `/api/health`
Vérifier la santé du serveur

**Response:**
```json
{
    "status": "ok"
}
```

---

## 📋 Rate Limiting

- **Limite**: 10 requêtes par minute par IP
- **Headers**: 
  - `X-RateLimit-Limit`: 10
  - `X-RateLimit-Remaining`: 9
  - `X-RateLimit-Reset`: 1620000000

**Error Response (429):**
```json
{
    "error": "Too many requests"
}
```

---

## 🔄 Filtering et Pagination

### Query Parameters

#### Par catégorie
```
GET /api/activities?category=relax
```

#### Par mood
```
GET /api/activities?mood=relaxant
```

#### Par localisation (rayon en km)
```
GET /api/activities?lat=48.8566&lng=2.3522&radius=10
```

#### Par texte (recherche)
```
GET /api/activities?search=yoga
```

#### Pagination
```
GET /api/activities?page=1&limit=20
```

#### Combiné
```
GET /api/activities?category=relax&mood=relaxant&page=1&limit=10
```

---

## 📊 Statistiques

### GET `/api/statistics`
Récupère les statistiques globales

**Response:**
```json
{
    "totalActivities": 150,
    "byCategory": {
        "sport": 45,
        "relax": 30,
        "talk": 20,
        "creative": 25,
        "social": 20,
        "nature": 10
    },
    "byMood": {
        "relaxant": 50,
        "social": 40,
        "actif": 35,
        "faible-énergie": 25
    },
    "averageRating": 4.5,
    "totalParticipants": 2850
}
```

---

## ⚠️ Codes d'erreur courants

| Code | Message | Cause |
|------|---------|-------|
| 400 | Bad Request | Données invalides |
| 401 | Unauthorized | Authentification requise |
| 403 | Forbidden | Accès refusé |
| 404 | Not Found | Ressource non trouvée |
| 429 | Too Many Requests | Rate limit dépassée |
| 500 | Internal Server Error | Erreur serveur |
| 503 | Service Unavailable | Service indisponible |

---

## 🧪 Test des endpoints

### Avec cURL
```bash
# Récupérer toutes les activités
curl http://localhost:8080/api/activities

# Créer une activité
curl -X POST http://localhost:8080/api/activities \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Yoga",
    "city": "Paris",
    "category": "relax",
    "mood": "relaxant",
    "lat": 48.8566,
    "lng": 2.3522,
    "description": "Yoga",
    "maxPlaces": 10
  }'

# Modifier une activité
curl -X PUT http://localhost:8080/api/activities/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Yoga du soir", ...}'

# Supprimer une activité
curl -X DELETE http://localhost:8080/api/activities/1
```

### Avec Postman
1. Importer les requêtes
2. Configurer l'URL de base
3. Exécuter les requêtes

---

## 📚 Documentation complète

Voir les fichiers:
- [README_GOOGLE_MAPS.md](./README_GOOGLE_MAPS.md)
- [GOOGLE_MAPS_INTEGRATION.md](./GOOGLE_MAPS_INTEGRATION.md)
- [map_api.go](./backend/handlers/map_api.go)

---

**Dernière mise à jour**: 2026-05-06
