// ====================================
// GOOGLE MAPS - Intégration complète
// ====================================

var map;
var places = [];
var markers = [];
var markerCluster = null;
var selectedCategory = "all";
var selectedMood = "all";
var searchText = "";
var userLocation = null;
var autocomplete = null;
var mapInitialized = false;
var infoWindow = null;
var activitiesCache = {};

// === UTILITAIRES ===

function escapeHtml(text) {
    return String(text)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

function getCategoryMeta(category) {
    const meta = {
        "sport": { icon: "⚽", color: "#3b82f6", label: "Activités sportives" },
        "relax": { icon: "🧘", color: "#8b5cf6", label: "Relaxation & bien-être" },
        "talk": { icon: "💬", color: "#22c55e", label: "Groupes de parole" },
        "creative": { icon: "🎨", color: "#ec4899", label: "Activités créatives" },
        "social": { icon: "🤝", color: "#f97316", label: "Lieux sociaux" },
        "nature": { icon: "🌿", color: "#14b8a6", label: "Activités nature" }
    };
    return meta[category] || { icon: "📍", color: "#64748b", label: "Activité" };
}

function getDistanceFromUser(lat, lng) {
    if (!userLocation) return null;
    const R = 6371; // Rayon de la Terre en km
    const dLat = (lat - userLocation.lat) * Math.PI / 180;
    const dLng = (lng - userLocation.lng) * Math.PI / 180;
    const a = Math.sin(dLat/2) * Math.sin(dLat/2) +
              Math.cos(userLocation.lat * Math.PI / 180) * Math.cos(lat * Math.PI / 180) *
              Math.sin(dLng/2) * Math.sin(dLng/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    return (R * c).toFixed(1);
}

// === INITIALISATION ===

function initMapPage() {
    if (mapInitialized) return; // Éviter les double-appels
    mapInitialized = true;
    
    initMap();
    initAutocomplete();
    bindFilterButtons();
    bindForm();
    loadActivities();
    detectUserLocation();
}

function initMap() {
    // Centre sur la France
    const defaultCenter = { lat: 46.603354, lng: 1.888334 };
    
    map = new google.maps.Map(document.getElementById("map"), {
        zoom: 6,
        center: defaultCenter,
        mapTypeControl: false,
        fullscreenControl: true,
        fullscreenControlOptions: { position: google.maps.ControlPosition.RIGHT_TOP },
        streetViewControl: false,
        zoomControl: true,
        zoomControlOptions: { position: google.maps.ControlPosition.RIGHT_TOP },
        styles: [
            {
                featureType: "poi",
                stylers: [{ visibility: "off" }]
            }
        ]
    });

    addLocationButton();
    
    // Créer une infowindow réutilisable
    infoWindow = new google.maps.InfoWindow({
        disableAutoPan: false
    });
}

function addLocationButton() {
    const controlDiv = document.createElement("div");
    controlDiv.style.margin = "10px";
    
    const button = document.createElement("button");
    button.textContent = "📍";
    button.style.backgroundColor = "#fff";
    button.style.border = "1px solid #ccc";
    button.style.borderRadius = "4px";
    button.style.boxShadow = "0 2px 1px rgba(0,0,0,0.298)";
    button.style.cursor = "pointer";
    button.style.padding = "7px 10px";
    button.style.marginRight = "10px";
    button.style.fontSize = "16px";
    button.style.fontWeight = "bold";
    button.title = "Localiser ma position";
    
    button.addEventListener("click", () => detectUserLocation());
    controlDiv.appendChild(button);
    map.controls[google.maps.ControlPosition.RIGHT_CENTER].push(controlDiv);
}

// === GÉOLOCALISATION ===

function detectUserLocation() {
    if (!navigator.geolocation) {
        console.warn("Géolocalisation non supportée.");
        return;
    }
    
    navigator.geolocation.getCurrentPosition(
        (position) => {
            userLocation = {
                lat: position.coords.latitude,
                lng: position.coords.longitude
            };
            map.panTo(userLocation);
            map.setZoom(12);
            
            // Ajouter un marqueur pour l'utilisateur
            new google.maps.Marker({
                position: userLocation,
                map: map,
                title: "Ma position",
                icon: "https://maps.google.com/mapfiles/ms/icons/blue-dot.png"
            });
            
            // Mettre à jour les distances
            showPlaces();
        },
        () => {
            console.warn("Impossible d'obtenir la position.");
        }
    );
}

// === CHARGEMENT DES ACTIVITÉS ===

function loadActivities() {
    fetch("/api/activities")
        .then(response => response.json())
        .then(data => {
            places = data || [];
            
            // Mettre en cache les activités
            places.forEach(place => {
                activitiesCache[place.id] = place;
            });
            
            updateCounts();
            showPlaces();
            
            if (places.length > 0) {
                updateCard(places[0]);
            }
        })
        .catch(error => {
            console.error("Erreur chargement activités:", error);
        });
}

function showPlaces() {
    // Nettoyer l'ancien cluster
    if (markerCluster) {
        markerCluster.clearMarkers();
    }
    
    // Supprimer les anciens marqueurs
    markers.forEach(marker => marker.setMap(null));
    markers = [];

    const bounds = new google.maps.LatLngBounds();
    let hasMarkers = false;

    places.forEach(place => {
        if (!matchesFilter(place)) return;

        const marker = createMarker(place);
        markers.push(marker);
        hasMarkers = true;
        
        bounds.extend(new google.maps.LatLng(place.lat, place.lng));

        // Événement au clic sur le marqueur
        marker.addListener("click", () => {
            updateCard(place);
            showInfoWindow(marker, place);
            map.panTo(marker.getPosition());
            map.setZoom(14);
        });
    });

    // Créer le cluster de marqueurs
    if (markers.length > 0) {
        markerCluster = new markerClusterer.MarkerClusterer({ map, markers });
    }

    // Adapter la vue
    if (markers.length > 1) {
        map.fitBounds(bounds, { top: 50, right: 50, bottom: 200, left: 50 });
    } else if (markers.length === 1) {
        map.panTo(markers[0].getPosition());
        map.setZoom(13);
    }
}

function createMarker(place) {
    const meta = getCategoryMeta(place.category);
    const distance = getDistanceFromUser(place.lat, place.lng);
    
    const marker = new google.maps.Marker({
        position: new google.maps.LatLng(place.lat, place.lng),
        map: map,
        title: place.title,
        icon: createSvgIcon(meta.color, meta.icon),
        animation: google.maps.Animation.DROP
    });

    marker.place = place;
    marker.distance = distance;
    return marker;
}

function createSvgIcon(color, icon) {
    const svg = encodeURIComponent(
        `<svg width="42" height="42" viewBox="0 0 42 42" xmlns="http://www.w3.org/2000/svg">
            <rect width="42" height="42" rx="12" fill="${color}" />
            <text x="21" y="28" font-size="22" text-anchor="middle" fill="white" font-family="Arial">${icon}</text>
        </svg>`
    );

    return {
        url: `data:image/svg+xml;charset=UTF-8,${svg}`,
        scaledSize: new google.maps.Size(42, 42),
        anchor: new google.maps.Point(21, 42)
    };
}

function showInfoWindow(marker, place) {
    const meta = getCategoryMeta(place.category);
    const distance = marker.distance ? ` (${marker.distance} km)` : "";
    
    const content = `
        <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial; width: 320px; padding: 0;">
            <div style="background: linear-gradient(135deg, ${meta.color}, ${adjustBrightness(meta.color, -20)}); color: white; padding: 16px; border-radius: 8px 8px 0 0;">
                <div style="font-size: 24px; margin-bottom: 8px;">${meta.icon}</div>
                <h3 style="margin: 0 0 4px 0; font-size: 18px;">${escapeHtml(place.title)}</h3>
                <p style="margin: 0; font-size: 12px; opacity: 0.9;">${escapeHtml(meta.label)}</p>
            </div>
            <div style="padding: 16px; background: white;">
                <p style="margin: 0 0 12px 0; font-size: 14px; color: #333; line-height: 1.5;">
                    ${escapeHtml(place.description || "Pas de description disponible")}
                </p>
                <div style="font-size: 13px; color: #666; margin-bottom: 12px;">
                    <div style="margin-bottom: 6px;">
                        📍 <strong>${escapeHtml(place.address || place.city)}</strong>${distance}
                    </div>
                    <div style="margin-bottom: 6px;">
                        ⭐ <strong>${place.rating || "0.0"}</strong> | 👥 <strong>${place.people} personnes</strong>
                    </div>
                    <div style="margin-bottom: 6px;">
                        🎯 Mood: <strong>${escapeHtml(place.mood)}</strong>
                    </div>
                </div>
                <div style="display: flex; gap: 8px;">
                    <a href="https://www.google.com/maps/dir/?api=1&destination=${place.lat},${place.lng}" 
                       target="_blank" rel="noreferrer"
                       style="flex: 1; text-align: center; padding: 8px; background: #3b82f6; color: white; text-decoration: none; border-radius: 4px; font-weight: bold; font-size: 13px;">
                        Itinéraire
                    </a>
                    <button onclick="shareActivity(${place.id})"
                            style="flex: 1; padding: 8px; background: #8b5cf6; color: white; border: none; border-radius: 4px; cursor: pointer; font-weight: bold; font-size: 13px;">
                        Partager
                    </button>
                </div>
            </div>
        </div>
    `;

    infoWindow.setContent(content);
    infoWindow.open(map, marker);
}

function adjustBrightness(color, percent) {
    // Convertir le couleur et ajuster la luminosité
    const num = parseInt(color.replace("#", ""), 16);
    const amt = Math.round(2.55 * percent);
    const R = (num >> 16) + amt;
    const G = (num >> 8 & 0x00FF) + amt;
    const B = (num & 0x0000FF) + amt;
    
    return "#" + (0x1000000 + (R < 255 ? R < 1 ? 0 : R : 255) * 0x10000 +
        (G < 255 ? G < 1 ? 0 : G : 255) * 0x100 +
        (B < 255 ? B < 1 ? 0 : B : 255))
        .toString(16).slice(1);
}

function matchesFilter(place) {
    const text = searchText.toLowerCase();
    if (selectedCategory !== "all" && place.category !== selectedCategory) {
        return false;
    }
    if (selectedMood !== "all" && place.mood !== selectedMood) {
        return false;
    }
    if (text) {
        const haystack = (place.title + " " + place.description + " " + place.city).toLowerCase();
        if (haystack.indexOf(text) === -1) {
            return false;
        }
    }
    return true;
}

function updateCounts() {
    const counts = {
        all: 0,
        sport: 0,
        relax: 0,
        talk: 0,
        creative: 0,
        social: 0,
        nature: 0,
    };
    
    places.forEach(place => {
        counts.all += 1;
        if (counts[place.category] !== undefined) {
            counts[place.category] += 1;
        }
    });
    
    Object.keys(counts).forEach(key => {
        const el = document.getElementById("count-" + key);
        if (el) el.textContent = counts[key];
    });
}

function updateCard(place) {
    const meta = getCategoryMeta(place.category);
    const distance = getDistanceFromUser(place.lat, place.lng);
    const distanceText = distance ? ` (${distance} km)` : "";

    const title = document.getElementById("place-title");
    const description = document.getElementById("place-description");
    const city = document.getElementById("place-city");
    const rating = document.getElementById("place-rating");
    const people = document.getElementById("place-people");
    const category = document.getElementById("place-category");
    const icon = document.querySelector(".place-icon");
    const photo = document.getElementById("place-image");
    const addressEl = document.getElementById("place-address");
    const navigate = document.getElementById("place-navigate");

    if (title) title.textContent = place.title || "Activité";
    if (description) description.textContent = place.description || "Pas de description";
    if (city) city.textContent = place.city || "France";
    if (rating) rating.textContent = "⭐ " + (place.rating || "0.0");
    if (people) people.textContent = "👥 " + (place.people || "0") + " personnes";
    if (category) category.textContent = meta.label;
    if (icon) {
        icon.textContent = meta.icon;
        icon.style.color = meta.color;
    }
    if (addressEl) addressEl.textContent = (place.address || place.city || "France") + distanceText;

    // Photo
    let photoUrl = place.photo;
    if (!photoUrl || photoUrl.trim() === "") {
        photoUrl = `https://picsum.photos/seed/${encodeURIComponent((place.title || "activity") + (place.city || "france"))}/600/280`;
    }

    if (photo) {
        photo.hidden = false;
        photo.style.backgroundImage = `url("${photoUrl}")`;
    }

    // Itinéraire
    if (navigate) {
        if (place.lat && place.lng) {
            navigate.href = `https://www.google.com/maps/dir/?api=1&destination=${place.lat},${place.lng}`;
            navigate.hidden = false;
        } else {
            navigate.hidden = true;
        }
    }
}

// === GOOGLE PLACES AUTOCOMPLETE ===

function initAutocomplete() {
    const input = document.getElementById("activity-city");
    if (!input) return;

    autocomplete = new google.maps.places.Autocomplete(input, {
        componentRestrictions: { country: ["fr"] },
        types: ["geocode", "establishment"]
    });

    autocomplete.addListener("place_changed", () => {
        const place = autocomplete.getPlace();
        if (place.geometry) {
            const lat = document.getElementById("activity-lat");
            const lng = document.getElementById("activity-lng");
            if (lat) lat.value = place.geometry.location.lat().toFixed(5);
            if (lng) lng.value = place.geometry.location.lng().toFixed(5);
            
            // Centrer la map sur le lieu saisi
            map.panTo(place.geometry.location);
            map.setZoom(13);
        }
    });
}

// === FORMULAIRE ===

function bindForm() {
    const form = document.getElementById("activity-form");
    if (!form) return;

    form.addEventListener("submit", (e) => {
        e.preventDefault();
        submitActivityForm();
    });

    const useLocation = document.getElementById("use-location-button");
    if (useLocation) {
        useLocation.addEventListener("click", useMyLocation);
    }
}

function useMyLocation() {
    if (!navigator.geolocation) {
        alert("Géolocalisation non supportée.");
        return;
    }
    navigator.geolocation.getCurrentPosition(
        (position) => {
            const lat = document.getElementById("activity-lat");
            const lng = document.getElementById("activity-lng");
            if (lat) lat.value = position.coords.latitude.toFixed(5);
            if (lng) lng.value = position.coords.longitude.toFixed(5);
        },
        () => {
            alert("Impossible de récupérer la position.");
        }
    );
}

function submitActivityForm() {
    const title = document.getElementById("activity-title").value.trim();
    const city = document.getElementById("activity-city").value.trim();
    const category = document.getElementById("activity-category").value;
    const mood = document.getElementById("activity-mood").value;
    const description = document.getElementById("activity-description").value.trim();
    const lat = parseFloat(document.getElementById("activity-lat").value);
    const lng = parseFloat(document.getElementById("activity-lng").value);
    const maxPlaces = parseInt(document.getElementById("activity-max-places").value, 10) || 1;
    const photoInput = document.getElementById("activity-photo");

    if (!title || !city || !category || !mood || !lat || !lng) {
        alert("Merci de remplir tous les champs obligatoires.");
        return;
    }

    const payload = {
        title: title,
        city: city,
        category: category,
        mood: mood,
        lat: lat,
        lng: lng,
        description: description,
        maxPlaces: maxPlaces,
        photo: "",
    };

    if (photoInput && photoInput.files && photoInput.files[0]) {
        const reader = new FileReader();
        reader.onload = () => {
            payload.photo = reader.result;
            sendActivity(payload);
        };
        reader.readAsDataURL(photoInput.files[0]);
    } else {
        sendActivity(payload);
    }
}

function sendActivity(payload) {
    fetch("/api/activities", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || "Erreur serveur");
                });
            }
            return response.json();
        })
        .then(activity => {
            places.unshift(activity);
            activitiesCache[activity.id] = activity;
            updateCounts();
            showPlaces();
            updateCard(activity);
            map.panTo(new google.maps.LatLng(activity.lat, activity.lng));
            map.setZoom(13);
            document.getElementById("activity-form").reset();
            alert("✅ Activité ajoutée avec succès !");
        })
        .catch(error => {
            alert("❌ Erreur : " + error.message);
        });
}

// === FILTRES ===

function bindFilterButtons() {
    document.querySelectorAll(".category-button").forEach(button => {
        button.addEventListener("click", () => {
            selectedCategory = button.getAttribute("data-category") || "all";
            document.querySelectorAll(".category-button").forEach(btn => {
                btn.classList.toggle("active", btn === button);
            });
            showPlaces();
        });
    });

    document.querySelectorAll(".mood-button").forEach(button => {
        button.addEventListener("click", () => {
            selectedMood = button.getAttribute("data-mood") || "all";
            document.querySelectorAll(".mood-button").forEach(btn => {
                btn.classList.toggle("active", btn === button);
            });
            showPlaces();
        });
    });

    const searchInput = document.getElementById("search-input");
    if (searchInput) {
        searchInput.addEventListener("input", () => {
            searchText = searchInput.value || "";
            showPlaces();
        });
    }
}

// === PARTAGE ===

function shareActivity(activityId) {
    const activity = activitiesCache[activityId];
    if (!activity) return;
    
    const shareText = `Je découvre "${activity.title}" sur Zukuk - une activité bien-être en France ! 🌟`;
    const shareUrl = window.location.href;
    
    if (navigator.share) {
        navigator.share({
            title: activity.title,
            text: shareText,
            url: shareUrl
        }).catch(err => console.log('Erreur partage:', err));
    } else {
        // Fallback: copier dans le presse-papier
        const fullText = `${shareText}\n${shareUrl}`;
        navigator.clipboard.writeText(fullText).then(() => {
            alert("📋 Lien copié dans le presse-papier !");
        });
    }
}

// === ÉVÉNEMENT DE CHARGEMENT ===

window.addEventListener("load", initMapPage);
