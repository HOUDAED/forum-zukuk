# Intégration Google Maps - Documentation

## Vue d'ensemble
L'intégration Google Maps est maintenant complètement fonctionnelle avec les activités bien-être. La carte affiche les activités avec clustering intelligent, infowindows enrichies et fonctionnalités de géolocalisation.

## Fonctionnalités principales

### 1. **Affichage des activités sur la carte**
- Les activités sont chargées via l'API `/api/activities`
- Chaque activité est représentée par un marqueur coloré selon sa catégorie
- Les marqueurs sont en cache pour améliorer les performances

### 2. **Clustering des marqueurs**
- Utilise la bibliothèque `@googlemaps/markerclusterer`
- Regroupe les marqueurs proches pour une meilleure lisibilité
- Se réajuste automatiquement lors du zoom

### 3. **Métadonnées des catégories**
```javascript
- sport (⚽): Bleu - Activités sportives
- relax (🧘): Violet - Relaxation & bien-être
- talk (💬): Vert - Groupes de parole
- creative (🎨): Rose - Activités créatives
- social (🤝): Orange - Lieux sociaux
- nature (🌿): Teal - Activités nature
```

### 4. **InfoWindows enrichies**
Au clic sur un marqueur:
- Affiche un popup stylisé avec gradient de couleur
- Montre la distance depuis la position de l'utilisateur
- Affiche tous les détails: description, rating, nombre de personnes, mood
- Boutons d'itinéraire Google Maps et de partage

### 5. **Géolocalisation**
- Détection automatique de la position utilisateur
- Affichage d'un marqueur bleu pour la position
- Calcul automatique des distances (en km) vers les activités
- Mise à jour des distances lors de chaque rechargement d'activités

### 6. **Autocomplete Google Places**
- Champ "Ville / adresse" avec suggestions en temps réel
- Restreint à la France (type: geocode, establishment)
- Récupère automatiquement les coordonnées GPS
- Centre la carte sur le lieu sélectionné

### 7. **Filtrage des activités**
- Filtres par catégorie (7 catégories)
- Filtres par mood (5 moods)
- Recherche textuelle en temps réel
- Comptage automatique des activités par catégorie

### 8. **Ajout d'activités**
Formulaire complet avec:
- Titre et description
- Localisation avec autocomplete
- Catégorie et mood
- Coordonnées GPS (auto-remplies ou manuelles)
- Upload de photo (converti en base64)
- Bouton "Utiliser ma position" pour les coordonnées

### 9. **Partage d'activités**
- Utilise l'API native `navigator.share()` si disponible
- Fallback: copie le lien dans le presse-papier
- Inclut le titre de l'activité et un message personnalisé

## Configuration requise

### API Google Maps
```html
<script src="https://maps.googleapis.com/maps/api/js?key=GOOGLE_MAPS_API_KEY&libraries=places,geometry,marker&callback=initMapPage"></script>
```

Remplacer `GOOGLE_MAPS_API_KEY` par votre clé API.

### Restrictions recommandées dans Google Cloud Console:
- **HTTP referrers**: `http://localhost:8080/*`
- **API restrictions**: 
  - Maps JavaScript API
  - Places API
  - Geometry Library

### Bibliothèque Marker Clusterer
```html
<script src="https://cdn.jsdelivr.net/npm/@googlemaps/markerclusterer@2.0.0/dist/index.min.js"></script>
```

## Structure des données

### Objet Place/Activity
```javascript
{
    id: number,
    title: string,
    description: string,
    address: string,
    lat: number,
    lng: number,
    category: "sport" | "relax" | "talk" | "creative" | "social" | "nature",
    mood: "relaxant" | "social" | "actif" | "faible-énergie",
    rating: number,
    people: number,
    photo: string (URL ou base64),
    city: string,
    icon: string (emoji)
}
```

## Points clés de l'implémentation

### Distance utilisateur
La distance est calculée en temps réel depuis la position de l'utilisateur:
```javascript
function getDistanceFromUser(lat, lng) {
    // Utilise la formule de Haversine
    // Retourne la distance en km avec 1 décimale
}
```

### Couleurs dynamiques
Chaque catégorie a sa propre couleur qui:
- Style le marqueur SVG
- Teinte l'infowindow (avec dégradé)
- Teinte la carte de détail à droite

### Icônes SVG
Les marqueurs sont des SVG encodés en dataURL pour:
- Chargement rapide (pas d'images externes)
- Rendering natif Google Maps
- Personnalisation totale

## Dépannage

### Les marqueurs ne s'affichent pas
1. Vérifier la clé API Google Maps
2. Vérifier que les bibliothèques sont chargées
3. Ouvrir la console et chercher les erreurs

### Les distances ne s'affichent pas
1. Autoriser l'accès à la géolocalisation
2. S'assurer que les coordonnées des activités sont valides
3. Vérifier que `userLocation` est défini

### L'autocomplete ne fonctionne pas
1. Vérifier que la bibliothèque `places` est chargée
2. S'assurer que l'input a l'ID `activity-city`
3. Vérifier la clé API (Places API activée)

## Performance

- **Caching des activités**: Les activités sont stockées dans `activitiesCache` pour accès rapide
- **Réutilisation d'infowindow**: Une seule infowindow est créée et réutilisée
- **Lazy loading**: Les marqueurs sont créés à la demande lors du filtrage
- **Clustering**: Améliore les performances avec beaucoup de marqueurs

## Améliorations possibles

1. **Heatmap**: Afficher une heatmap des zones d'activités densifiées
2. **Directions API**: Calcul d'itinéraire sur la carte
3. **Street View**: Intégration Street View sur la carte
4. **Statistiques**: Graphiques de répartition géographique
5. **Favoris**: Sauvegarde des activités favorites localement
