# Configuration Google Maps API

## Étapes d'installation

### 1. Obtenir une clé API Google Maps

1. Aller sur [Google Cloud Console](https://console.cloud.google.com/)
2. Créer un nouveau projet ou sélectionner un existant
3. Activer les APIs suivantes:
   - Google Maps JavaScript API
   - Places API
   - Geometry Library
   - Maps Static API (optionnel)

### 2. Créer une clé API
1. Aller à "APIs & Services" > "Credentials"
2. Cliquer sur "Create Credentials" > "API Key"
3. Copier la clé générée

### 3. Restreindre la clé API (recommandé)
1. Éditer la clé API
2. Sous "Application restrictions", sélectionner "HTTP referrers"
3. Ajouter:
   - `http://localhost:8080/*`
   - `http://localhost:3000/*` (dev)
   - `https://yourdomain.com/*` (production)

4. Sous "API restrictions", sélectionner:
   - Google Maps JavaScript API
   - Places API
   - Geometry Library

### 4. Intégrer la clé dans le projet

#### Option A: Variable d'environnement (recommandé)
Créer un fichier `.env`:
```
GOOGLE_MAPS_API_KEY=YOUR_API_KEY_HERE
```

Puis importer dans `carte.html`:
```html
<script src="https://maps.googleapis.com/maps/api/js?key=${process.env.GOOGLE_MAPS_API_KEY}&libraries=places,geometry,marker&callback=initMapPage"></script>
```

#### Option B: Remplacer directement dans le HTML
```html
<script src="https://maps.googleapis.com/maps/api/js?key=YOUR_API_KEY_HERE&libraries=places,geometry,marker&callback=initMapPage"></script>
```

## Variables d'environnement

### Frontend (.env ou .env.local)
```
REACT_APP_GOOGLE_MAPS_API_KEY=YOUR_API_KEY_HERE
REACT_APP_API_BASE_URL=http://localhost:8080
```

### Backend (optionnel, pour Geocoding API)
```
GOOGLE_MAPS_API_KEY=YOUR_API_KEY_HERE
GOOGLE_GEOCODING_API_KEY=YOUR_API_KEY_HERE
```

## Test de la configuration

### 1. Vérifier la console
```javascript
// Dans la console du navigateur
google.maps.Map // Doit être défini
google.maps.places.Autocomplete // Doit être défini
google.maps.Geometry // Doit être défini
```

### 2. Tester le chargement de la carte
- Ouvrir http://localhost:8080/carte
- La carte doit s'afficher centrée sur la France
- Les marqueurs d'activités doivent apparaître

### 3. Tester l'autocomplete
- Cliquer sur le champ "Ville / adresse"
- Taper "Paris"
- Les suggestions doivent s'afficher

### 4. Tester la géolocalisation
- Cliquer sur le bouton "📍" (geolocation)
- Accepter l'accès à la position
- La carte doit se centrer sur votre position

## Dépannage

### Erreur: "RefererNotAllowedMapError"
- La clé API n'est pas correctement restreinte
- Vérifier que le domaine/referrer actuel est dans la liste blanche
- Ajouter `*` temporairement pour tester

### Erreur: "Google Maps API error: MissingKeyMapError"
- La clé API n'est pas fournie ou est vide
- Vérifier que `GOOGLE_MAPS_API_KEY` est défini
- Vérifier que la clé est correctement insérée dans le script

### Erreur: "Geometry is not defined"
- La bibliothèque `geometry` n'est pas chargée
- Vérifier que `&libraries=places,geometry,marker` est présent dans l'URL

### L'autocomplete ne fonctionne pas
- Vérifier que Places API est activée
- Vérifier que la clé API a accès à Places API
- Vérifier les erreurs dans la console

### Les icônes d'adresse n'apparaissent pas
- Les SVG DataURL doivent être correctement échappés
- Vérifier qu'aucun caractère spécial ne casse l'URL
- Tester avec des icônes simples d'abord

## Coûts

### Tarification Google Maps Platform
- **Maps JavaScript API**: $7 par 1000 cartes chargées (gratuit jusqu'à 28k/mois)
- **Places API**: $17 par 1000 requêtes (gratuit jusqu'à 150k/mois)
- **Geometry Library**: Gratuit

Estimé pour 10k utilisateurs/mois: ~$200-300

## Sécurité

### Bonnes pratiques
1. **Ne pas exposer la clé API en public**
   - Utiliser des variables d'environnement
   - Ne pas la commit dans Git

2. **Restreindre les APIs**
   - Ne permettre que les APIs nécessaires
   - Restreindre par domaine

3. **Monitorer l'utilisation**
   - Activer les alertes de coût
   - Vérifier les statistiques d'utilisation

4. **Rotation des clés**
   - Renouveler les clés tous les 6-12 mois
   - Garder des clés de sauvegarde

## Ressources

- [Documentation Google Maps API](https://developers.google.com/maps/documentation)
- [Places API Documentation](https://developers.google.com/maps/documentation/places/web-service)
- [Geometry Library Documentation](https://developers.google.com/maps/documentation/javascript/geometry)
- [Pricing Calculator](https://cloud.google.com/maps-platform/pricing/sheet)
