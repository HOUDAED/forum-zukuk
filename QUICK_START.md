# 🚀 Quick Start - Google Maps Zukuk

## ⚡ 5 minutes pour démarrer

### Étape 1: Obtenir une clé API (2 min)

1. Aller sur [Google Cloud Console](https://console.cloud.google.com/)
2. Créer un nouveau projet
3. Activer les APIs:
   - Google Maps JavaScript API
   - Places API
   - Geometry Library
4. Créer une clé API (Credentials → Create)
5. Copier la clé

### Étape 2: Intégrer la clé (1 min)

Remplacer dans `frontend/html/carte.html`:
```html
<script src="https://maps.googleapis.com/maps/api/js?key=VOTRE_CLE_ICI&libraries=places,geometry,marker&callback=initMapPage"></script>
```

### Étape 3: Démarrer le serveur (1 min)

```bash
# Depuis le répertoire du projet
go run main.go
```

### Étape 4: Tester (1 min)

1. Aller sur `http://localhost:8080/carte`
2. La carte doit s'afficher avec les activités
3. Cliquer sur un marqueur → infowindow s'affiche
4. Taper dans la recherche → filtre fonctionne

## 📊 Résultat attendu

- ✅ Carte centrée sur la France
- ✅ Marqueurs colorés avec emojis
- ✅ Clustering automatique
- ✅ Panel de filtres à gauche
- ✅ Fiche détail en bas
- ✅ Autocomplete dans le formulaire

## 🎮 Tester les fonctionnalités

### Test 1: Affichage des activités
```javascript
// Console du navigateur
loadActivities();
console.log(places.length); // Doit être > 0
```

### Test 2: Filtrer par catégorie
```javascript
selectedCategory = "sport";
showPlaces();
// Doit afficher seulement les activités "sport"
```

### Test 3: Recherche
```javascript
searchText = "yoga";
showPlaces();
// Doit afficher les activités contenant "yoga"
```

### Test 4: Géolocalisation
```javascript
detectUserLocation();
// Doit centrer la carte sur votre position
// Un marqueur bleu doit apparaître
```

### Test 5: Ajouter une activité
1. Remplir le formulaire dans le panel
2. Cliquer "Ajouter dans la BD"
3. L'activité doit apparaître sur la carte

## 🐛 Dépannage rapide

### Erreur: "RefererNotAllowedMapError"
→ Ajouter votre domaine dans Google Cloud Console

### La carte est blanche
→ Vérifier la clé API et les erreurs dans la console

### L'autocomplete ne marche pas
→ Vérifier que Places API est activée

### Les marqueurs ne s'affichent pas
→ Exécuter `loadActivities()` puis `showPlaces()`

## 📁 Fichiers importants

```
frontend/
├── html/carte.html          # Page principale
├── js/map.js                # Logique (2.0 refactorisée)
├── css/maps.css             # Styles complets
```

## 📚 Documentation

| Fichier | Contenu |
|---------|---------|
| README_GOOGLE_MAPS.md | Guide complet |
| GOOGLE_MAPS_INTEGRATION.md | Doc technique |
| SETUP_GOOGLE_MAPS.md | Installation |
| API_ENDPOINTS.md | Endpoints API |
| TESTS_GOOGLE_MAPS.js | Tests automatisés |

## 🔧 Configuration optionnelle

### Variables d'environnement
```bash
# .env
GOOGLE_MAPS_API_KEY=YOUR_KEY_HERE
ALLOWED_ORIGIN=http://localhost:3000
```

### Restrictions Google Cloud
1. HTTP referrers: `http://localhost:8080/*`
2. API restrictions: Maps, Places, Geometry

## 🎯 Checklist déploiement

- [ ] Clé API obtenue et restreinte
- [ ] Clé intégrée dans carte.html
- [ ] Serveur démarre sans erreurs
- [ ] Carte s'affiche correctement
- [ ] Marqueurs s'affichent
- [ ] Tests passent (TESTS_GOOGLE_MAPS.js)
- [ ] Géolocalisation fonctionne
- [ ] Autocomplete fonctionne
- [ ] Formulaire fonctionne
- [ ] Filtres fonctionnent

## 💡 Astuces

### Charger rapidement les activités
```javascript
loadActivities(); // Récupère du serveur
updateCard(places[0]); // Affiche le premier
showPlaces(); // Affiche les marqueurs
```

### Centrer sur une activité spécifique
```javascript
const activity = places[0];
map.panTo(new google.maps.LatLng(activity.lat, activity.lng));
map.setZoom(13);
updateCard(activity);
```

### Obtenir les informations utilisateur
```javascript
console.log("Position:", userLocation);
console.log("Catégorie sélectionnée:", selectedCategory);
console.log("Mood sélectionné:", selectedMood);
console.log("Texte recherché:", searchText);
```

### Rafraîchir les marqueurs
```javascript
showPlaces(); // Redessine tous les marqueurs
// Applique les filtres actuels automatiquement
```

## 🌐 URLs utiles

- **App Locale**: http://localhost:8080/carte
- **Google Cloud Console**: https://console.cloud.google.com/
- **Maps Documentation**: https://developers.google.com/maps
- **Places API Docs**: https://developers.google.com/maps/documentation/places

## 📞 Support

### Problème ?
1. Lire le README_GOOGLE_MAPS.md
2. Vérifier les erreurs dans la console
3. Lancer les tests (TESTS_GOOGLE_MAPS.js)
4. Consulter SETUP_GOOGLE_MAPS.md

### Erreur Google Maps ?
1. Console → Onglet Réseau
2. Chercher `maps.googleapis.com`
3. Vérifier le status et la réponse

## 🎓 Prochaines étapes

Après le démarrage:
1. Personnaliser les couleurs des catégories
2. Ajouter plus de catégories si besoin
3. Configurer l'upload de photos
4. Mettre en place les notifications
5. Ajouter une heatmap

## ✅ C'est prêt !

Vous avez maintenant une intégration Google Maps complète et fonctionnelle avec:
- ✅ Affichage des activités
- ✅ Clustering intelligent
- ✅ Filtrage avancé
- ✅ Géolocalisation
- ✅ Autocomplete
- ✅ Ajout d'activités
- ✅ Partage d'activités

**Bon usage ! 🚀**

---

**Besoin d'aide ?** Consultez la documentation complète dans les fichiers .md
