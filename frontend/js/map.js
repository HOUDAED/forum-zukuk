const API_BASE = 'http://localhost:8081/api';
let map, markerCluster, markers = [];
let activities = [];
let currentUser = null;
let selectedActivity = null;

const qs = (s) => document.querySelector(s);

async function bootstrap() {
  currentUser = await fetch(`${API_BASE}/me`, { credentials: 'include' }).then(r => r.ok ? r.json() : null);
  if (currentUser) qs('#created_by').value = currentUser.id;

  // Activation du Calendrier Flatpickr Ultra-Moderne
  flatpickr("#schedule", {
    enableTime: true,
    time_24hr: true,
    locale: "fr", // Langue française (Lun, Mar, Mer...)
    dateFormat: "Y-m-d\\TH:i", // Format brut pour ta BDD (ex: 2026-05-11T18:00)
    altInput: true, // Crée un 2ème champ virtuel juste pour l'affichage visuel
    altFormat: "l j F Y à H:i", // Format humain (ex: Lundi 11 mai 2026 à 18:00)
    minDate: "today", // Empêche de créer dans le passé
    disableMobile: "true" // Force le beau calendrier même sur téléphone
  });

  // Attente Google Maps
  const check = () => (window.google && google.maps && google.maps.places) ? init() : setTimeout(check, 100);
  check();
}

function init() {
  map = new google.maps.Map(qs('#map'), {
    center: { lat: 45.7640, lng: 4.8357 }, zoom: 13,
    styles: [{ featureType: "poi", stylers: [{ visibility: "off" }] }],
    disableDefaultUI: true, zoomControl: true
  });

  initAutocomplete();
  bindForm();
  loadActivities();
}

function initAutocomplete() {
  const ac = new google.maps.places.Autocomplete(qs('#address'), { componentRestrictions: { country: 'fr' } });
  ac.addListener('place_changed', () => {
    const p = ac.getPlace();
    if (!p.geometry) return;
    qs('#latitude').value = p.geometry.location.lat();
    qs('#longitude').value = p.geometry.location.lng();
    map.panTo(p.geometry.location);
  });
}

async function loadActivities() {
  const res = await fetch(`${API_BASE}/activities`, { credentials: 'include' });
  activities = await res.json();
  renderMarkers();
}

function renderMarkers() {
  markers.forEach(m => m.setMap(null));
  markers = activities.map(act => {
    const m = new google.maps.Marker({
      position: { lat: act.latitude, lng: act.longitude },
      map, icon: { path: google.maps.SymbolPath.CIRCLE, fillColor: '#818cf8', fillOpacity: 1, strokeWeight: 2, strokeColor: '#fff', scale: 10 }
    });
    m.addListener('click', () => showCard(act));
    return m;
  });
}

function showCard(act) {
  selectedActivity = act;
  qs('#pc-title').textContent = act.name;
  qs('#pc-address').textContent = `📍 ${act.address}`;
  qs('#pc-description').textContent = act.description;
  
  if(act.schedule) {
    const d = new Date(act.schedule);
    qs('#pc-schedule').textContent = `📅 ${d.toLocaleDateString('fr-FR', {weekday: 'long', day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'})}`;
  }

  qs('#pc-participants').textContent = `👥 ${act.participants_count || 0} / ${act.max_places} inscrits`;
  
  const card = qs('#place-card');
  card.classList.add('visible');
  map.panTo({ lat: act.latitude, lng: act.longitude });
}

function bindForm() {
  qs('#activity-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const payload = Object.fromEntries(formData.entries());
    
    // Le vrai champ "schedule" caché contient le bon format YYYY-MM-DDTHH:MM pour la DB.
    // L'altInput (celui qu'on voit) s'appellera différemment ou n'est pas soumis.
    
    payload.category_id = parseInt(payload.category_id);
    payload.latitude = parseFloat(payload.latitude);
    payload.longitude = parseFloat(payload.longitude);
    payload.max_places = parseInt(payload.max_places);
    payload.created_by = parseInt(payload.created_by);

    const res = await fetch(`${API_BASE}/activities`, {
      method: 'POST', credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if(res.ok) {
      alert("Activité publiée !");
      loadActivities();
      e.target.reset();
    }
  });

  qs('#close-card').onclick = () => qs('#place-card').classList.remove('visible');
}

bootstrap();