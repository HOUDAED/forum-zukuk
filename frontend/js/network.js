// network.js — Page communauté
const API_BASE  = 'http://localhost:8081/api';
const fetchOpts = { credentials: 'include' };

let searchTimer = null;

// ─────────────────────────────────────────────────────────────────────────────
// 🎨 SYSTÈME DE THÈME GLOBAL ZUKUK
// ─────────────────────────────────────────────────────────────────────────────

// 1. Anti-scintillement : Appliquer le thème local IMMÉDIATEMENT
const savedTheme = localStorage.getItem('zukuk_theme') || 'light';
document.body.setAttribute('data-theme', savedTheme);

// 2. Synchronisation BDD : Vérifier le vrai thème de l'utilisateur au chargement
document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch('http://localhost:8081/api/settings/', { credentials: 'include' });
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

// ── Init ──────────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  loadMembers();
  setupListeners();
});

// ─────────────────────────────────────────────────────────────────────────────
// CHARGEMENT DES MEMBRES
// ─────────────────────────────────────────────────────────────────────────────
async function loadMembers() {
  const search = document.getElementById('networkSearch')?.value.trim() || '';
  const params = new URLSearchParams({ limit: 100 });
  if (search) params.set('search', search);

  try {
    const res = await fetch(`${API_BASE}/network?${params}`, fetchOpts);
    if (!res.ok) throw new Error('API error');
    const data = await res.json();
    renderMembers(data.members || []);
  } catch (err) {
    console.error(err);
    document.getElementById('membersList').innerHTML = `
      <div class="empty-state">
        <div class="emoji">😕</div>
        <p>Impossible de charger les membres.</p>
      </div>`;
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// RENDU
// ─────────────────────────────────────────────────────────────────────────────
const MOOD_EMOJIS = {
  'Bien': '😊', 'Calme': '😌', 'Triste': '😢', 'Anxieux': '😰', 'Colère': '😡',
};
const MOOD_CLASSES = {
  'Bien': 'mood-bien', 'Calme': 'mood-calme', 'Triste': 'mood-triste',
  'Anxieux': 'mood-anxieux', 'Colère': 'mood-colère',
};

const AVATAR_PALETTES = [
  ['#dbeafe','#1d4ed8'], ['#dcfce7','#166534'], ['#f3e8ff','#7c3aed'],
  ['#fef3c7','#92400e'], ['#fce7f3','#9d174d'], ['#e0f2fe','#0369a1'],
  ['#fce7f3','#be185d'], ['#fef9c3','#a16207'],
];

function avatarColors(name) {
  const idx = name ? name.charCodeAt(0) % AVATAR_PALETTES.length : 0;
  return AVATAR_PALETTES[idx];
}

function initials(name) {
  return (name || '?').substring(0, 2).toUpperCase();
}

function renderMembers(members) {
  const list      = document.getElementById('membersList');
  const badge     = document.getElementById('membersCount');
  const statMem   = document.getElementById('statMembers');
  const statPosts = document.getElementById('statPosts');
  const statLikes = document.getElementById('statLikes');

  // Mise à jour stats globales
  if (badge)     badge.textContent     = `${members.length} membre${members.length !== 1 ? 's' : ''}`;
  if (statMem)   statMem.textContent   = members.length;
  if (statPosts) statPosts.textContent = members.reduce((s, m) => s + m.posts_count, 0);
  if (statLikes) statLikes.textContent = members.reduce((s, m) => s + m.likes_given, 0);

  if (members.length === 0) {
    list.innerHTML = `
      <div class="empty-state">
        <div class="emoji">🔍</div>
        <p>Aucun membre trouvé pour cette recherche.</p>
      </div>`;
    return;
  }

  list.innerHTML = members.map((m, i) => buildMemberCard(m, i)).join('');
}

function buildMemberCard(m, index) {
  const [bgColor, textColor] = avatarColors(m.pseudo);
  const moodClass  = MOOD_CLASSES[m.current_mood] || 'mood-default';
  const moodEmoji  = MOOD_EMOJIS[m.current_mood] || '';
  const delay      = `animation-delay:${index * 40}ms`;

  // 🔴 C'EST CETTE LIGNE QUI RÈGLE TON PROBLÈME (Ajout de http://localhost:8081)
const avatarHTML = m.avatar_url
    ? `<img src="${m.avatar_url.startsWith('http') ? m.avatar_url : 'http://localhost:8081' + m.avatar_url}" alt="avatar">`
    : `<div style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:${bgColor};color:${textColor};font-weight:700;font-size:18px;border-radius:50%">${initials(m.pseudo)}</div>`;

  const moodHTML = m.current_mood
    ? `<span class="member-mood ${moodClass}">${moodEmoji} ${escHtml(m.current_mood)}</span>`
    : `<span class="member-mood mood-default" style="opacity:.6">Aucune humeur</span>`;

  return `
    <a class="member-card" href="/profile/${m.id}" style="animation:slideUp .35s ease both;${delay}">
      <div class="member-avatar">${avatarHTML}</div>
      <div class="member-info">
        <div class="member-name">${escHtml(m.pseudo)}</div>
        ${moodHTML}
        <div class="member-stats">
          <span class="member-stat">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>
            ${m.posts_count}
          </span>
          <span class="member-stat">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2-2z"></path></svg>
            ${m.comments_count}
          </span>
          <span class="member-stat">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>
            ${m.likes_given}
          </span>
        </div>
      </div>
      <div class="member-arrow">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </div>
    </a>`;
}

// ─────────────────────────────────────────────────────────────────────────────
// EVENT LISTENERS
// ─────────────────────────────────────────────────────────────────────────────
function setupListeners() {
  // Navigation sidebar
  document.querySelectorAll('.nav-item[data-href]').forEach(btn => {
    btn.addEventListener('click', () => { window.location.href = btn.dataset.href; });
  });

  // Recherche (debounced 300ms)
  const searchInput = document.getElementById('networkSearch');
  if (searchInput) {
    searchInput.addEventListener('input', () => {
      clearTimeout(searchTimer);
      searchTimer = setTimeout(loadMembers, 300);
    });
    // Effacer avec Escape
    searchInput.addEventListener('keydown', e => {
      if (e.key === 'Escape') { searchInput.value = ''; loadMembers(); }
    });
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// UTILS
// ─────────────────────────────────────────────────────────────────────────────
function escHtml(str) {
  if (typeof str !== 'string') return '';
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
            .replace(/"/g,'&quot;').replace(/'/g,'&#039;');
}