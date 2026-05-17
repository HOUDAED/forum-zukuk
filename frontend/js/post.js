// post.js — Page détail d'un post (Version corrigée Anonymat)
const API_BASE = '/api'; 
const fetchOpts = { credentials: 'include' };

// ── État global ───────────────────────────────────────────────────────────────
let currentUser = null;
let userSettings = null; // 🔴 NOUVEAU : Sauvegarde des paramètres utilisateurs
let postData    = null;
let postId      = null;

const getPostContainer = () => document.getElementById('postContainer') || document.getElementById('post-container');
const getCommentsContainer = () => document.getElementById('commentsContainer') || document.getElementById('comments-container');
const getCommentFormContainer = () => document.getElementById('commentFormContainer') || document.getElementById('comment-form-container');

const savedTheme = localStorage.getItem('zukuk_theme') || 'light';
document.body.setAttribute('data-theme', savedTheme);

document.addEventListener('DOMContentLoaded', async () => {
  const parts = window.location.pathname.split('/').filter(p => p !== "");
  postId = parseInt(parts[parts.length - 1], 10);
  
  if (!postId || isNaN(postId)) { 
    renderError('ID de post invalide ou introuvable dans l\'URL.'); 
    return; 
  }

  await checkAuth();
  await loadPost();

  if (window.location.hash === '#comments') {
    setTimeout(() => { 
        const anchor = getCommentsContainer();
        anchor?.scrollIntoView({ behavior: 'smooth' }); 
    }, 300);
  }
});

async function checkAuth() {
  try {
    // 1. Vérification de l'utilisateur (Indispensable pour liker/commenter)
    const res = await fetch(`${API_BASE}/me`, fetchOpts);
    if (res.ok) currentUser = await res.json();

    // 2. Récupération des paramètres (Thème + Police + Mode Anonyme)
    const setRes = await fetch(`${API_BASE}/settings/`, fetchOpts);
    if (setRes.ok) {
        userSettings = await setRes.json();
        
        // Mise à jour du Thème
        if (userSettings.theme && userSettings.theme !== savedTheme) {
            document.documentElement.setAttribute('data-theme', userSettings.theme);
            localStorage.setItem('zukuk_theme', userSettings.theme);
        }
        
        // 🔴 TON AJOUT : Synchronise et applique la police
        if (userSettings.font) {
            document.documentElement.setAttribute('data-font', userSettings.font);
            localStorage.setItem('zukuk_font', userSettings.font);
        }
    }
    
    // 3. Récupération de l'humeur depuis la base de données
    const moodRes = await fetch(`${API_BASE}/me/mood-history`, fetchOpts);
    if (moodRes.ok) {
        const moodData = await moodRes.json();
        if (moodData.history && moodData.history.length > 0) {
            const currentMood = moodData.history[0].mood.toLowerCase().replace('è', 'e');
            document.documentElement.setAttribute('data-mood', currentMood);
            localStorage.setItem('zukuk_mood', currentMood);
        }
    }
  } catch (_) {}
}

async function loadPost() {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, fetchOpts);
    if (!res.ok) { renderError('Le serveur a renvoyé une erreur : ' + res.status); return; }
    
    let data = await res.json();
    let extractedPost = data.post || data.data || data;
    if (Array.isArray(extractedPost)) extractedPost = extractedPost[0];
    postData = extractedPost;

    if (!postData || (!postData.title && !postData.Title)) {
        renderError('Impossible de lire la structure de la discussion.', JSON.stringify(data));
        return;
    }

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
    console.error("🔥 Crash JS intercepté :", err); 
    renderError('Erreur interne d\'affichage du script.', err.stack || err.message);
  }
}

const TAG_COLORS = { 'Stress': '#ff5722', 'Anxiété': '#ff5722', 'Bien-être': '#00c853', 'Sport': '#00c853', 'Santé mentale': '#00c853', 'Solitude': '#9c27b0', 'Relations': '#9c27b0', 'Études': '#3b82f6', 'Travail': '#3b82f6', 'Autre': '#6b7280'};
function tagColor(cat) { return TAG_COLORS[cat] || '#6b7280'; }

function timeAgo(dateStr) {
  if (!dateStr) return 'Date inconnue';
  const diff = (Date.now() - new Date(dateStr)) / 1000;
  if (isNaN(diff)) return 'Récemment';
  if (diff < 60)    return 'À l\'instant';
  if (diff < 3600)  return `${Math.floor(diff/60)} min`;
  if (diff < 86400) return `${Math.floor(diff/3600)} h`;
  if (diff < 2592000) return `${Math.floor(diff/86400)} j`;
  return new Date(dateStr).toLocaleDateString('fr-FR', { day:'numeric', month:'short', year:'numeric' });
}

function initials(name) { return (name || '?').substring(0, 2).toUpperCase(); }

function avatarColors(name) {
  const palettes = [['#dbeafe','#1d4ed8'], ['#dcfce7','#166534'], ['#f3e8ff','#7c3aed'], ['#fef3c7','#92400e'], ['#fce7f3','#9d174d'], ['#e0f2fe','#0369a1']];
  const idx = name ? name.charCodeAt(0) % palettes.length : 0;
  return palettes[idx];
}

function renderPost(post, liked) {
  const container = getPostContainer();
  if (!container) return;

  const title = post.title || post.Title || 'Sans titre';
  const content = post.content || post.Content || '';
  const author = post.author || post.Author || 'Anonyme';
  const user_id = post.user_id || post.UserID || 0;
  const created_at = post.created_at || post.CreatedAt;
  const category = post.category || post.Category || '';
  const likes_count = post.likes_count !== undefined ? post.likes_count : (post.LikesCount || 0);
  const comments_count = post.comments_count !== undefined ? post.comments_count : (post.CommentsCount || 0);

  document.title = `Zukuk — ${title}`;
  const isOwner = currentUser && currentUser.id === user_id;
  const [bgColor, textColor] = avatarColors(author);

  const avatar = post.avatar_url || post.AvatarURL || '';
  const cleanAvatar = avatar.startsWith('http') ? avatar : (window.location.hostname === 'localhost' ? 'http://localhost:8081' : '') + avatar;
  const avatarHTML = avatar
    ? `<img src="${cleanAvatar}" style="width:48px;height:48px;border-radius:50%;object-fit:cover" alt="avatar">`
    : `<div style="width:48px;height:48px;border-radius:50%;background:${bgColor};color:${textColor};display:flex;align-items:center;justify-content:center;font-weight:700;font-size:16px;flex-shrink:0">${initials(author)}</div>`;

  let fullDate = "Date non définie";
  let timeAgoStr = "";
  if (created_at) {
    try {
      fullDate = new Date(created_at).toLocaleDateString('fr-FR', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });
      timeAgoStr = timeAgo(created_at);
    } catch(_) {}
  }

  container.innerHTML = `
    <div class="post-card">
      <div class="post-meta">
        <div class="post-avatar">${avatarHTML}</div>
        <div class="post-author-block">
          <div class="author-name">${escHtml(author)}</div>
          <div class="author-time" title="${fullDate}">${timeAgoStr}</div>
        </div>
        ${category ? `<span class="post-category" style="background:${tagColor(category)}">${escHtml(category)}</span>` : ''}
      </div>
      <h1 class="post-title">${escHtml(title)}</h1>
      <div class="post-content" style="white-space: pre-wrap;">${escHtml(content)}</div>
      <div class="post-actions-row">
        <button class="action-btn${liked ? ' liked' : ''}" id="likeBtn">
          <svg viewBox="0 0 24 24" fill="${liked ? 'currentColor' : 'none'}" stroke="currentColor"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>
          <span id="likeCount">${likes_count}</span>
        </button>
        <button class="action-btn" id="scrollToComments">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
          <span id="commentCount">${comments_count}</span>
        </button>
        ${isOwner ? `
          <button class="action-btn btn-edit" id="editPostBtn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg> Modifier</button>
          <button class="action-btn btn-delete" id="deletePostBtn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6l-1 14H6L5 6"></path><path d="M10 11v6M14 11v6"></path><path d="M9 6V4h6v2"></path></svg> Supprimer</button>` : ''}
      </div>
    </div>`;

  document.getElementById('likeBtn')?.addEventListener('click', handleLike);
  document.getElementById('scrollToComments')?.addEventListener('click', () => {
     const anchor = getCommentsContainer();
     anchor?.scrollIntoView({ behavior:'smooth' });
  });
  if (isOwner) {
    document.getElementById('editPostBtn')?.addEventListener('click', openEditPostModal);
    document.getElementById('deletePostBtn')?.addEventListener('click', handleDeletePost);
  }
}

async function handleLike() {
  if (!currentUser) { showAuthToast(); return; }
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/like`, { method: 'POST', credentials: 'include' });
    if (!res.ok) return;
    const data = await res.json();
    const btn = document.getElementById('likeBtn');
    if(btn) {
      btn.classList.toggle('liked', data.liked);
      btn.querySelector('svg').setAttribute('fill', data.liked ? 'currentColor' : 'none');
      document.getElementById('likeCount').textContent = data.count;
    }
  } catch (err) {}
}

function openEditPostModal() {
  if (!postData) return;
  const overlay = document.createElement('div');
  overlay.style.cssText = `position:fixed;inset:0;background:rgba(0,0,0,0.35);backdrop-filter:blur(4px);z-index:1000;display:flex;align-items:center;justify-content:center;padding:16px`;
  overlay.id = 'editModal';
  const titleVal = escHtml(postData.title || postData.Title);
  const contentVal = escHtml(postData.content || postData.Content);
  
  overlay.innerHTML = `
    <div style="background:white;border-radius:24px;padding:32px;width:100%;max-width:540px;box-shadow:0 20px 60px rgba(0,0,0,0.15);position:relative">
      <button onclick="document.getElementById('editModal').remove()" style="position:absolute;top:14px;right:16px;background:none;border:none;font-size:22px;cursor:pointer;color:#9ca3af">✕</button>
      <h3 style="font-size:1.2rem;font-weight:700;color:#1e293b;margin-bottom:20px">Modifier le post</h3>
      <div class="modal-field" style="margin-bottom:12px;"><label style="display:block;margin-bottom:6px;font-weight:bold;">Titre</label><input id="edit-title" class="modern-input" style="width:100%;padding:10px;border-radius:8px;border:1px solid #ccc;" value="${titleVal}"></div>
      <div class="modal-field" style="margin-bottom:12px;"><label style="display:block;margin-bottom:6px;font-weight:bold;">Contenu</label><textarea id="edit-content" class="modern-input" style="width:100%;padding:10px;border-radius:8px;border:1px solid #ccc;" rows="6">${contentVal}</textarea></div>
      <div style="display:flex;gap:12px;margin-top:16px">
        <button onclick="document.getElementById('editModal').remove()" style="flex:1;padding:12px;border:2px solid #e2e8f0;border-radius:12px;background:white;font-weight:600;cursor:pointer;">Annuler</button>
        <button id="saveEditPost" style="flex:2;padding:12px;background:#818cf8;color:white;border:none;border-radius:12px;font-weight:600;cursor:pointer;">Enregistrer</button>
      </div>
    </div>`;
  document.body.appendChild(overlay);
  overlay.addEventListener('click', e => { if (e.target === overlay) overlay.remove(); });

  document.getElementById('saveEditPost').addEventListener('click', async () => {
    const title   = document.getElementById('edit-title').value.trim();
    const content = document.getElementById('edit-content').value.trim();
    if (title.length < 3 || content.length < 10) { showToast("Contenu trop court.", true); return; }
    
    const btn = document.getElementById('saveEditPost');
    btn.disabled = true; btn.textContent = "Sauvegarde...";

    try {
      const res = await fetch(`${API_BASE}/posts/${postId}`, { method: 'PUT', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ title, content }) });
      if (!res.ok) { showToast("Erreur.", true); btn.disabled = false; btn.textContent = "Enregistrer"; return; }
      
      document.querySelector('.post-title').textContent = title;
      document.querySelector('.post-content').textContent = content;
      postData.title = title; postData.content = content;
      overlay.remove();
      showToast("✏️ Modification enregistrée !", false);
    } catch (err) { showToast("Serveur injoignable", true); btn.disabled = false; }
  });
}

async function handleDeletePost() {
    showConfirmModal("Supprimer ce post ?", "Cette action est irréversible.", async () => {
        try {
            const res = await fetch(`${API_BASE}/posts/${postId}`, { method:'DELETE', credentials:'include' });
            if (res.ok) {
                showToast("🗑️ Post supprimé.", false);
                setTimeout(() => window.location.href = '/board', 1000);
            } else { showToast("Erreur lors de la suppression.", true); }
        } catch (err) { showToast("Serveur injoignable.", true); }
    });
}

// ── COMMENTAIRES ───────────────────────────────────────────────────────────

function renderComments(comments) {
  const container = getCommentsContainer();
  if (!container) return;

  if (comments.length === 0) {
    container.innerHTML = `<div class="comments-section"><p style="color:#9ca3af;font-size:0.9rem;text-align:center;padding:20px 0">Aucun commentaire pour l'instant.</p></div>`;
    return;
  }

  const html = comments.map(cm => buildCommentHTML(cm)).join('');
  container.innerHTML = `<div class="comments-section"><div class="comments-header">Commentaires (<span class="comments-count-badge">${comments.length}</span>)</div><div id="commentsList">${html}</div></div>`;
  attachCommentListeners();
  updateCommentCount(comments.length);
}

function buildCommentHTML(cm) {
  const author = cm.author || cm.Author || 'Anonyme';
  const content = cm.content || cm.Content || '';
  const created_at = cm.created_at || cm.CreatedAt;
  const isOwner = currentUser && currentUser.id === (cm.user_id || cm.UserID);
  const [bgColor, textColor] = avatarColors(author);

  const avatar = cm.avatar_url || cm.AvatarURL || '';
  const cleanAvatar = avatar.startsWith('http') ? avatar : (window.location.hostname === 'localhost' ? 'http://localhost:8081' : '') + avatar;
  const avatarHTML = avatar
    ? `<img src="${cleanAvatar}" style="width:100%;height:100%;object-fit:cover;border-radius:50%" alt="avatar">`
    : `<div style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:${bgColor};color:${textColor};font-weight:700;font-size:11px;border-radius:50%">${initials(author)}</div>`;

  return `
    <div class="comment-card" data-comment-id="${cm.id}">
      <div class="comment-meta">
        <div class="comment-avatar">${avatarHTML}</div>
        <span class="comment-author">${escHtml(author)}</span>
        <span class="comment-time">• ${timeAgo(created_at)}</span>
        ${isOwner ? `
          <div class="comment-owner-btns">
            <button class="comment-owner-btn edit-c" data-id="${cm.id}">✏️</button>
            <button class="comment-owner-btn delete-c" data-id="${cm.id}">🗑️</button>
          </div>` : ''}
      </div>
      <div class="comment-content" id="cm-content-${cm.id}">${escHtml(content)}</div>
    </div>`;
}

function attachCommentListeners() {
  document.querySelectorAll('.edit-c').forEach(btn => btn.addEventListener('click', () => openEditComment(btn.dataset.id)));
  document.querySelectorAll('.delete-c').forEach(btn => btn.addEventListener('click', () => handleDeleteComment(btn.dataset.id)));
}

function updateCommentCount(n) {
  const badge = document.querySelector('.comments-count-badge');
  if (badge) badge.textContent = n;
  const countEl = document.getElementById('commentCount');
  if (countEl) countEl.textContent = n;
}

// 🔴 CORRECTION DU COMPORTEMENT ANONYME PAR DÉFAUT
function renderCommentForm() {
  const container = getCommentFormContainer();
  if (!container) return;
  if (!currentUser) {
    container.innerHTML = `<div class="auth-nudge">Connecte-toi pour laisser un commentaire.</div>`;
    return;
  }

  // Lecture du réglage récupéré plus haut
  const isAnon = userSettings && userSettings.anon_by_default;
  const cbBg = isAnon ? '#818cf8' : 'white';
  const cbBorder = isAnon ? '#818cf8' : '#d1d5db';
  const svgOpacity = isAnon ? '1' : '0';

  container.innerHTML = `
    <div class="comment-form-card">
      <textarea class="comment-textarea" id="commentInput" rows="3" placeholder="Laisse un commentaire de soutien..." style="width:100%; padding:10px; border-radius:8px; border:1px solid #ccc;"></textarea>
      <div class="comment-form-footer" style="margin-top:10px; display:flex; justify-content:space-between; align-items:center;">
        <label class="anon-label" style="display:flex; align-items:center; gap:8px; cursor:pointer; font-size:0.9rem; color:#64748b;">
            <input type="checkbox" id="commentAnon" style="display:none;" ${isAnon ? 'checked' : ''}>
            <div class="anon-checkbox" style="width:18px; height:18px; border-radius:6px; border:2px solid ${cbBorder}; background:${cbBg}; display:flex; align-items:center; justify-content:center; transition:all 0.2s;">
                <svg viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="width:12px; height:12px; opacity:${svgOpacity}; transition:opacity 0.2s;">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
            </div> 
            Anonyme
        </label>
        <button class="submit-comment-btn" id="submitComment" style="padding:8px 16px; background:#818cf8; color:white; border:none; border-radius:8px; cursor:pointer;">Publier</button>
      </div>
    </div>`;
  
  const anonCheckbox = document.getElementById('commentAnon');
  if (anonCheckbox) {
      anonCheckbox.addEventListener('change', function() {
        const cb = this.nextElementSibling;
        cb.style.background  = this.checked ? '#818cf8' : 'white';
        cb.style.borderColor = this.checked ? '#818cf8' : '#d1d5db';
        cb.querySelector('svg').style.opacity = this.checked ? '1' : '0';
      });
  }

  document.getElementById('submitComment')?.addEventListener('click', handleAddComment);
}

async function handleAddComment() {
  const content = document.getElementById('commentInput').value.trim();
  // 🔴 CORRECTION IMPORTANTE : on lit bien l'état de la case au lieu de forcer '0'
  const anon = document.getElementById('commentAnon').checked ? 1 : 0;

  if (!content) { showToast('Le commentaire est vide.', true); return; }

  const btn = document.getElementById('submitComment');
  btn.disabled = true; btn.textContent = 'Publication…';

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ content, is_anonymous: anon }),
    });
    const newComment = await res.json();
    if (!res.ok) { showToast("Erreur serveur.", true); btn.disabled = false; btn.textContent = 'Publier'; return; }

    let list = document.getElementById('commentsList');
    if (!list) {
      getCommentsContainer().innerHTML = `<div class="comments-section"><div class="comments-header">Commentaires (<span class="comments-count-badge">1</span>)</div><div id="commentsList"></div></div>`;
      list = document.getElementById('commentsList');
    }

    const newEl = document.createElement('div');
    newEl.innerHTML = buildCommentHTML(newComment);
    list.appendChild(newEl.firstElementChild);
    attachCommentListeners();

    const currentCount = parseInt(document.querySelector('.comments-count-badge')?.textContent || '0', 10);
    updateCommentCount(currentCount + 1);

    document.getElementById('commentInput').value = '';
    showToast("✨ Commentaire ajouté !", false);
  } catch (err) { showToast('Erreur lors de la publication.', true); } 
  finally { btn.disabled = false; btn.textContent = 'Publier'; }
}

function openEditComment(commentId) {
  const contentEl = document.getElementById(`cm-content-${commentId}`);
  if (!contentEl) return;
  const original = contentEl.textContent;

  contentEl.innerHTML = `<textarea id="edit-area-${commentId}" rows="3" style="width:100%;">${escHtml(original)}</textarea><div><button id="save-cm-${commentId}">Enregistrer</button> <button id="cancel-cm-${commentId}">Annuler</button></div>`;

  document.getElementById(`save-cm-${commentId}`).addEventListener('click', async () => {
      const content = document.getElementById(`edit-area-${commentId}`).value.trim();
      if (!content) return;
      try {
        const res = await fetch(`${API_BASE}/comments/${commentId}`, { method: 'PUT', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ content }) });
        if (res.ok) { contentEl.textContent = content; showToast("Modifié", false); }
      } catch (e) {}
  });
  document.getElementById(`cancel-cm-${commentId}`).addEventListener('click', () => { contentEl.textContent = original; });
}

async function handleDeleteComment(commentId) {
    showConfirmModal("Supprimer ce commentaire ?", "", async () => {
        try {
            const res = await fetch(`${API_BASE}/comments/${commentId}`, { method: 'DELETE', credentials: 'include' });
            if (res.ok) {
                document.querySelector(`.comment-card[data-comment-id="${commentId}"]`)?.remove();
                const currentCount = parseInt(document.querySelector('.comments-count-badge')?.textContent || '1', 10);
                updateCommentCount(Math.max(0, currentCount - 1));
                showToast("Supprimé", false);
            }
        } catch (err) {}
    });
}

// ── UTILITAIRES (Toasts, Modales, Erreurs) ──────────────────────────────────

function showToast(msg, isError = false) {
    const existing = document.getElementById('post-toast');
    if (existing) existing.remove();
    const toast = document.createElement('div');
    toast.id = 'post-toast';
    toast.textContent = msg;
    toast.style.cssText = `position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%) translateY(100px); background: ${isError ? '#fef2f2' : '#f0fdf4'}; color: ${isError ? '#ef4444' : '#166534'}; border: 1px solid ${isError ? '#fecaca' : '#bbf7d0'}; padding: 14px 28px; border-radius: 16px; font-weight: 600; font-size: 0.95rem; box-shadow: 0 10px 30px rgba(0,0,0,0.1); z-index: 9999; transition: transform 0.4s cubic-bezier(0.2, 0.9, 0.2, 1);`;
    document.body.appendChild(toast);
    requestAnimationFrame(() => toast.style.transform = 'translateX(-50%) translateY(0)');
    setTimeout(() => { toast.style.transform = 'translateX(-50%) translateY(100px)'; setTimeout(() => toast.remove(), 400); }, 3500);
}

function showAuthToast() { showToast("🔒 Connecte-toi pour interagir !", true); }

function showConfirmModal(title, message, onConfirm) {
    const existing = document.getElementById('zukuk-confirm');
    if (existing) existing.remove();
    const overlay = document.createElement('div');
    overlay.id = 'zukuk-confirm';
    overlay.style.cssText = `position: fixed; inset: 0; background: rgba(15, 23, 42, 0.4); backdrop-filter: blur(8px); z-index: 10000; display: flex; align-items: center; justify-content: center; padding: 20px;`;
    overlay.innerHTML = `
        <div style="background: white; border-radius: 24px; padding: 32px; width: 100%; max-width: 400px; text-align: center;">
            <h3 style="font-size: 1.25rem; font-weight: 700; margin-bottom: 12px;">${title}</h3>
            <p style="color: #64748b; margin-bottom: 28px;">${message}</p>
            <div style="display: flex; gap: 12px;">
                <button id="cancel-confirm" style="flex: 1; padding: 12px; border-radius: 14px; border: 1px solid #ccc; background: white; cursor: pointer;">Annuler</button>
                <button id="accept-confirm" style="flex: 1; padding: 12px; border: none; border-radius: 14px; background: #ef4444; color: white; cursor: pointer;">Confirmer</button>
            </div>
        </div>`;
    document.body.appendChild(overlay);
    document.getElementById('cancel-confirm').onclick = () => overlay.remove();
    document.getElementById('accept-confirm').onclick = () => { overlay.remove(); onConfirm(); };
}

function renderError(msg, debugInfo = "") {
  const container = getPostContainer();
  if (!container) return;
  container.innerHTML = `
    <div style="text-align:center;padding:60px 20px; font-family: sans-serif;">
      <div style="font-size:3rem;margin-bottom:16px">😕</div>
      <h2 style="color:#1f2937;margin-bottom:8px; font-size: 1.4rem;">${escHtml(msg)}</h2>
      ${debugInfo ? `<pre style="background:#f8fafc; padding:16px; border-radius:12px; color:#ef4444; font-size:0.8rem; text-align:left; max-width:100%; overflow-x:auto; margin-top:20px; border:1px solid #fee2e2; font-family:monospace;">[Rapport d'anomalie] :\n${escHtml(debugInfo)}</pre>` : ''}
      <a href="/board" style="color:#818cf8;font-weight:600;text-decoration:none; display:inline-block; margin-top: 24px;">← Retour au forum</a>
    </div>`;
}

function escHtml(str) {
  if (!str || typeof str !== 'string') return '';
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#039;');
}