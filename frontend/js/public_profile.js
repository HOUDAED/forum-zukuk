// public_profile.js — Profil public d'un membre (Version Polymorphe 🛡️ + Bio)
const API_BASE = '/api'; 
const fetchOpts = { credentials: 'include' };

document.addEventListener('DOMContentLoaded', async () => {
    const parts = window.location.pathname.split('/').filter(p => p !== "");
    const userId = parseInt(parts[parts.length - 1], 10);
    
    if (!userId || isNaN(userId)) { 
        renderError('ID d\'utilisateur invalide.'); 
        return; 
    }

    try {
        const res = await fetch(`${API_BASE}/network/${userId}`, fetchOpts);
        
        if (!res.ok) {
            renderError('Membre introuvable ou profil privé (Code: ' + res.status + ')');
            return;
        }
        
        const data = await res.json();
        const user = data.user || data.member || data;
        
        renderProfile(user);
        
    } catch (err) {
        console.error("🔥 Crash JS au chargement :", err);
        renderError('Erreur lors du chargement des données : ' + err.message);
    }
});

// ── UTILITAIRES ─────────────────────────────────────────────────────────────

function escHtml(str) {
    if (!str || typeof str !== 'string') return '';
    return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#039;');
}

function timeAgo(dateStr) {
    if (!dateStr) return 'Récemment';
    const diff = (Date.now() - new Date(dateStr)) / 1000;
    if (isNaN(diff)) return 'Récemment';
    if (diff < 60) return 'À l\'instant';
    if (diff < 3600) return `${Math.floor(diff/60)} min`;
    if (diff < 86400) return `${Math.floor(diff/3600)} h`;
    return new Date(dateStr).toLocaleDateString('fr-FR');
}

const MOOD_EMOJIS = { 'Bien': '😊', 'Calme': '😌', 'Triste': '😢', 'Anxieux': '😰', 'Colère': '😡', 'Colere': '😡' };
const MOOD_CLASSES = { 'Bien': 'mood-bien', 'Calme': 'mood-calme', 'Triste': 'mood-triste', 'Anxieux': 'mood-anxieux', 'Colère': 'mood-colère', 'Colere': 'mood-colère' };

function avatarColors(name) {
    const palettes = [['#dbeafe','#1d4ed8'], ['#dcfce7','#166534'], ['#f3e8ff','#7c3aed'], ['#fef3c7','#92400e'], ['#fce7f3','#9d174d'], ['#e0f2fe','#0369a1']];
    const idx = name ? name.charCodeAt(0) % palettes.length : 0;
    return palettes[idx];
}

function initials(name) { 
    return (name || '?').substring(0, 2).toUpperCase(); 
}

// ── MOTEUR D'AFFICHAGE ──────────────────────────────────────────────────────

function renderProfile(user) {
    const container = document.getElementById('publicProfileContainer');
    if (!container) return;

    try {
        // Tolérance majuscules/minuscules sur les clés renvoyées par Go
        const pseudo = user.pseudo || user.Pseudo || 'Anonyme';
        const [bgColor, textColor] = avatarColors(pseudo);
        
        let avatarHTML = '';
        const avatarRaw = user.avatar_url || user.AvatarURL || '';
        if (avatarRaw) {
            const cleanAvatar = avatarRaw.startsWith('http') ? avatarRaw : (window.location.hostname === 'localhost' ? 'http://localhost:8081' : '') + avatarRaw;
            avatarHTML = `<img src="${cleanAvatar}" alt="avatar">`;
        } else {
            avatarHTML = `<div style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:${bgColor};color:${textColor};font-weight:700;font-size:28px;">${initials(pseudo)}</div>`;
        }

        const currentMood = user.current_mood || user.CurrentMood || user.mood || '';
        const moodEmoji = MOOD_EMOJIS[currentMood] || '';
        const moodClass = MOOD_CLASSES[currentMood] || 'mood-default';
        const moodHTML = currentMood 
            ? `<div class="hero-mood ${moodClass}">${moodEmoji} ${escHtml(currentMood)}</div>`
            : `<div class="hero-mood mood-default">Aucune humeur partagée</div>`;

        // 🔴 RÉINTÉGRATION DE LA BIO ICI
        const bioText = user.bio || user.Bio || '';
        const bioHTML = bioText
            ? `<p style="font-size:1rem; color:var(--text-secondary); margin-bottom:20px; line-height: 1.6; font-style: italic;">"${escHtml(bioText)}"</p>`
            : '';

        const posts = Array.isArray(user.posts) ? user.posts : (Array.isArray(user.Posts) ? user.Posts : []);
        const comments = Array.isArray(user.comments) ? user.comments : (Array.isArray(user.Comments) ? user.Comments : []);

        // Génération adaptative des posts
        let postsHTML = posts.length > 0 
            ? posts.map(p => `
                <a href="/post/${p.id || p.ID}" class="public-post-item">
                    <div class="post-item-title">${escHtml(p.title || p.Title)}</div>
                    <div class="post-item-meta">
                        ${(p.category || p.Category) ? `<span class="post-item-tag" style="background:var(--primary-color)">${escHtml(p.category || p.Category)}</span>` : ''}
                        <span class="post-item-stat">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg> 
                            ${p.likes_count !== undefined ? p.likes_count : (p.LikesCount || 0)}
                        </span>
                        <span class="post-item-stat">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg> 
                            ${p.comments_count !== undefined ? p.comments_count : (p.CommentsCount || 0)}
                        </span>
                        <span class="post-item-date">${timeAgo(p.created_at || p.CreatedAt)}</span>
                    </div>
                </a>
            `).join('')
            : `<div class="empty">Aucun post partagé pour l'instant.</div>`;

        // Génération adaptative des commentaires
        let commentsHTML = comments.length > 0
            ? comments.map(c => `
                <div class="public-comment-item">
                    <div class="comment-on">Discussion : <a href="/post/${c.post_id || c.PostID || ''}">${escHtml(c.post_title || c.PostTitle || 'Voir le post')}</a></div>
                    <div class="comment-text">"${escHtml(c.content || c.Content)}"</div>
                    <div class="post-item-date" style="margin-top:4px;">${timeAgo(c.created_at || c.CreatedAt)}</div>
                </div>
            `).join('')
            : `<div class="empty">Aucun commentaire rédigé.</div>`;

        // Récupération des compteurs globaux
        const pCount = user.posts_count !== undefined ? user.posts_count : (user.PostsCount !== undefined ? user.PostsCount : posts.length);
        const cCount = user.comments_count !== undefined ? user.comments_count : (user.CommentsCount !== undefined ? user.CommentsCount : comments.length);
        const lCount = user.likes_given !== undefined ? user.likes_given : (user.LikesGiven !== undefined ? user.LikesGiven : 0);

        container.innerHTML = `
            <div class="public-profile-wrapper">
                <div class="hero-card">
                    <div class="hero-avatar">${avatarHTML}</div>
                    <div style="flex:1">
                        <div class="hero-name">${escHtml(pseudo)}</div>
                        ${moodHTML}
                        ${bioHTML} <div class="hero-stats">
                            <div class="hero-stat">
                                <span class="hero-stat-value">${pCount}</span>
                                <span class="hero-stat-label">Posts</span>
                            </div>
                            <div class="hero-stat">
                                <span class="hero-stat-value">${cCount}</span>
                                <span class="hero-stat-label">Réponses</span>
                            </div>
                            <div class="hero-stat">
                                <span class="hero-stat-value">${lCount}</span>
                                <span class="hero-stat-label">Likes donnés</span>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="public-bento">
                    <div style="display:flex; flex-direction:column; gap:16px;">
                        <h3 style="font-size:1.15rem; font-weight:700; color:var(--text-primary);">Posts récents</h3>
                        <div style="display:flex; flex-direction:column; gap:12px;">
                            ${postsHTML}
                        </div>
                    </div>
                    <div style="display:flex; flex-direction:column; gap:16px;">
                        <h3 style="font-size:1.15rem; font-weight:700; color:var(--text-primary);">Derniers commentaires</h3>
                        <div style="display:flex; flex-direction:column; gap:12px;">
                            ${commentsHTML}
                        </div>
                    </div>
                </div>
            </div>
        `;
    } catch (renderErr) {
        console.error("🔥 Crash JS au rendu HTML :", renderErr);
        renderError('Erreur d\'affichage de la structure du profil : ' + renderErr.message);
    }
}

function renderError(msg) {
    document.getElementById('publicProfileContainer').innerHTML = `
        <div style="text-align:center;padding:60px 20px; font-family: sans-serif;">
          <div style="font-size:3rem;margin-bottom:16px">😕</div>
          <h2 style="color:var(--text-primary);margin-bottom:8px; font-size: 1.4rem;">${escHtml(msg)}</h2>
          <a href="/network" style="color:var(--primary-color);font-weight:600;text-decoration:none; display:inline-block; margin-top: 24px;">← Retour à la communauté</a>
        </div>
    `;
}