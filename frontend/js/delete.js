const API_BASE = window.location.hostname === 'localhost' 
    ? 'http://localhost:8081/api' 
    : '/api';// REMPLACE PAR TON VRAI LIEN RENDER
const fetchOpts = { credentials: 'include' };

// ─────────────────────────────────────────────────────────────────────────────
// 🎨 SYSTÈME DE THÈME GLOBAL ZUKUK
// ─────────────────────────────────────────────────────────────────────────────

// 1. Anti-scintillement : Appliquer le thème local IMMÉDIATEMENT
const savedTheme = localStorage.getItem('zukuk_theme') || 'light';
document.body.setAttribute('data-theme', savedTheme);

// 2. Synchronisation BDD : Vérifier le vrai thème de l'utilisateur au chargement
document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch(`${API_BASE}/settings/`, { credentials: 'include' });
        if (res.ok) {
            const settings = await res.json();
            if (settings.theme && settings.theme !== savedTheme) {
                // Si la BDD a un thème différent (ex: connexion sur un nouveau PC), on met à jour
                document.body.setAttribute('data-theme', settings.theme);
                localStorage.setItem('zukuk_theme', settings.theme);
            }
        }
    } catch (e) {
        // Utilisateur non connecté, on garde le thème local
    }
});

let selectedDays = 30;
let currentPseudo = '';

// ── Charger le pseudo de l'utilisateur ──────────────────────────────
(async () => {
    try {
    const res = await fetch(`${API_BASE}/me`, fetchOpts);
    if (!res.ok) { window.location.href = '/login'; return; }
    const user = await res.json();
    currentPseudo = user.pseudo;
    document.getElementById('pseudo-hint').textContent = `"${currentPseudo}"`;
    } catch (_) {
    window.location.href = '/login';
    }
})();

// ── Duration chips ───────────────────────────────────────────────────
document.querySelectorAll('.duration-chip').forEach(chip => {
    chip.addEventListener('click', () => {
    document.querySelectorAll('.duration-chip').forEach(c => c.classList.remove('active'));
    chip.classList.add('active');
    selectedDays = parseInt(chip.dataset.days, 10);
    });
});

// ── Pause button ─────────────────────────────────────────────────────
document.getElementById('pause-btn').addEventListener('click', () => {
    const durText = selectedDays >= 180
    ? '6 mois' : selectedDays >= 90
    ? '90 jours' : selectedDays >= 30
    ? '30 jours' : '7 jours';
    document.getElementById('overlay-duration').textContent = durText;
    showOverlay('pause-overlay');
});

document.getElementById('confirm-pause-btn').addEventListener('click', async () => {
    const btn = document.getElementById('pause-btn');
    setLoading(btn, true);
    closeOverlay('pause-overlay');

    try {
    const res = await fetch(`${API_BASE}/me/pause`, {
        method: 'POST', credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ duration_days: selectedDays }),
    });
    const data = await res.json();
    if (!res.ok) { showAlert('pause-alert', 'error', data.error || 'Erreur serveur.'); return; }

    showAlert('pause-alert', 'success',
        `Ta pause de ${selectedDays} jours est activée. À bientôt ! ☕`);
    btn.disabled = true;
    setTimeout(() => { window.location.href = '/login'; }, 2000);
    } catch (_) {
    showAlert('pause-alert', 'error', 'Erreur réseau. Réessaie dans un moment.');
    } finally {
    setLoading(btn, false);
    }
});

// ── Confirm input → enable delete button ────────────────────────────
const confirmInput = document.getElementById('confirm-input');
const deleteBtn    = document.getElementById('delete-btn');

confirmInput.addEventListener('input', () => {
    const match = confirmInput.value.trim() === currentPseudo && currentPseudo !== '';
    deleteBtn.disabled = !match;
    confirmInput.classList.toggle('valid', match);
});

// ── Delete button ────────────────────────────────────────────────────
deleteBtn.addEventListener('click', async () => {
    if (confirmInput.value.trim() !== currentPseudo) return;
    setLoading(deleteBtn, true);

    try {
    const res = await fetch(`${API_BASE}/me`, {
        method: 'DELETE', credentials: 'include',
    });
    const data = await res.json();
    if (!res.ok) { showAlert('delete-alert', 'error', data.error || 'Erreur serveur.'); return; }

    showAlert('delete-alert', 'success',
        'Ton compte est masqué. Il sera supprimé définitivement dans 30 jours. Prends soin de toi 💙');
    deleteBtn.disabled = true;
    setTimeout(() => { window.location.href = '/'; }, 3000);
    } catch (_) {
    showAlert('delete-alert', 'error', 'Erreur réseau. Réessaie dans un moment.');
    } finally {
    setLoading(deleteBtn, false);
    }
});

// ── Helpers ──────────────────────────────────────────────────────────
function showOverlay(id) {
    document.getElementById(id).classList.add('show');
}
function closeOverlay(id) {
    document.getElementById(id).classList.remove('show');
}
document.querySelectorAll('.overlay').forEach(o => {
    o.addEventListener('click', e => { if (e.target === o) o.classList.remove('show'); });
});

function showAlert(id, type, text) {
    const el = document.getElementById(id);
    el.className = `alert show ${type}`;
    el.querySelector('span').textContent = text;
}

function setLoading(btn, loading) {
    btn.classList.toggle('loading', loading);
    btn.disabled = loading;
}