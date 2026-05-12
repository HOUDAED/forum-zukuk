// Board JavaScript - Frontend Logic (v2 – posts réels + recherche + mode invité + PAGINATION + TOASTS)
const API_BASE = window.location.hostname === 'localhost' 
    ? 'http://localhost:8081/api' 
    : '/api';// REMPLACE PAR TON VRAI LIEN RENDER

// ─── UI : TOAST & MODALE DE CONFIRMATION ────────────────────────────────────
function showToast(msg, isError = false) {
    const existing = document.getElementById('board-toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.id = 'board-toast';
    toast.textContent = msg;
    toast.style.cssText = `position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%) translateY(100px); background: ${isError ? '#fef2f2' : '#f0fdf4'}; color: ${isError ? '#ef4444' : '#166534'}; border: 1px solid ${isError ? '#fecaca' : '#bbf7d0'}; padding: 14px 28px; border-radius: 16px; font-weight: 600; font-size: 0.95rem; box-shadow: 0 10px 30px rgba(0,0,0,0.1); z-index: 9999; transition: transform 0.4s cubic-bezier(0.2, 0.9, 0.2, 1);`;
    document.body.appendChild(toast);
    
    requestAnimationFrame(() => toast.style.transform = 'translateX(-50%) translateY(0)');
    setTimeout(() => {
        toast.style.transform = 'translateX(-50%) translateY(100px)';
        setTimeout(() => toast.remove(), 400);
    }, 3500);
}

function showConfirmModal(title, message, onConfirm) {
    const existing = document.getElementById('zukuk-confirm');
    if (existing) existing.remove();

    const overlay = document.createElement('div');
    overlay.id = 'zukuk-confirm';
    overlay.style.cssText = `
        position: fixed; inset: 0; background: rgba(15, 23, 42, 0.4); backdrop-filter: blur(8px);
        z-index: 10000; display: flex; align-items: center; justify-content: center; padding: 20px;
        opacity: 0; transition: opacity 0.3s ease;
    `;

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

    requestAnimationFrame(() => {
        overlay.style.opacity = '1';
        modal.style.transform = 'translateY(0) scale(1)';
    });

    const close = () => {
        overlay.style.opacity = '0';
        modal.style.transform = 'translateY(20px) scale(0.95)';
        setTimeout(() => overlay.remove(), 300);
    };

    document.getElementById('cancel-confirm').onclick = close;
    overlay.onclick = (e) => { if (e.target === overlay) close(); };
    document.getElementById('accept-confirm').onclick = () => { close(); onConfirm(); };
}

// ─────────────────────────────────────────────────────────────────────────────
// 🎨 SYSTÈME DE THÈME GLOBAL ZUKUK
// ─────────────────────────────────────────────────────────────────────────────

const savedTheme = localStorage.getItem('zukuk_theme') || 'light';
document.body.setAttribute('data-theme', savedTheme);

document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch(`${API_BASE}/settings/`, { credentials: 'include' });
        if (res.ok) {
            const settings = await res.json();
            if (settings.theme && settings.theme !== savedTheme) {
                document.body.setAttribute('data-theme', settings.theme);
                localStorage.setItem('zukuk_theme', settings.theme);
            }
        }
    } catch (e) {}
});

// ── DOM refs ──────────────────────────────────────────────────────────────────
const moodGrid       = document.getElementById('moodGrid');
const statsSection   = document.getElementById('statsSection');
const discussionsList = document.getElementById('discussionsList');
const quoteBanner    = document.getElementById('quoteBanner');
const quoteLoader    = document.getElementById('quoteLoader');
const quoteText      = document.getElementById('quoteText');
const navButtons     = document.querySelectorAll('.nav-item');
const searchInput    = document.getElementById('boardSearch');
const sortSelect     = document.getElementById('boardSort');
const orderSelect    = document.getElementById('boardOrder');
const newPostBtn     = document.getElementById('newPostBtn');

let currentMood   = 'Calme';
let currentUser   = null;
let searchTimer   = null;
let currentOffset = 0; 
const POSTS_PER_PAGE = 20;
const fetchOpts = { credentials: 'include' };

document.addEventListener('DOMContentLoaded', async () => {
  await checkAuth();
  updateGreeting();
  renderMoodsLocal();
  await loadCurrentMood();
  loadPosts(false);
  renderStatsLocal();
  fetchRandomQuote();
  setupEventListeners();
  updateUIForAuthState();
});

async function checkAuth() {
  try {
    const res = await fetch(`${API_BASE}/me`, fetchOpts);
    if (res.ok) currentUser = await res.json();
  } catch (_) {}
}

function updateUIForAuthState() {
  if (!currentUser) {
    if (newPostBtn) {
      newPostBtn.disabled = true;
      newPostBtn.title    = 'Connectez-vous pour publier';
      newPostBtn.style.opacity = '0.4';
    }
  }
}

async function loadPosts(append = false) {
  if (!append) {
      currentOffset = 0;
      if (discussionsList) discussionsList.innerHTML = '<p style="text-align:center; color:var(--text-gray); padding:24px;">Chargement des discussions...</p>';
  }

  const query = searchInput ? searchInput.value.trim() : '';
  const sort  = sortSelect  ? sortSelect.value : 'date';
  const order = orderSelect ? orderSelect.value : 'desc';

  const params = new URLSearchParams({ sort, order, limit: POSTS_PER_PAGE, offset: currentOffset });
  if (query) params.set('query', query);

  try {
    const res = await fetch(`${API_BASE}/posts?${params}`, fetchOpts);
    if (!res.ok) throw new Error('API error');
    const data = await res.json();
    renderPosts(data.posts || [], append);
  } catch (err) {
    if (!append) renderPostsLocal();
  }
}

const TAG_COLORS = {
  'Stress': 'tag-stress', 'Anxiété': 'tag-stress',
  'Bien-être': 'tag-wellbeing', 'Sport': 'tag-wellbeing', 'Santé mentale': 'tag-wellbeing',
  'Solitude': 'tag-loneliness', 'Relations': 'tag-loneliness',
};

function tagClass(cat) { return TAG_COLORS[cat] || 'tag-stress'; }

function timeAgo(dateStr) {
  const diff = (Date.now() - new Date(dateStr)) / 1000;
  if (diff < 60)   return 'À l\'instant';
  if (diff < 3600) return `${Math.floor(diff/60)}min`;
  if (diff < 86400) return `${Math.floor(diff/3600)}h`;
  return `${Math.floor(diff/86400)}j`;
}

function renderPosts(posts, append = false) {
  const loadMoreBtn = document.getElementById('btn-load-more');
  if (!append) discussionsList.innerHTML = '';

  if (posts.length === 0 && currentOffset === 0) {
    discussionsList.innerHTML = '<p style="color:var(--text-gray);text-align:center;padding:24px">Aucun post trouvé.</p>';
    if (loadMoreBtn) loadMoreBtn.style.display = 'none';
    return;
  }

  posts.forEach(post => {
    const card = document.createElement('div');
    card.className = 'discussion-card';
    card.dataset.postId = post.id;
    card.style.opacity   = '0';
    card.style.transform = 'translateY(10px)';
    card.style.transition = `opacity .4s ease, transform .4s ease`;

    const isOwner = currentUser && currentUser.id === post.user_id;
    const ownerActions = isOwner ? `
      <button class="discussion-action edit-post" data-id="${post.id}" title="Modifier">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
      </button>
      <button class="discussion-action delete-post" data-id="${post.id}" title="Supprimer">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6l-1 14H6L5 6"></path><path d="M10 11v6M14 11v6"></path><path d="M9 6V4h6v2"></path></svg>
      </button>` : '';

    card.innerHTML = `
      <div class="discussion-header">
        <div class="discussion-avatar">
          ${post.avatar_url
            ? `<img src="${post.avatar_url.startsWith('http') ? post.avatar_url : 'http://localhost:8081' + post.avatar_url}" style="width:40px;height:40px;border-radius:50%;object-fit:cover" alt="avatar">`
            : `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`}
        </div>
        <div class="discussion-author-info">
          <span class="discussion-author">${escHtml(post.author)}</span>
          <span class="discussion-time">• ${timeAgo(post.created_at)}</span>
        </div>
      </div>
      <h3 class="discussion-title">${escHtml(post.title)}</h3>
      <p class="discussion-excerpt">${escHtml(post.content.substring(0,120))}${post.content.length>120?'…':''}</p>
      <div class="discussion-footer">
        <span class="discussion-tag ${tagClass(post.category)}">${escHtml(post.category||'Général')}</span>
        <div class="discussion-actions">
          <button class="discussion-action like${currentUser?'':' disabled-action'}" data-id="${post.id}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>
            <span>${post.likes_count}</span>
          </button>
          <button class="discussion-action comment" data-id="${post.id}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
            <span>${post.comments_count}</span>
          </button>
          ${ownerActions}
        </div>
      </div>`;

    card.addEventListener('click', (e) => {
      if (e.target.closest('button')) return;
      window.location.href = `/post/${post.id}`;
    });

    const likeBtn = card.querySelector('.like');
    likeBtn.addEventListener('click', e => { e.stopPropagation(); if (!currentUser) { showAuthToast(); return; } toggleLike(post.id, likeBtn); });

    card.querySelector('.comment').addEventListener('click', e => { e.stopPropagation(); window.location.href = `/post/${post.id}#comments`; });

    const editBtn = card.querySelector('.edit-post');
    if (editBtn) editBtn.addEventListener('click', e => { e.stopPropagation(); openEditPostModal(post); });

    const delBtn = card.querySelector('.delete-post');
    if (delBtn) delBtn.addEventListener('click', e => { e.stopPropagation(); deletePost(post.id, card); });

    discussionsList.appendChild(card);
    requestAnimationFrame(() => { card.style.opacity = '1'; card.style.transform = 'translateY(0)'; });
  });

  if (loadMoreBtn) {
      loadMoreBtn.style.display = (posts.length < POSTS_PER_PAGE) ? 'none' : 'inline-block';
  }
}

function renderPostsLocal() {
  const mock = [
    { id:1, author:'Marie_123', created_at: new Date(Date.now()-7200000).toISOString(), title:"J'ai du mal à gérer mon stress au travail", content:"Depuis quelques semaines...", category:'Stress', likes_count:12, comments_count:8, user_id:-1, avatar_url:'' }
  ];
  renderPosts(mock, false);
}

async function toggleLike(postId, btn) {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/like`, { method: 'POST', credentials: 'include' });
    if (res.ok) {
      const data = await res.json();
      btn.querySelector('span').textContent = data.count;
      btn.style.color = data.liked ? '#ff5722' : '';
    }
  } catch (err) { console.error(err); }
}

let categories = [];
async function fetchCategories() {
  try {
    const res = await fetch(`${API_BASE}/categories`, fetchOpts);
    if (res.ok) categories = await res.json();
  } catch (_) {}
}

function buildCategoryOptions(selectedId) {
  return categories.map(c => `<option value="${c.id}"${c.id===selectedId?' selected':''}>${escHtml(c.name)}</option>`).join('');
}

// 🔴 ICI : REMPLACEMENT DE L'ALERTE POUR LA CRÉATION DE POST
function openNewPostModal() {
  if (!currentUser) { showAuthToast(); return; }
  fetchCategories().then(() => {
    showModal('Nouveau post',
      `<div class="modal-field"><label>Titre</label><input id="m-title" class="modern-input" placeholder="Titre de ta discussion"></div>
       <div class="modal-field"><label>Contenu</label><textarea id="m-content" class="modern-input" rows="5" placeholder="Partage ta pensée..."></textarea></div>
       <div class="modal-field"><label>Catégorie</label><select id="m-cat" class="modern-input"><option value="">-- Choisir --</option>${buildCategoryOptions(null)}</select></div>
       <label class="checkbox-label" style="margin-top:8px">
         <input type="checkbox" id="m-anon"><div class="custom-checkbox"><svg viewBox="0 0 24 24" fill="none" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg></div>
         Poster anonymement
       </label>`,
      async () => {
        const title   = document.getElementById('m-title').value.trim();
        const content = document.getElementById('m-content').value.trim();
        const catId   = document.getElementById('m-cat').value || null;
        const anon    = document.getElementById('m-anon').checked ? 1 : 0;

        // Pré-validation Frontend avec Toast
        if (title.length < 3) { showToast("Le titre doit faire au moins 3 caractères.", true); return; }
        if (content.length < 10) { showToast("Le contenu doit faire au moins 10 caractères.", true); return; }

        const btn = document.getElementById('modalConfirm');
        const originalText = btn.textContent;
        btn.textContent = "Publication...";
        btn.disabled = true;

        try {
            const res = await fetch(`${API_BASE}/posts`, {
              method:'POST', credentials:'include',
              headers:{'Content-Type':'application/json'},
              body: JSON.stringify({ title, content, category_id: catId ? parseInt(catId) : null, is_anonymous: anon }),
            });
            const data = await res.json();
            
            if (!res.ok) { 
                showToast(data.error || "Erreur serveur.", true); 
                btn.textContent = originalText;
                btn.disabled = false;
                return; 
            }
            closeModal();
            showToast("✨ Post publié avec succès !", false);
            setTimeout(() => window.location.href = `/post/${data.id}`, 1000);
        } catch (err) {
            showToast("Serveur injoignable.", true);
            btn.textContent = originalText;
            btn.disabled = false;
        }
      }
    );
  });
}

// 🔴 ICI : REMPLACEMENT DE L'ALERTE POUR LA MODIFICATION
function openEditPostModal(post) {
  fetchCategories().then(() => {
    showModal('Modifier le post',
      `<div class="modal-field"><label>Titre</label><input id="m-title" class="modern-input" value="${escHtml(post.title)}"></div>
       <div class="modal-field"><label>Contenu</label><textarea id="m-content" class="modern-input" rows="5">${escHtml(post.content)}</textarea></div>`,
      async () => {
        const title   = document.getElementById('m-title').value.trim();
        const content = document.getElementById('m-content').value.trim();
        
        if (title.length < 3) { showToast("Le titre doit faire au moins 3 caractères.", true); return; }
        if (content.length < 10) { showToast("Le contenu doit faire au moins 10 caractères.", true); return; }

        const btn = document.getElementById('modalConfirm');
        const originalText = btn.textContent;
        btn.textContent = "Sauvegarde...";
        btn.disabled = true;

        try {
            const res = await fetch(`${API_BASE}/posts/${post.id}`, {
              method:'PUT', credentials:'include',
              headers:{'Content-Type':'application/json'},
              body: JSON.stringify({ title, content }),
            });
            const data = await res.json();
            
            if (!res.ok) { 
                showToast(data.error || "Erreur serveur.", true); 
                btn.textContent = originalText;
                btn.disabled = false;
                return; 
            }
            closeModal();
            showToast("✏️ Modification enregistrée !", false);
            loadPosts(false);
        } catch (err) {
            showToast("Serveur injoignable.", true);
            btn.textContent = originalText;
            btn.disabled = false;
        }
      }
    );
  });
}

// 🔴 ICI : REMPLACEMENT DU "CONFIRM" PAR LA MODALE ÉLÉGANTE
async function deletePost(postId, card) {
    showConfirmModal(
        "Supprimer la discussion ?",
        "Cette action est irréversible. Le post et tous ses commentaires seront effacés pour tout le monde.",
        async () => {
            try {
                const res = await fetch(`${API_BASE}/posts/${postId}`, { method:'DELETE', credentials:'include' });
                if (res.ok) {
                    showToast("🗑️ Discussion supprimée.", false);
                    card.style.opacity = '0';
                    card.style.transform = 'translateY(-10px)';
                    setTimeout(() => card.remove(), 300);
                } else {
                    showToast("Erreur lors de la suppression.", true);
                }
            } catch (err) { showToast("Erreur serveur.", true); }
        }
    );
}

function renderMoodsLocal() {
  const moods = [{ name:'Bien', emoji:'😊' }, { name:'Calme', emoji:'😌' }, { name:'Triste', emoji:'😢' }, { name:'Anxieux', emoji:'😰' }, { name:'Colère', emoji:'😡' }];
  moodGrid.innerHTML = '';
  moods.forEach(mood => {
    const btn = document.createElement('button');
    btn.className = `mood-button${mood.name===currentMood?' active':''}`;
    btn.innerHTML = `<span class="mood-emoji">${mood.emoji}</span><span class="mood-name">${mood.name}</span>`;
    btn.addEventListener('click', () => {
      if (!currentUser) { showAuthToast(); return; }
      selectMood(mood.name);
      saveMood(mood);
    });
    moodGrid.appendChild(btn);
  });
}

function selectMood(name) {
  currentMood = name;
  document.querySelectorAll('.mood-button').forEach(b => b.classList.toggle('active', b.textContent.includes(name)));
  const colors = {
    'Bien':    'linear-gradient(135deg,#fdf4ff 0%,#fae8ff 50%,#fce7f3 100%)',
    'Calme':   'linear-gradient(135deg,#f0fdf4 0%,#dcfce7 50%,#e0f2fe 100%)',
    'Triste':  'linear-gradient(135deg,#eff6ff 0%,#dbeafe 50%,#e0e7ff 100%)',
    'Anxieux': 'linear-gradient(135deg,#f8fafc 0%,#f1f5f9 50%,#e2e8f0 100%)',
    'Colère':  'linear-gradient(135deg,#fff7ed 0%,#ffedd5 50%,#fef08a 100%)',
  };
  const bg = document.querySelector('.background-gradient');
  if (bg) { bg.style.transition = 'background 1.5s ease-in-out'; bg.style.background = colors[name] || colors['Calme']; }
  document.body.dataset.mood = name.toLowerCase();
}

async function saveMood(mood) {
  try { await fetch(`${API_BASE}/mood`, { method:'POST', credentials:'include', headers:{'Content-Type':'application/json'}, body: JSON.stringify(mood) }); } catch (_) {}
}

function renderStatsLocal() {
  statsSection.innerHTML = '';
  [{ title:'Posts partagés', color:'blue' }, { title:'Réponses reçues', color:'red' }, { title:'Jours actifs', color:'purple' }].forEach(s => {
    const card = document.createElement('div');
    card.className = 'stat-card';
    card.innerHTML = `<div class="stat-icon ${s.color}"></div><span class="stat-title">${s.title}</span>`;
    statsSection.appendChild(card);
  });
}

async function fetchRandomQuote() {
  try {
    const res = await fetch(`${API_BASE}/quote`, fetchOpts);
    const data = await res.json();
    displayQuote(data.quote);
  } catch (_) { displayQuote("Tu n'es pas seul. Chaque jour est une nouvelle opportunité."); }
}

function displayQuote(q) {
  quoteLoader.style.display = 'none';
  quoteText.textContent = `"${q}"`;
  quoteText.style.display = 'block';
}

async function loadCurrentMood() {
  if (!currentUser) return;
  try {
    const res = await fetch(`${API_BASE}/me/mood-history`, fetchOpts);
    if (!res.ok) return;
    const data = await res.json();
    if (data.history && data.history.length > 0) {
      currentMood = data.history[0].mood;
      selectMood(currentMood);
    }
  } catch (err) {}
}

async function updateGreeting() {
  const el = document.querySelector('.header-title');
  if (!el) return;
  const greeting = new Date().getHours() < 18 ? 'Bonjour' : 'Bonsoir';
  try {
    const res = await fetch(`${API_BASE}/me`, fetchOpts);
    if (res.ok) {
      const user = await res.json();
      el.innerHTML = `${greeting}, ${escHtml(user.pseudo)} <span class="wave-emoji">👋</span>`;
      return;
    }
  } catch (_) {}
  el.innerHTML = `${greeting} <span class="wave-emoji">👋</span>`;
}

function setupEventListeners() {
  navButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      navButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      if (btn.dataset.href) window.location.href = btn.dataset.href;
    });
  });

  if (newPostBtn) newPostBtn.addEventListener('click', openNewPostModal);

  if (searchInput) searchInput.addEventListener('input', () => { clearTimeout(searchTimer); searchTimer = setTimeout(() => loadPosts(false), 350); });
  if (sortSelect)  sortSelect.addEventListener('change', () => loadPosts(false));
  if (orderSelect) orderSelect.addEventListener('change', () => loadPosts(false));

  const loadMoreBtn = document.getElementById('btn-load-more');
  if (loadMoreBtn) {
      loadMoreBtn.addEventListener('click', () => { currentOffset += POSTS_PER_PAGE; loadPosts(true); });
  }

  document.addEventListener('click', e => { if (e.target.id === 'boardModal') closeModal(); });
}

function showModal(title, bodyHTML, onConfirm) {
  const existing = document.getElementById('boardModal');
  if (existing) existing.remove();

  const modal = document.createElement('div');
  modal.id = 'boardModal';
  modal.style.cssText = `position:fixed;inset:0;background:rgba(0,0,0,0.35);backdrop-filter:blur(4px);z-index:1000;display:flex;align-items:center;justify-content:center;padding:16px;`;
  modal.innerHTML = `
    <div style="background:white;border-radius:24px;padding:32px;width:100%;max-width:520px;box-shadow:0 20px 60px rgba(0,0,0,0.15);position:relative;animation:fadeIn .2s ease">
      <button id="closeModal" style="position:absolute;top:16px;right:16px;background:none;border:none;font-size:22px;cursor:pointer;color:#9ca3af">✕</button>
      <h3 style="font-size:1.3rem;font-weight:700;color:#1e293b;margin-bottom:20px">${escHtml(title)}</h3>
      <div id="modalBody">${bodyHTML}</div>
      <div style="display:flex;gap:12px;margin-top:20px">
        <button id="modalCancel" style="flex:1;padding:12px;border:2px solid #e2e8f0;border-radius:12px;background:white;font-weight:600;cursor:pointer;font-family:inherit">Annuler</button>
        <button id="modalConfirm" style="flex:2;padding:12px;background:linear-gradient(135deg,#818cf8,#a78bfa);color:white;border:none;border-radius:12px;font-weight:600;cursor:pointer;font-family:inherit">Confirmer</button>
      </div>
    </div>`;
  document.body.appendChild(modal);

  const anonCheckbox = modal.querySelector('#m-anon');
  if (anonCheckbox) {
    anonCheckbox.addEventListener('change', function() {
      const cb = this.nextElementSibling;
      cb.style.background   = this.checked ? 'var(--primary,#8b5cf6)' : 'white';
      cb.style.borderColor  = this.checked ? 'var(--primary,#8b5cf6)' : '#e2e8f0';
      cb.querySelector('svg').style.opacity = this.checked ? '1' : '0';
    });
  }

  document.getElementById('closeModal').addEventListener('click', closeModal);
  document.getElementById('modalCancel').addEventListener('click', closeModal);
  document.getElementById('modalConfirm').addEventListener('click', onConfirm);
}

function closeModal() {
  const m = document.getElementById('boardModal');
  if (m) m.remove();
}

function showAuthToast() {
  const existing = document.getElementById('authToast');
  if (existing) { existing.style.animation = 'none'; void existing.offsetWidth; existing.style.animation=''; return; }

  const toast = document.createElement('div');
  toast.id = 'authToast';
  toast.innerHTML = `<span>🔒 <a href="/login" style="color:#818cf8;font-weight:700;text-decoration:underline">Connectez-vous</a> pour interagir</span>`;
  toast.style.cssText = `position:fixed;bottom:24px;left:50%;transform:translateX(-50%);background:white;padding:14px 24px;border-radius:16px;box-shadow:0 8px 30px rgba(0,0,0,0.12);font-size:0.95rem;color:#1e293b;z-index:2000;animation:slideUp .3s ease;border:1px solid rgba(129,140,248,0.3)`;
  document.body.appendChild(toast);
  setTimeout(() => { if (toast.parentNode) toast.remove(); }, 3500);
}

function escHtml(str) {
  if (typeof str !== 'string') return '';
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#039;');
}