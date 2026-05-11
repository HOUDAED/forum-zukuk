const API_BASE = 'http://localhost:8081/api';
let map, markers = [];
let activities = [];
let currentUser = null;
let selectedActivity = null;

const qs = (s) => document.querySelector(s);

// ── Fonction Toast (Remplace les Alert) ──
function showToast(msg, isError = false) {
    const existing = document.getElementById('map-toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.id = 'map-toast';
    toast.textContent = msg;
    toast.style.cssText = `
        position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%) translateY(100px);
        background: ${isError ? '#fef2f2' : '#f0fdf4'};
        color: ${isError ? '#ef4444' : '#166534'};
        border: 1px solid ${isError ? '#fecaca' : '#bbf7d0'};
        padding: 14px 28px; border-radius: 16px; font-weight: 600; font-size: 0.95rem;
        box-shadow: 0 10px 30px rgba(0,0,0,0.1); z-index: 9999;
        transition: transform 0.4s cubic-bezier(0.2, 0.9, 0.2, 1);
    `;
    document.body.appendChild(toast);
    
    requestAnimationFrame(() => toast.style.transform = 'translateX(-50%) translateY(0)');
    
    setTimeout(() => {
        toast.style.transform = 'translateX(-50%) translateY(100px)';
        setTimeout(() => toast.remove(), 400);
    }, 3500);
}

async function bootstrap() {
    try {
        const res = await fetch(`${API_BASE}/me`, { credentials: 'include' });
        if (res.ok) currentUser = await res.json();
    } catch (e) {
        console.warn("Mode invité actif");
    }

    if (currentUser && qs('#created_by')) qs('#created_by').value = currentUser.id;

    flatpickr("#schedule", {
        enableTime: true, time_24hr: true, locale: "fr",
        dateFormat: "Y-m-d\\TH:i", altInput: true, altFormat: "l j F Y à H:i", minDate: "today"
    });

    const check = () => (window.google && google.maps && google.maps.places) ? init() : setTimeout(check, 100);
    check();
}

function init() {
    map = new google.maps.Map(qs('#map'), {
        center: { lat: 45.7640, lng: 4.8357 }, zoom: 13,
        styles: [{ featureType: "poi", stylers: [{ visibility: "off" }] }],
        disableDefaultUI: true, zoomControl: true
    });

    const card = qs('#place-card');
    if (card) card.style.pointerEvents = 'none'; // Fix fantôme

    initAutocomplete();
    bindEvents();
    loadActivities();
}

function initAutocomplete() {
    const addressInput = qs('#address');
    if (!addressInput) return;
    const ac = new google.maps.places.Autocomplete(addressInput, { componentRestrictions: { country: 'fr' } });
    ac.addListener('place_changed', () => {
        const p = ac.getPlace();
        if (!p.geometry) return;
        qs('#latitude').value = p.geometry.location.lat();
        qs('#longitude').value = p.geometry.location.lng();
        map.panTo(p.geometry.location);
    });
}

async function loadActivities() {
    try {
        const res = await fetch(`${API_BASE}/activities`, { credentials: 'include' });
        if (res.ok) {
            activities = await res.json();
            renderMarkers();
        }
    } catch (err) {
        console.error(err);
    }
}

function renderMarkers() {
    markers.forEach(m => m.setMap(null));
    markers = activities.map(act => {
        const m = new google.maps.Marker({
            position: { lat: act.latitude, lng: act.longitude }, map,
            icon: { path: google.maps.SymbolPath.CIRCLE, fillColor: '#818cf8', fillOpacity: 1, strokeWeight: 2, strokeColor: '#fff', scale: 10 }
        });
        m.addListener('click', () => showCard(act));
        return m;
    });
}

function showCard(act) {
    selectedActivity = act;
    
    // Protection anti-crash
    if(qs('#pc-title')) qs('#pc-title').textContent = act.name || 'Sans titre';
    if(qs('#pc-address')) qs('#pc-address').textContent = `📍 ${act.address || 'Adresse inconnue'}`;
    if(qs('#pc-description')) qs('#pc-description').textContent = act.description || '';

    if(qs('#pc-schedule')) {
        try {
            const safeDateStr = act.schedule ? act.schedule.replace(' ', 'T') : '';
            const d = new Date(safeDateStr);
            if (isNaN(d.getTime())) throw new Error();
            qs('#pc-schedule').textContent = `📅 ${d.toLocaleDateString('fr-FR', {weekday: 'long', day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'})}`;
        } catch (e) {
            qs('#pc-schedule').textContent = `📅 Date non définie`;
        }
    }

    if(qs('#pc-participants')) {
        qs('#pc-participants').textContent = `👥 ${act.participants_count || 0} / ${act.max_places || 0} inscrits`;
    }

    const joinBtn = qs('#join-btn');
    if (joinBtn) {
        // 🔴 FIX MAJEUR : Utilisation de "is_participating" envoyé par le Go
        if (act.is_participating) {
            joinBtn.textContent = "Se désinscrire";
            joinBtn.style.background = "#fef2f2";
            joinBtn.style.color = "#ef4444";
        } else {
            joinBtn.textContent = "S'inscrire";
            joinBtn.style.background = "";
            joinBtn.style.color = "";
        }
    }

    const card = qs('#place-card');
    if (card) {
        card.classList.add('visible');
        card.style.pointerEvents = 'auto'; 
    }
    map.panTo({ lat: act.latitude, lng: act.longitude });
}

function bindEvents() {
    const form = qs('#activity-form');
    let msgDiv = qs('#form-msg');
    
    const showMessage = (text, isError) => {
        if(!msgDiv) return;
        msgDiv.style.display = 'block';
        msgDiv.textContent = text;
        msgDiv.style.color = isError ? '#ef4444' : '#166534';
        msgDiv.style.backgroundColor = isError ? '#fef2f2' : '#f0fdf4';
        msgDiv.style.padding = '12px';
        msgDiv.style.marginTop = '12px';
        msgDiv.style.borderRadius = '12px';
        msgDiv.style.border = `1px solid ${isError ? '#fecaca' : '#bbf7d0'}`;
        msgDiv.style.fontWeight = '600';
        msgDiv.style.fontSize = '0.9rem';
    };

    if (form) {
        const dateInput = qs('#schedule');
        if (dateInput) dateInput.removeAttribute('required'); 

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (msgDiv) msgDiv.style.display = 'none';

            if (!currentUser) {
                showMessage("🔒 Tu dois te connecter pour publier !", true);
                setTimeout(() => window.location.href = "/login", 2000);
                return;
            }

            const formData = new FormData(e.target);
            const payload = Object.fromEntries(formData.entries());

            // 🔴 FIX MAJEUR DE LA DATE : On interroge directement Flatpickr
            const fp = qs('#schedule')._flatpickr;
            let rawDate = "";
            if (fp && fp.selectedDates.length > 0) {
                // On force le format informatique exact requis par la base de données
                rawDate = fp.formatDate(fp.selectedDates[0], "Y-m-d\\TH:i"); 
            }

            if (!rawDate) {
                showMessage("📅 N'oublie pas de choisir une date.", true);
                return;
            }
            
            payload.schedule = rawDate;
            payload.category_id = parseInt(payload.category_id);
            payload.latitude = parseFloat(payload.latitude);
            payload.longitude = parseFloat(payload.longitude);
            payload.max_places = parseInt(payload.max_places);
            if (payload.created_by) payload.created_by = parseInt(payload.created_by);

            if (isNaN(payload.latitude) || isNaN(payload.longitude)) {
                showMessage("📍 Clique bien sur l'adresse suggérée par Google.", true);
                return;
            }

            const btn = qs('#create-activity');
            const originalText = btn.textContent;
            btn.textContent = "Publication...";
            btn.disabled = true;

            try {
                const res = await fetch(`${API_BASE}/activities`, {
                    method: 'POST', credentials: 'include',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if(res.ok) {
                    e.target.reset();
                    if(qs("#schedule")._flatpickr) qs("#schedule")._flatpickr.clear();
                    await loadActivities();
                    showMessage("✨ Activité publiée avec succès sur la carte !", false);
                } else {
                    const errData = await res.json().catch(() => ({}));
                    showMessage(errData.error || "Erreur du serveur (401). Tu es peut-être déconnecté.", true);
                }
            } catch (error) {
                showMessage("Serveur injoignable.", true);
            } finally {
                // On remet le bouton à son état normal quoiqu'il arrive
                btn.textContent = originalText;
                btn.disabled = false;
            }
        });
    }

    const closeBtn = qs('#close-card');
    if (closeBtn) {
        closeBtn.onclick = () => {
            const card = qs('#place-card');
            card.classList.remove('visible');
            card.style.pointerEvents = 'none'; 
            selectedActivity = null;
        };
    }

    const joinBtn = qs('#join-btn');
    if (joinBtn) {
        joinBtn.onclick = async () => {
            if (!currentUser) {
                showToast("Tu dois te connecter pour participer.", true);
                setTimeout(() => window.location.href = "/login", 1500);
                return;
            }
            if (!selectedActivity) return;

            // On sauvegarde le texte pour éviter qu'il reste sur "Chargement..."
            const originalText = joinBtn.textContent;
            joinBtn.disabled = true;
            joinBtn.textContent = "Chargement...";

            try {
                const res = await fetch(`${API_BASE}/activities/${selectedActivity.id}/join`, {
                    method: 'POST', credentials: 'include'
                });

                if (res.ok) {
                    showToast("Participation mise à jour !", false);
                    await loadActivities();
                    const updatedAct = activities.find(a => a.id === selectedActivity.id);
                    if (updatedAct) showCard(updatedAct);
                } else {
                    const errData = await res.json().catch(() => ({}));
                    showToast(errData.error || "Action impossible. (Session expirée ?)", true);
                    joinBtn.textContent = originalText; // On restaure si échec
                }
            } catch (err) {
                showToast("Erreur de connexion.", true);
                joinBtn.textContent = originalText; // On restaure si crash
            } finally {
                joinBtn.disabled = false;
            }
        };
    }
}

bootstrap();