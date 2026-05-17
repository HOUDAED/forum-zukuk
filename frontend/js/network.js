const API_BASE = '/api';
const fetchOpts = { credentials: 'include' };

let searchTimer = null;

// ─────────────────────────────────────────────────────────────────────────────
// 🎨 SYSTÈME DE THÈME GLOBAL ZUKUK
// ─────────────────────────────────────────────────────────────────────────────

// 1. Anti-scintillement : Appliquer le thème local IMMÉDIATEMENT
const savedTheme = localStorage.getItem('zukuk_theme') || 'glass';
const savedMood = localStorage.getItem('zukuk_mood') || 'calme';
document.documentElement.setAttribute('data-theme', savedTheme);
document.documentElement.setAttribute('data-mood', savedMood);

// 2. Synchronisation BDD : Vérifier le vrai thème et la vraie humeur
document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch(`${API_BASE}/settings/`, { credentials: 'include' });
        if (res.ok) {
            const settings = await res.json();
            if (settings.theme && settings.theme !== savedTheme) {
                document.documentElement.setAttribute('data-theme', settings.theme);
                localStorage.setItem('zukuk_theme', settings.theme);
            }
            // 🔴 AJOUT : Synchronise et applique la police
            if (settings.font) {
                document.documentElement.setAttribute('data-font', settings.font);
                localStorage.setItem('zukuk_font', settings.font);
            }
        }

        const moodRes = await fetch(`${API_BASE}/me/mood-history`, { credentials: 'include' });
        if (moodRes.ok) {
            const moodData = await moodRes.json();
            if (moodData.history && moodData.history.length > 0) {
                const currentMood = moodData.history[0].mood.toLowerCase().replace('è', 'e');
                document.documentElement.setAttribute('data-mood', currentMood);
                localStorage.setItem('zukuk_mood', currentMood);
            }
        }
    } catch (e) {}
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
  'Bien': '😊', 'Calme': '😌', 'Triste': '😢', 'Anxieux': '😰', 'Colère': '😡', 'Colere': '😡'
};
const MOOD_CLASSES = {
  'Bien': 'mood-bien', 'Calme': 'mood-calme', 'Triste': 'mood-triste',
  'Anxieux': 'mood-anxieux', 'Colère': 'mood-colère', 'Colere': 'mood-colère'
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

  const avatarHTML = m.avatar_url
    ? `<img src="${m.avatar_url.startsWith('http') ? m.avatar_url : (window.location.hostname === 'localhost' ? 'http://localhost:8081' : '') + m.avatar_url}" alt="avatar">`
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
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
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
  document.querySelectorAll('.nav-item[data-href]').forEach(btn => {
    btn.addEventListener('click', () => { window.location.href = btn.dataset.href; });
  });

  const searchInput = document.getElementById('networkSearch');
  if (searchInput) {
    searchInput.addEventListener('input', () => {
      clearTimeout(searchTimer);
      searchTimer = setTimeout(loadMembers, 300);
    });
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