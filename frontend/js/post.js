// post.js — Page détail d'un post
const API_BASE = window.location.hostname === 'localhost' 
    ? 'http://localhost:8081/api' 
    : '/api'; // REMPLACE PAR TON VRAI LIEN RENDER
const fetchOpts = { credentials: 'include' };

// ── État global ───────────────────────────────────────────────────────────────
let currentUser = null;
let postData    = null;
let postId      = null;
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
document.addEventListener('DOMContentLoaded', async () => {
  // Extraire l'ID depuis l'URL : /post/42
  const parts = window.location.pathname.split('/');
  postId = parseInt(parts[parts.length - 1], 10);
  if (!postId || isNaN(postId)) {
    renderError('ID de post invalide.');
    return;
  }

  await checkAuth();
  await loadPost();

  // Scroll vers les commentaires si #comments dans l'URL
  if (window.location.hash === '#comments') {
    setTimeout(() => {
      document.getElementById('commentsContainer')?.scrollIntoView({ behavior: 'smooth' });
    }, 300);
  }
});

// ─────────────────────────────────────────────────────────────────────────────
// AUTH
// ─────────────────────────────────────────────────────────────────────────────
async function checkAuth() {
  try {
    const res = await fetch(`${API_BASE}/me`, fetchOpts);
    if (res.ok) currentUser = await res.json();
  } catch (_) {}
}

// ─────────────────────────────────────────────────────────────────────────────
// CHARGEMENT DU POST
// ─────────────────────────────────────────────────────────────────────────────
async function loadPost() {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, fetchOpts);
    if (!res.ok) { renderError('Post introuvable ou supprimé.'); return; }
    const data = await res.json();
    postData = data.post;

    // Vérifier si l'utilisateur a déjà liké ce post
    let liked = false;
    if (currentUser) {
      try {
        const lr = await fetch(`${API_BASE}/posts/${postId}/liked`, fetchOpts);
        if (lr.ok) { const ld = await lr.json(); liked = ld.liked; }
      } catch (_) {}
    }

    renderPost(postData, liked);
    renderComments(data.comments || []);
    renderCommentForm();
  } catch (err) {
    console.error(err);
    renderError('Erreur lors du chargement du post.');
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// RENDU DU POST
// ─────────────────────────────────────────────────────────────────────────────
const TAG_COLORS = {
  'Stress': '#ff5722', 'Anxiété': '#ff5722',
  'Bien-être': '#00c853', 'Sport': '#00c853', 'Santé mentale': '#00c853',
  'Solitude': '#9c27b0', 'Relations': '#9c27b0',
  'Études': '#3b82f6', 'Travail': '#3b82f6',
  'Autre': '#6b7280',
};

function tagColor(cat) { return TAG_COLORS[cat] || '#6b7280'; }

function timeAgo(dateStr) {
  const diff = (Date.now() - new Date(dateStr)) / 1000;
  if (diff < 60)    return 'À l\'instant';
  if (diff < 3600)  return `${Math.floor(diff/60)} min`;
  if (diff < 86400) return `${Math.floor(diff/3600)} h`;
  if (diff < 2592000) return `${Math.floor(diff/86400)} j`;
  return new Date(dateStr).toLocaleDateString('fr-FR', { day:'numeric', month:'short', year:'numeric' });
}

function initials(name) {
  return (name || '?').substring(0, 2).toUpperCase();
}

function avatarColors(name) {
  const palettes = [
    ['#dbeafe','#1d4ed8'], ['#dcfce7','#166534'], ['#f3e8ff','#7c3aed'],
    ['#fef3c7','#92400e'], ['#fce7f3','#9d174d'], ['#e0f2fe','#0369a1'],
  ];
  const idx = name ? name.charCodeAt(0) % palettes.length : 0;
  return palettes[idx];
}

function renderPost(post, liked) {
  document.title = `Zukuk — ${post.title}`;
  const isOwner = currentUser && currentUser.id === post.user_id;
  const [bgColor, textColor] = avatarColors(post.author);

const avatarHTML = post.avatar_url
    ? `<img src="${post.avatar_url.startsWith('http') ? post.avatar_url : 'http://localhost:8081' + post.avatar_url}" style="width:48px;height:48px;border-radius:50%;object-fit:cover" alt="avatar">`
    : `<div style="width:48px;height:48px;border-radius:50%;background:${bgColor};color:${textColor};display:flex;align-items:center;justify-content:center;font-weight:700;font-size:16px;flex-shrink:0">${initials(post.author)}</div>`;

  const fullDate = new Date(post.created_at).toLocaleDateString('fr-FR', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit'
  });

  document.getElementById('postContainer').innerHTML = `
    <div class="post-card">
      <div class="post-meta">
        <div class="post-avatar">${avatarHTML}</div>
        <div class="post-author-block">
          <div class="author-name">${escHtml(post.author)}</div>
          <div class="author-time" title="${fullDate}">${timeAgo(post.created_at)}</div>
        </div>
        ${post.category
          ? `<span class="post-category" style="background:${tagColor(post.category)}">${escHtml(post.category)}</span>`
          : ''}
      </div>

      <h1 class="post-title">${escHtml(post.title)}</h1>
      <div class="post-content">${escHtml(post.content)}</div>

      <div class="post-actions-row">
        <button class="action-btn${liked ? ' liked' : ''}" id="likeBtn" title="${currentUser ? 'Liker ce post' : 'Connectez-vous pour liker'}">
          <svg viewBox="0 0 24 24" fill="${liked ? 'currentColor' : 'none'}" stroke="currentColor">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
          </svg>
          <span id="likeCount">${post.likes_count}</span>
        </button>

        <button class="action-btn" id="scrollToComments">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
          </svg>
          <span id="commentCount">${post.comments_count}</span>
        </button>

        ${isOwner ? `
          <button class="action-btn btn-edit" id="editPostBtn" title="Modifier ce post">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
            Modifier
          </button>
          <button class="action-btn btn-delete" id="deletePostBtn" title="Supprimer ce post">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6l-1 14H6L5 6"></path><path d="M10 11v6M14 11v6"></path><path d="M9 6V4h6v2"></path></svg>
            Supprimer
          </button>` : ''}
      </div>
    </div>`;

  // ── Listeners ──────────────────────────────────────────────────────────────
  document.getElementById('likeBtn').addEventListener('click', handleLike);
  document.getElementById('scrollToComments').addEventListener('click', () =>
    document.getElementById('commentsContainer')?.scrollIntoView({ behavior:'smooth' }));

  if (isOwner) {
    document.getElementById('editPostBtn').addEventListener('click', openEditPostModal);
    document.getElementById('deletePostBtn').addEventListener('click', handleDeletePost);
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// LIKE / UNLIKE
// ─────────────────────────────────────────────────────────────────────────────
async function handleLike() {
  if (!currentUser) { showAuthToast(); return; }
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/like`, {
      method: 'POST', credentials: 'include',
    });
    if (!res.ok) return;
    const data = await res.json();
    const btn = document.getElementById('likeBtn');
    const svg = btn.querySelector('svg');
    btn.classList.toggle('liked', data.liked);
    svg.setAttribute('fill', data.liked ? 'currentColor' : 'none');
    document.getElementById('likeCount').textContent = data.count;
  } catch (err) { console.error(err); }
}

// ─────────────────────────────────────────────────────────────────────────────
// RENDU DES COMMENTAIRES
// ─────────────────────────────────────────────────────────────────────────────
function renderComments(comments) {
  const container = document.getElementById('commentsContainer');
  container.id    = 'commentsContainer';

  if (comments.length === 0) {
    container.innerHTML = `
      <div class="comments-section" id="commentsAnchor">
        <div class="comments-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:18px;height:18px"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
          Commentaires <span class="comments-count-badge">0</span>
        </div>
        <p style="color:#9ca3af;font-size:0.9rem;text-align:center;padding:20px 0">Aucun commentaire pour l'instant. Sois le premier !</p>
      </div>`;
    return;
  }

  const html = comments.map(cm => buildCommentHTML(cm)).join('');
  container.innerHTML = `
    <div class="comments-section" id="commentsAnchor">
      <div class="comments-header">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:18px;height:18px"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
        Commentaires <span class="comments-count-badge">${comments.length}</span>
      </div>
      <div id="commentsList">${html}</div>
    </div>`;

  attachCommentListeners();
  updateCommentCount(comments.length);
}

function buildCommentHTML(cm) {
  const isOwner = currentUser && currentUser.id === cm.user_id;
  const [bgColor, textColor] = avatarColors(cm.author);

  const avatarHTML = cm.avatar_url
    ? `<img src="${cm.avatar_url.startsWith('http') ? cm.avatar_url : 'http://localhost:8081' + cm.avatar_url}" style="width:100%;height:100%;object-fit:cover;border-radius:50%" alt="avatar">`
    : `<div style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:${bgColor};color:${textColor};font-weight:700;font-size:11px;border-radius:50%">${initials(cm.author)}</div>`;

  return `
    <div class="comment-card" data-comment-id="${cm.id}">
      <div class="comment-meta">
        <div class="comment-avatar">${avatarHTML}</div>
        <span class="comment-author">${escHtml(cm.author)}</span>
        <span class="comment-time">• ${timeAgo(cm.created_at)}</span>
        ${isOwner ? `
          <div class="comment-owner-btns">
            <button class="comment-owner-btn edit-c" data-id="${cm.id}" title="Modifier">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:14px;height:14px"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
            </button>
            <button class="comment-owner-btn delete-c" data-id="${cm.id}" title="Supprimer">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:14px;height:14px"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6l-1 14H6L5 6"></path><path d="M10 11v6M14 11v6"></path><path d="M9 6V4h6v2"></path></svg>
            </button>
          </div>` : ''}
      </div>
      <div class="comment-content" id="cm-content-${cm.id}">${escHtml(cm.content)}</div>
    </div>`;
}

function attachCommentListeners() {
  document.querySelectorAll('.edit-c').forEach(btn => {
    btn.addEventListener('click', () => openEditComment(btn.dataset.id));
  });
  document.querySelectorAll('.delete-c').forEach(btn => {
    btn.addEventListener('click', () => handleDeleteComment(btn.dataset.id));
  });
}

function updateCommentCount(n) {
  const badge = document.querySelector('.comments-count-badge');
  if (badge) badge.textContent = n;
  const countEl = document.getElementById('commentCount');
  if (countEl) countEl.textContent = n;
}

// ─────────────────────────────────────────────────────────────────────────────
// FORMULAIRE COMMENTAIRE
// ─────────────────────────────────────────────────────────────────────────────
function renderCommentForm() {
  const container = document.getElementById('commentFormContainer');
  if (!currentUser) {
    container.innerHTML = `
      <div class="auth-nudge">
        <a href="/login">Connectez-vous</a> ou <a href="/register">créez un compte</a> pour laisser un commentaire.
      </div>`;
    return;
  }

  container.innerHTML = `
    <div class="comment-form-card">
      <div class="comment-form-title">Laisser un commentaire</div>
      <textarea class="comment-textarea" id="commentInput" rows="3" placeholder="Partage ton avis, ton vécu, un conseil..."></textarea>
      <div class="comment-form-footer">
        <label class="anon-label">
          <input type="checkbox" id="commentAnon">
          <div class="anon-checkbox">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="20 6 9 17 4 12"></polyline></svg>
          </div>
          Poster anonymement
        </label>
        <button class="submit-comment-btn" id="submitComment">Publier</button>
      </div>
      <div id="commentError" style="display:none;color:#dc2626;font-size:0.82rem;margin-top:8px"></div>
    </div>`;

  // Toggle checkbox visuel
  document.getElementById('commentAnon').addEventListener('change', function() {
    const cb = this.nextElementSibling;
    cb.style.background  = this.checked ? '#818cf8' : 'white';
    cb.style.borderColor = this.checked ? '#818cf8' : '#d1d5db';
    cb.querySelector('svg').style.opacity = this.checked ? '1' : '0';
  });

  document.getElementById('submitComment').addEventListener('click', handleAddComment);
  document.getElementById('commentInput').addEventListener('keydown', e => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) handleAddComment();
  });
}

async function handleAddComment() {
  const content = document.getElementById('commentInput').value.trim();
  const anon    = document.getElementById('commentAnon').checked ? 1 : 0;
  const errEl   = document.getElementById('commentError');

  if (!content) { showCommentError('Le commentaire ne peut pas être vide.'); return; }

  const btn = document.getElementById('submitComment');
  btn.disabled    = true;
  btn.textContent = 'Publication…';

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      method: 'POST', credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content, is_anonymous: anon }),
    });
    const newComment = await res.json();
    if (!res.ok) { showCommentError(newComment.error || 'Erreur.'); return; }

    // Insérer le nouveau commentaire dans le DOM
    const list = document.getElementById('commentsList');
    if (!list) {
      // Reconstruire la section s'il n'y avait pas de commentaires
      document.getElementById('commentsContainer').innerHTML = `
        <div class="comments-section" id="commentsAnchor">
          <div class="comments-header">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:18px;height:18px"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
            Commentaires <span class="comments-count-badge">1</span>
          </div>
          <div id="commentsList"></div>
        </div>`;
    }

    const newEl = document.createElement('div');
    newEl.innerHTML = buildCommentHTML(newComment);
    const commentCard = newEl.firstElementChild;
    commentCard.style.opacity   = '0';
    commentCard.style.transform = 'translateY(10px)';
    document.getElementById('commentsList').appendChild(commentCard);
    requestAnimationFrame(() => {
      commentCard.style.transition = 'opacity .3s ease, transform .3s ease';
      commentCard.style.opacity   = '1';
      commentCard.style.transform = 'translateY(0)';
    });
    attachCommentListeners();

    const currentCount = parseInt(document.querySelector('.comments-count-badge')?.textContent || '0', 10);
    updateCommentCount(currentCount + 1);

    // Reset form
    document.getElementById('commentInput').value = '';
    errEl.style.display = 'none';
    commentCard.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  } catch (err) {
    showCommentError('Erreur lors de la publication.');
    console.error(err);
  } finally {
    btn.disabled    = false;
    btn.textContent = 'Publier';
  }
}

function showCommentError(msg) {
  const el = document.getElementById('commentError');
  if (!el) return;
  el.textContent   = msg;
  el.style.display = 'block';
}

// ─────────────────────────────────────────────────────────────────────────────
// ÉDITION COMMENTAIRE (inline)
// ─────────────────────────────────────────────────────────────────────────────
function openEditComment(commentId) {
  const contentEl = document.getElementById(`cm-content-${commentId}`);
  if (!contentEl) return;
  const original = contentEl.textContent;

  contentEl.innerHTML = `
    <textarea class="comment-edit-area" id="edit-area-${commentId}" rows="3">${escHtml(original)}</textarea>
    <div class="comment-edit-actions">
      <button id="save-cm-${commentId}" style="background:#818cf8;color:white">Enregistrer</button>
      <button id="cancel-cm-${commentId}" style="background:#f1f5f9;color:#374151">Annuler</button>
    </div>`;

  document.getElementById(`save-cm-${commentId}`).addEventListener('click', () =>
    saveEditComment(commentId, original));
  document.getElementById(`cancel-cm-${commentId}`).addEventListener('click', () => {
    contentEl.innerHTML = escHtml(original);
    // Réactiver les boutons de la carte
    const card = document.querySelector(`.comment-card[data-comment-id="${commentId}"]`);
    const editBtn = card?.querySelector('.edit-c');
    if (editBtn) editBtn.addEventListener('click', () => openEditComment(commentId));
  });
}

async function saveEditComment(commentId, original) {
  const textarea = document.getElementById(`edit-area-${commentId}`);
  const content  = textarea?.value.trim();
  if (!content) { alert('Le commentaire ne peut pas être vide.'); return; }

  try {
    const res = await fetch(`${API_BASE}/comments/${commentId}`, {
      method: 'PUT', credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    });
    const data = await res.json();
    if (!res.ok) { alert(data.error || 'Erreur.'); return; }
    document.getElementById(`cm-content-${commentId}`).textContent = content;
    attachCommentListeners();
  } catch (err) {
    console.error(err);
    document.getElementById(`cm-content-${commentId}`).textContent = original;
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// SUPPRESSION COMMENTAIRE
// ─────────────────────────────────────────────────────────────────────────────
async function handleDeleteComment(commentId) {
  if (!confirm('Supprimer ce commentaire ?')) return;
  try {
    const res = await fetch(`${API_BASE}/comments/${commentId}`, {
      method: 'DELETE', credentials: 'include',
    });
    if (!res.ok) return;
    const card = document.querySelector(`.comment-card[data-comment-id="${commentId}"]`);
    if (card) {
      card.style.transition = 'opacity .3s ease, transform .3s ease';
      card.style.opacity    = '0';
      card.style.transform  = 'translateY(-8px)';
      setTimeout(() => card.remove(), 300);
      const currentCount = parseInt(document.querySelector('.comments-count-badge')?.textContent || '1', 10);
      updateCommentCount(Math.max(0, currentCount - 1));
    }
  } catch (err) { console.error(err); }
}

// ─────────────────────────────────────────────────────────────────────────────
// ÉDITION POST (modal)
// ─────────────────────────────────────────────────────────────────────────────
function openEditPostModal() {
  if (!postData) return;
  const overlay = document.createElement('div');
  overlay.style.cssText = `
    position:fixed;inset:0;background:rgba(0,0,0,0.35);backdrop-filter:blur(4px);
    z-index:1000;display:flex;align-items:center;justify-content:center;padding:16px`;
  overlay.id = 'editModal';
  overlay.innerHTML = `
    <div style="background:white;border-radius:24px;padding:32px;width:100%;max-width:540px;
                box-shadow:0 20px 60px rgba(0,0,0,0.15);position:relative">
      <button onclick="document.getElementById('editModal').remove()" style="position:absolute;top:14px;right:16px;
              background:none;border:none;font-size:22px;cursor:pointer;color:#9ca3af">✕</button>
      <h3 style="font-size:1.2rem;font-weight:700;color:#1e293b;margin-bottom:20px">Modifier le post</h3>
      <div class="modal-field">
        <label>Titre</label>
        <input id="edit-title" class="modal-modern-input" value="${escHtml(postData.title)}">
      </div>
      <div class="modal-field">
        <label>Contenu</label>
        <textarea id="edit-content" class="modal-modern-input" rows="6">${escHtml(postData.content)}</textarea>
      </div>
      <div id="edit-error" style="display:none;color:#dc2626;font-size:0.85rem;margin-bottom:10px"></div>
      <div style="display:flex;gap:12px;margin-top:16px">
        <button onclick="document.getElementById('editModal').remove()" style="flex:1;padding:12px;
          border:2px solid #e2e8f0;border-radius:12px;background:white;font-weight:600;cursor:pointer;font-family:inherit">Annuler</button>
        <button id="saveEditPost" style="flex:2;padding:12px;background:linear-gradient(135deg,#818cf8,#a78bfa);
          color:white;border:none;border-radius:12px;font-weight:600;cursor:pointer;font-family:inherit">Enregistrer</button>
      </div>
    </div>`;
  document.body.appendChild(overlay);
  overlay.addEventListener('click', e => { if (e.target === overlay) overlay.remove(); });

  document.getElementById('saveEditPost').addEventListener('click', async () => {
    const title   = document.getElementById('edit-title').value.trim();
    const content = document.getElementById('edit-content').value.trim();
    if (title.length < 3 || content.length < 10) {
      document.getElementById('edit-error').textContent = 'Titre (3+) et contenu (10+) requis.';
      document.getElementById('edit-error').style.display = 'block';
      return;
    }
    try {
      const res = await fetch(`${API_BASE}/posts/${postId}`, {
        method: 'PUT', credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content }),
      });
      const data = await res.json();
      if (!res.ok) { document.getElementById('edit-error').textContent = data.error; document.getElementById('edit-error').style.display='block'; return; }
      // Mettre à jour le DOM
      document.querySelector('.post-title').textContent   = title;
      document.querySelector('.post-content').textContent = content;
      postData.title   = title;
      postData.content = content;
      overlay.remove();
    } catch (err) {
      document.getElementById('edit-error').textContent = 'Erreur serveur.';
      document.getElementById('edit-error').style.display = 'block';
    }
  });
}

// ─────────────────────────────────────────────────────────────────────────────
// SUPPRESSION POST
// ─────────────────────────────────────────────────────────────────────────────
async function handleDeletePost() {
  if (!confirm('Supprimer définitivement ce post et tous ses commentaires ?')) return;
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, { method:'DELETE', credentials:'include' });
    if (res.ok) window.location.href = '/board';
  } catch (err) { console.error(err); }
}

// ─────────────────────────────────────────────────────────────────────────────
// AUTH TOAST
// ─────────────────────────────────────────────────────────────────────────────
function showAuthToast() {
  const existing = document.getElementById('authToast');
  if (existing) { existing.style.animation='none'; void existing.offsetWidth; existing.style.animation=''; return; }
  const toast = document.createElement('div');
  toast.id = 'authToast';
  toast.innerHTML = `<span>🔒 <a href="/login" style="color:#818cf8;font-weight:700">Connectez-vous</a> pour interagir</span>`;
  toast.style.cssText = `
    position:fixed;bottom:24px;left:50%;transform:translateX(-50%);
    background:white;padding:14px 24px;border-radius:16px;
    box-shadow:0 8px 30px rgba(0,0,0,0.12);font-size:0.95rem;
    color:#1e293b;z-index:2000;animation:slideUp .3s ease;
    border:1px solid rgba(129,140,248,0.3)`;
  document.body.appendChild(toast);
  setTimeout(() => { if (toast.parentNode) toast.remove(); }, 3500);
}

// ─────────────────────────────────────────────────────────────────────────────
// ERREUR PAGE
// ─────────────────────────────────────────────────────────────────────────────
function renderError(msg) {
  document.getElementById('postContainer').innerHTML = `
    <div style="text-align:center;padding:60px 20px">
      <div style="font-size:3rem;margin-bottom:16px">😕</div>
      <h2 style="color:#1f2937;margin-bottom:8px">${escHtml(msg)}</h2>
      <a href="/board" style="color:#818cf8;font-weight:600;text-decoration:none">← Retour au forum</a>
    </div>`;
}

// ─────────────────────────────────────────────────────────────────────────────
// UTILS
// ─────────────────────────────────────────────────────────────────────────────
function escHtml(str) {
  if (typeof str !== 'string') return '';
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
            .replace(/"/g,'&quot;').replace(/'/g,'&#039;');
}