const API_BASE = '/api';

let map, markers = [];
let activities = [];
let currentUser = null;
let selectedActivity = null;

const savedTheme = localStorage.getItem('zukuk_theme') || 'light';
document.body.setAttribute('data-theme', savedTheme);

const qs = (s) => document.querySelector(s);

// ── TOAST UNIVERSEL ────────────────────────────────────────────────────────
function showToast(msg, isError = false) {
    const existing = document.getElementById('map-toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.id = 'map-toast';
    toast.textContent = msg;
    toast.style.cssText = `position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%) translateY(100px); background: ${isError ? '#fef2f2' : '#f0fdf4'}; color: ${isError ? '#ef4444' : '#166534'}; border: 1px solid ${isError ? '#fecaca' : '#bbf7d0'}; padding: 14px 28px; border-radius: 16px; font-weight: 600; font-size: 0.95rem; box-shadow: 0 10px 30px rgba(0,0,0,0.1); z-index: 9999; transition: transform 0.4s cubic-bezier(0.2, 0.9, 0.2, 1);`;
    document.body.appendChild(toast);
    
    requestAnimationFrame(() => toast.style.transform = 'translateX(-50%) translateY(0)');
    setTimeout(() => {
        toast.style.transform = 'translateX(-50%) translateY(100px)';
        setTimeout(() => toast.remove(), 400);
    }, 3500);
}

// ── MODALE DE CONFIRMATION ULTRA MODERNE ───────────────────────────────────
function showConfirmModal(title, message, onConfirm) {
    // Supprimer si elle existe déjà
    const existing = document.getElementById('zukuk-confirm');
    if (existing) existing.remove();

    // Création de l'overlay (fond sombre flouté)
    const overlay = document.createElement('div');
    overlay.id = 'zukuk-confirm';
    overlay.style.cssText = `
        position: fixed; inset: 0; background: rgba(15, 23, 42, 0.4); backdrop-filter: blur(8px);
        z-index: 10000; display: flex; align-items: center; justify-content: center; padding: 20px;
        opacity: 0; transition: opacity 0.3s ease;
    `;

    // Création de la boîte de dialogue
    const modal = document.createElement('div');
    modal.style.cssText = `
        background: white; border-radius: 24px; padding: 32px; width: 100%; max-width: 400px;
        box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25); transform: translateY(20px) scale(0.95);
        transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1); text-align: center;
    `;

    modal.innerHTML = `
        <div style="width: 64px; height: 64px; background: #fef2f2; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px;">
            <svg viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" style="width: 32px; height: 32px;">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
        </div>
        <h3 style="font-size: 1.25rem; font-weight: 700; color: #0f172a; margin-bottom: 12px; font-family: 'Merriweather', serif;">${title}</h3>
        <p style="color: #64748b; font-size: 0.95rem; margin-bottom: 28px; line-height: 1.5;">${message}</p>
        <div style="display: flex; gap: 12px;">
            <button id="cancel-confirm" style="flex: 1; padding: 12px; border: 2px solid #e2e8f0; border-radius: 14px; background: white; color: #475569; font-weight: 600; cursor: pointer; transition: all 0.2s;">Annuler</button>
            <button id="accept-confirm" style="flex: 1; padding: 12px; border: none; border-radius: 14px; background: #ef4444; color: white; font-weight: 600; cursor: pointer; box-shadow: 0 4px 14px rgba(239, 68, 68, 0.3); transition: all 0.2s;">Supprimer</button>
        </div>
    `;

    overlay.appendChild(modal);
    document.body.appendChild(overlay);

    // Animation d'entrée
    requestAnimationFrame(() => {
        overlay.style.opacity = '1';
        modal.style.transform = 'translateY(0) scale(1)';
    });

    // Fonction de fermeture
    const close = () => {
        overlay.style.opacity = '0';
        modal.style.transform = 'translateY(20px) scale(0.95)';
        setTimeout(() => overlay.remove(), 300);
    };

    // Événements
    document.getElementById('cancel-confirm').onclick = close;
    overlay.onclick = (e) => { if (e.target === overlay) close(); };
    
    document.getElementById('accept-confirm').onclick = () => {
        close();
        onConfirm(); // Exécute la fonction de suppression
    };

    // Effets de survol (Hover)
    document.getElementById('cancel-confirm').onmouseover = function() { this.style.background = '#f8fafc'; };
    document.getElementById('cancel-confirm').onmouseout = function() { this.style.background = 'white'; };
    document.getElementById('accept-confirm').onmouseover = function() { this.style.background = '#dc2626'; };
    document.getElementById('accept-confirm').onmouseout = function() { this.style.background = '#ef4444'; };
}


// ── STICKERS EMOJIS ────────────────────────────────────────────────────────
function getMarkerIcon(categoryId) {
    const emojis = { 1: '⚽', 2: '🧘', 3: '💬', 4: '🎨', 6: '🌿' };
    const emoji = emojis[categoryId] || '📍';
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 36 36">
        <circle cx="18" cy="18" r="16" fill="#ffffff" stroke="#818cf8" stroke-width="2.5"/>
        <text x="18" y="23" font-size="18" text-anchor="middle" font-family="sans-serif">${emoji}</text>
    </svg>`;
    return 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(svg);
}

// ── LANCEMENT ──────────────────────────────────────────────────────────────
async function bootstrap() {
    try {
        const res = await fetch(`${API_BASE}/me`, { credentials: 'include' });
        if (res.ok) currentUser = await res.json();
    } catch (e) { console.warn("Mode Invité"); }

    if (currentUser && qs('#created_by')) qs('#created_by').value = currentUser.id;

    try {
        flatpickr("#schedule", {
            enableTime: true, 
            time_24hr: true, 
            locale: typeof flatpickr.l10ns.fr !== 'undefined' ? "fr" : "default",
            dateFormat: "Y-m-d\\TH:i", 
            altInput: true, 
            altFormat: "l j F Y à H:i", 
            minDate: "today"
        });
    } catch (err) {
        console.error("Erreur chargement calendrier :", err);
    }

    let attempts = 0;
    const checkGoogle = () => {
        attempts++;
        if (window.google && google.maps && google.maps.places) {
            init();
        } else if (attempts < 50) {
            setTimeout(checkGoogle, 100);
        } else {
            console.error("Google Maps API n'a pas pu se charger.");
            showToast("Erreur de chargement de la carte (Google Maps).", true);
        }
    };
    checkGoogle();
}

function init() {
    map = new google.maps.Map(qs('#map'), {
        center: { lat: 45.7640, lng: 4.8357 }, zoom: 13,
        styles: [{ featureType: "poi", stylers: [{ visibility: "off" }] }],
        disableDefaultUI: true, zoomControl: true
    });

    const card = qs('#place-card');
    if (card) card.style.pointerEvents = 'none';

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
            const data = await res.json();
            activities = data || []; 
            renderMarkers();
        }
    } catch (err) { console.error(err); }
}

function renderMarkers() {
    markers.forEach(m => m.setMap(null));
    if (!Array.isArray(activities)) return;
    
    markers = activities.map(act => {
        const m = new google.maps.Marker({
            position: { lat: act.latitude, lng: act.longitude }, map,
            icon: { url: getMarkerIcon(act.category_id), scaledSize: new google.maps.Size(36, 36) }
        });
        m.addListener('click', () => showCard(act));
        return m;
    });
}

async function showCard(act) {
    selectedActivity = act;
    
    qs('#pc-title').textContent = act.name || 'Sans titre';
    qs('#pc-address').textContent = `📍 ${act.address || 'Adresse inconnue'}`;
    qs('#pc-description').textContent = act.description || '';

    try {
        const safeDateStr = act.schedule ? act.schedule.replace(' ', 'T') : '';
        const d = new Date(safeDateStr);
        qs('#pc-schedule').textContent = `📅 ${d.toLocaleDateString('fr-FR', {weekday: 'long', day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'})}`;
    } catch (e) { qs('#pc-schedule').textContent = `📅 Date non définie`; }

    const isFull = act.participants_count >= act.max_places;
    qs('#pc-participants').textContent = `👥 ${act.participants_count || 0} / ${act.max_places} inscrits`;

    const listContainer = qs('#pc-participants-list');
    const joinBtn = qs('#join-btn');
    const deleteBtn = qs('#delete-act-btn');

    if (currentUser && currentUser.id === act.created_by) {
        if (deleteBtn) deleteBtn.style.display = 'block';
        if (joinBtn) joinBtn.style.display = 'none';

        if (listContainer) {
            listContainer.innerHTML = '<span style="font-size:0.8rem; color:#94a3b8">Chargement...</span>';
            try {
                const res = await fetch(`${API_BASE}/activities/${act.id}/participants`, { credentials: 'include' });
                if (res.ok) {
                    const participants = await res.json();
                    listContainer.innerHTML = ''; 
                    if (participants.length === 0) {
                        listContainer.innerHTML = '<span style="font-size:0.8rem; color:#94a3b8; font-style:italic">Aucun inscrit pour le moment.</span>';
                    } else {
                        participants.forEach(p => {
                            const imgUrl = p.avatar_url ? (p.avatar_url.startsWith('http') ? p.avatar_url : `http://localhost:8081${p.avatar_url}`) : `https://api.dicebear.com/7.x/notionists/svg?seed=${p.pseudo}`;
                            listContainer.innerHTML += `<div title="${p.pseudo}" style="width:32px; height:32px; border-radius:50%; border:2px solid white; box-shadow:0 2px 5px rgba(0,0,0,0.1); overflow:hidden; background:#f1f5f9"><img src="${imgUrl}" style="width:100%; height:100%; object-fit:cover"></div>`;
                        });
                    }
                }
            } catch (err) { listContainer.innerHTML = ''; }
        }
    } else {
        if (deleteBtn) deleteBtn.style.display = 'none';
        if (listContainer) {
            listContainer.innerHTML = `<div style="display:flex; align-items:center; gap:8px; background:#f1f5f9; padding:8px 12px; border-radius:12px; width:100%;"><span>🔒</span><span style="font-size:0.8rem; color:#64748b;">Seul l'organisateur peut voir les inscrits</span></div>`;
        }

        if (joinBtn) {
            joinBtn.style.display = 'block';
            if (act.is_participating) {
                joinBtn.textContent = "Se désinscrire";
                joinBtn.style.background = "#fef2f2"; joinBtn.style.color = "#ef4444"; joinBtn.disabled = false;
            } else if (isFull) {
                joinBtn.textContent = "Complet 🛑";
                joinBtn.style.background = "#e2e8f0"; joinBtn.style.color = "#64748b"; joinBtn.disabled = true;
            } else {
                joinBtn.textContent = "S'inscrire";
                joinBtn.style.background = ""; joinBtn.style.color = ""; joinBtn.disabled = false;
            }
        }
    }

    const card = qs('#place-card');
    if (card) { card.classList.add('visible'); card.style.pointerEvents = 'auto'; }
    map.panTo({ lat: act.latitude, lng: act.longitude });
}

function bindEvents() {
    const form = qs('#activity-form');
    
    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            if (!currentUser) {
                showToast("🔒 Tu dois te connecter pour publier !", true);
                setTimeout(() => window.location.href = "/login", 2000);
                return;
            }

            const formData = new FormData(e.target);
            const payload = Object.fromEntries(formData.entries());

            const fp = qs('#schedule')._flatpickr;
            let rawDate = "";
            if (fp && fp.selectedDates.length > 0) rawDate = fp.formatDate(fp.selectedDates[0], "Y-m-d\\TH:i"); 

            if (!rawDate) { showToast("📅 N'oublie pas de choisir une date.", true); return; }
            payload.schedule = rawDate;
            payload.category_id = parseInt(payload.category_id);
            payload.latitude = parseFloat(payload.latitude);
            payload.longitude = parseFloat(payload.longitude);
            payload.max_places = parseInt(payload.max_places);
            if (payload.created_by) payload.created_by = parseInt(payload.created_by);

            if (isNaN(payload.latitude) || isNaN(payload.longitude)) {
                showToast("📍 Clique bien sur l'adresse suggérée par Google.", true);
                return;
            }

            const btn = qs('#create-activity');
            const originalText = btn.textContent;
            btn.textContent = "Publication..."; btn.disabled = true;

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
                    showToast("✨ Activité publiée avec succès sur la carte !", false);
                } else {
                    const errorData = await res.json();
                    showToast(errorData.error || "Erreur du serveur.", true);
                }
            } catch (error) { showToast("Serveur injoignable.", true); } 
            finally { btn.textContent = originalText; btn.disabled = false; }
        });
    }

    qs('#close-card').onclick = () => {
        qs('#place-card').classList.remove('visible');
        qs('#place-card').style.pointerEvents = 'none'; 
        selectedActivity = null;
    };

    qs('#join-btn').onclick = async () => {
        if (!currentUser) {
            showToast("Tu dois te connecter pour participer.", true);
            setTimeout(() => window.location.href = "/login", 1500); return;
        }
        if (!selectedActivity) return;

        const joinBtn = qs('#join-btn');
        const originalText = joinBtn.textContent;
        joinBtn.disabled = true; joinBtn.textContent = "Chargement...";

        try {
            const res = await fetch(`${API_BASE}/activities/${selectedActivity.id}/join`, {
                method: 'POST', credentials: 'include'
            });

            if (res.ok) {
                const data = await res.json();
                if (data.joined) showToast("✨ Parfait ! Tu es inscrit.", false);
                else showToast("ℹ️ Tu t'es désinscrit.", false);
                
                selectedActivity.is_participating = data.joined;
                selectedActivity.participants_count = data.count;
                showCard(selectedActivity); 

            } else { showToast("Action impossible.", true); joinBtn.textContent = originalText; }
        } catch (err) { showToast("Erreur de connexion.", true); joinBtn.textContent = originalText; } 
        finally { joinBtn.disabled = false; }
    };

    // 🔴 L'ANCIENNE ALERTE EST REMPLACÉE ICI
    qs('#delete-act-btn').onclick = () => {
        showConfirmModal(
            "Supprimer l'activité ?",
            "Cette action est irréversible. L'événement sera annulé et retiré de la carte pour tous les participants.",
            async () => {
                const deleteBtn = qs('#delete-act-btn');
                deleteBtn.disabled = true; deleteBtn.textContent = "⏳";

                try {
                    const res = await fetch(`${API_BASE}/activities/${selectedActivity.id}`, {
                        method: 'DELETE', credentials: 'include'
                    });
                    if (res.ok) {
                        showToast("🗑️ Activité supprimée avec succès.", false);
                        qs('#place-card').classList.remove('visible');
                        await loadActivities(); 
                    } else { showToast("Erreur lors de la suppression.", true); }
                } catch (err) { showToast("Erreur serveur.", true); } 
                finally { deleteBtn.disabled = false; deleteBtn.textContent = "🗑️"; }
            }
        );
    };
}

// 🔥 EXÉCUTION GARANTIE UNE FOIS LE HTML CHARGÉ
document.addEventListener('DOMContentLoaded', bootstrap);