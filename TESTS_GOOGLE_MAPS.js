// ======================================
// FICHIER DE TEST - Vérification de l'intégration
// ======================================

// À exécuter dans la console du navigateur une fois sur la page /carte

// === TEST 1: Vérifier que les objets globaux sont initialisés === 
function testGlobalObjects() {
    console.log("=== TEST 1: Objets globaux ===");
    console.log("map défini:", typeof map !== 'undefined' && map !== null);
    console.log("places[]:", Array.isArray(places), "(" + places.length + " activités)");
    console.log("markers[]:", Array.isArray(markers), "(" + markers.length + " marqueurs)");
    console.log("userLocation:", userLocation);
    console.log("selectedCategory:", selectedCategory);
    console.log("selectedMood:", selectedMood);
    console.log("✅ Test 1 complété\n");
}

// === TEST 2: Vérifier les APIs Google === 
function testGoogleAPIs() {
    console.log("=== TEST 2: Google APIs ===");
    console.log("google.maps.Map:", typeof google.maps.Map);
    console.log("google.maps.Marker:", typeof google.maps.Marker);
    console.log("google.maps.InfoWindow:", typeof google.maps.InfoWindow);
    console.log("google.maps.places.Autocomplete:", typeof google.maps.places.Autocomplete);
    console.log("google.maps.Geometry:", typeof google.maps.Geometry);
    console.log("markerClusterer:", typeof markerClusterer !== 'undefined');
    console.log("✅ Test 2 complété\n");
}

// === TEST 3: Vérifier les fonctions === 
function testFunctions() {
    console.log("=== TEST 3: Fonctions ===");
    const functions = [
        'initMapPage',
        'loadActivities',
        'showPlaces',
        'createMarker',
        'updateCard',
        'getCategoryMeta',
        'getDistanceFromUser',
        'showInfoWindow',
        'shareActivity'
    ];
    
    functions.forEach(fn => {
        console.log(fn + ":", typeof window[fn] === 'function' ? '✅' : '❌');
    });
    console.log("✅ Test 3 complété\n");
}

// === TEST 4: Charger et afficher les activités === 
async function testLoadActivities() {
    console.log("=== TEST 4: Chargement des activités ===");
    try {
        const response = await fetch('/api/activities');
        const data = await response.json();
        console.log("Activités chargées:", data.length);
        
        if (data.length > 0) {
            console.log("Première activité:", data[0]);
        }
        console.log("✅ Test 4 complété\n");
    } catch (error) {
        console.error("❌ Erreur:", error);
    }
}

// === TEST 5: Tester le calcul de distance === 
function testDistanceCalculation() {
    console.log("=== TEST 5: Calcul de distance ===");
    
    // Simuler une position utilisateur (Paris)
    userLocation = { lat: 48.8566, lng: 2.3522 };
    
    const testCoords = [
        { name: "Versailles", lat: 48.8049, lng: 2.1204 },
        { name: "Lyon", lat: 45.7640, lng: 4.8357 },
        { name: "Marseille", lat: 43.2965, lng: 5.3698 }
    ];
    
    testCoords.forEach(coord => {
        const distance = getDistanceFromUser(coord.lat, coord.lng);
        console.log(coord.name + ":", distance + " km");
    });
    console.log("✅ Test 5 complété\n");
}

// === TEST 6: Tester les métadonnées de catégories === 
function testCategoryMeta() {
    console.log("=== TEST 6: Métadonnées des catégories ===");
    
    const categories = ['sport', 'relax', 'talk', 'creative', 'social', 'nature'];
    
    categories.forEach(cat => {
        const meta = getCategoryMeta(cat);
        console.log(cat + ":", meta.icon, meta.label, meta.color);
    });
    console.log("✅ Test 6 complété\n");
}

// === TEST 7: Vérifier la géolocalisation === 
function testGeolocation() {
    console.log("=== TEST 7: Géolocalisation ===");
    
    if (!navigator.geolocation) {
        console.warn("Géolocalisation non supportée");
        return;
    }
    
    navigator.geolocation.getCurrentPosition(
        (position) => {
            console.log("Position obtenue:");
            console.log("Latitude:", position.coords.latitude);
            console.log("Longitude:", position.coords.longitude);
            console.log("Précision:", position.coords.accuracy + " m");
            console.log("✅ Test 7 complété\n");
        },
        (error) => {
            console.warn("Erreur géolocalisation:", error.message);
        }
    );
}

// === TEST 8: Tester le filtrage === 
function testFiltering() {
    console.log("=== TEST 8: Filtrage des activités ===");
    
    if (places.length === 0) {
        console.warn("Aucune activité chargée");
        return;
    }
    
    // Compter par catégorie
    const byCat = {};
    places.forEach(place => {
        byCat[place.category] = (byCat[place.category] || 0) + 1;
    });
    
    console.log("Activités par catégorie:");
    Object.entries(byCat).forEach(([cat, count]) => {
        console.log("  " + cat + ":", count);
    });
    
    // Compter par mood
    const byMood = {};
    places.forEach(place => {
        byMood[place.mood] = (byMood[place.mood] || 0) + 1;
    });
    
    console.log("Activités par mood:");
    Object.entries(byMood).forEach(([mood, count]) => {
        console.log("  " + mood + ":", count);
    });
    console.log("✅ Test 8 complété\n");
}

// === TEST 9: Valider les données des activités === 
function testActivityData() {
    console.log("=== TEST 9: Validation des données ===");
    
    let validCount = 0;
    let invalidCount = 0;
    
    places.forEach((place, index) => {
        const isValid = 
            place.id &&
            place.title &&
            place.category &&
            place.mood &&
            typeof place.lat === 'number' &&
            typeof place.lng === 'number' &&
            place.lat >= -90 && place.lat <= 90 &&
            place.lng >= -180 && place.lng <= 180;
        
        if (isValid) {
            validCount++;
        } else {
            invalidCount++;
            console.warn("Activité invalide:", place);
        }
    });
    
    console.log("Activités valides:", validCount);
    console.log("Activités invalides:", invalidCount);
    console.log("✅ Test 9 complété\n");
}

// === TEST 10: Performance === 
function testPerformance() {
    console.log("=== TEST 10: Performance ===");
    
    const start = performance.now();
    showPlaces();
    const end = performance.now();
    
    console.log("Temps de rendu des marqueurs:", (end - start).toFixed(2) + " ms");
    console.log("Nombre de marqueurs:", markers.length);
    
    if (markers.length > 0) {
        console.log("Temps moyen par marqueur:", ((end - start) / markers.length).toFixed(2) + " ms");
    }
    console.log("✅ Test 10 complété\n");
}

// === FONCTION MAÎTRE: Exécuter tous les tests === 
async function runAllTests() {
    console.clear();
    console.log("╔════════════════════════════════════════════════╗");
    console.log("║   TESTS D'INTÉGRATION GOOGLE MAPS - ZUKUK     ║");
    console.log("╚════════════════════════════════════════════════╝\n");
    
    testGlobalObjects();
    testGoogleAPIs();
    testFunctions();
    await testLoadActivities();
    testDistanceCalculation();
    testCategoryMeta();
    testGeolocation();
    testFiltering();
    testActivityData();
    testPerformance();
    
    console.log("╔════════════════════════════════════════════════╗");
    console.log("║          ✅ TOUS LES TESTS COMPLÉTÉS          ║");
    console.log("╚════════════════════════════════════════════════╝");
}

// Exécuter les tests
runAllTests();
