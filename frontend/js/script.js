// Si on est en ligne, on utilise un chemin relatif (vide).
const API_BASE = '';
const $ = (selector) => document.querySelector(selector);

const EYE_ICON     = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>`;
const EYE_OFF_ICON = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/><path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/><line x1="1" y1="1" x2="23" y2="23"/></svg>`;

// ─────────────────────────────────────────────────────────────────────────────
// 🎨 SYSTÈME DE THÈME GLOBAL ZUKUK
// ─────────────────────────────────────────────────────────────────────────────

const savedTheme = localStorage.getItem('zukuk_theme') || 'glass';
document.documentElement.setAttribute('data-theme', savedTheme);

document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch(`${API_BASE}/api/settings/`, { credentials: 'include' });
        if (res.ok) {
            const settings = await res.json();
            if (settings.theme && settings.theme !== savedTheme) {
                document.documentElement.setAttribute('data-theme', settings.theme);
                localStorage.setItem('zukuk_theme', settings.theme);
            }
        }
    } catch (e) {}
});

function wrapInlinePasswords() {
    document.querySelectorAll('input[type="password"].inline-input').forEach(input => {
        if (input.parentElement.classList.contains('password-wrapper')) return;
        const wrapper = document.createElement('span');
        wrapper.className = 'password-wrapper';
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'toggle-password';
        btn.innerHTML = EYE_ICON;
        btn.addEventListener('click', () => {
            const show = input.type === 'password';
            input.type = show ? 'text' : 'password';
            btn.innerHTML = show ? EYE_OFF_ICON : EYE_ICON;
        });
        input.parentNode.insertBefore(wrapper, input);
        wrapper.appendChild(input);
        wrapper.appendChild(btn);
    });
}

function wrapModernPasswords() {
    document.querySelectorAll('input[type="password"].modern-input').forEach(input => {
        if (input.parentElement.classList.contains('modern-input-wrapper')) return;
        const wrapper = document.createElement('div');
        wrapper.className = 'modern-input-wrapper';
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'modern-toggle-password';
        btn.innerHTML = EYE_ICON;
        btn.addEventListener('click', () => {
            const show = input.type === 'password';
            input.type = show ? 'text' : 'password';
            btn.innerHTML = show ? EYE_OFF_ICON : EYE_ICON;
        });
        input.parentNode.insertBefore(wrapper, input);
        wrapper.appendChild(input);
        wrapper.appendChild(btn);
    });
}

document.addEventListener('DOMContentLoaded', () => {
    wrapInlinePasswords();
    wrapModernPasswords();
});

const dynamicInputs = document.querySelectorAll(".dynamic-width");
if (dynamicInputs.length > 0) {
    const canvasContext = document.createElement("canvas").getContext("2d");
    canvasContext.font = "600 1.4rem 'Outfit', sans-serif";
    dynamicInputs.forEach((input) => {
        const adjustWidth = function () {
            let text = this.value || this.placeholder;
            if (this.type === "password" && this.value) text = "•".repeat(this.value.length);
            const metrics = canvasContext.measureText(text);
            this.style.width = `${Math.max(160, metrics.width + 18)}px`;
        };
        input.addEventListener("input", adjustWidth);
        adjustWidth.call(input);
    });
}

function showMessage(element, text, isError = false) {
    if (!element) return;
    element.hidden = false;
    element.textContent = text;
    element.style.borderColor = isError ? "var(--danger-color)" : "var(--success-color)";
    element.style.background  = isError ? "rgba(239, 68, 68, 0.1)" : "rgba(16, 185, 129, 0.1)";
    element.style.color       = isError ? "var(--danger-color)" : "var(--success-color)";
}

async function api(path, options = {}) {
    const response = await fetch(`${API_BASE}${path}`, {
        credentials: "include",
        headers: { "Content-Type": "application/json", ...(options.headers || {}) },
        ...options,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Erreur serveur.");
    return data;
}

async function uploadFile(path, formData) {
    const response = await fetch(`${API_BASE}${path}`, {
        method: "POST", credentials: "include", body: formData,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Erreur upload.");
    return data;
}

function initHome() {
    $("#home-login")?.addEventListener("click", () => { window.location.href = "/login"; });
    $("#home-join")?.addEventListener("click", () => { window.location.href = "/register"; });
}

function initLogin() {
    const button = $("#login-submit");
    if (!button) return;
    const submit = async () => {
        const message = $("#login-message");
        try {
            await api("/api/auth/login", {
                method: "POST",
                body: JSON.stringify({
                    identifier: $("#login-identifier")?.value?.trim() || "",
                    password:   $("#login-password")?.value || "",
                }),
            });
            showMessage(message, "Connexion réussie.");
            window.location.href = "/board";
        } catch (error) {
            showMessage(message, error.message, true);
        }
    };
    button.addEventListener("click", submit);
    $("#login-password")?.addEventListener("keydown", (e) => { if (e.key === "Enter") submit(); });
}

function initRegister() {
    const button = $("#register-submit");
    if (!button) return;
    button.addEventListener("click", async () => {
        const message = $("#register-message");
        try {
            await api("/api/auth/register", {
                method: "POST",
                body: JSON.stringify({
                    pseudo:          $("#register-pseudo")?.value?.trim()  || "",
                    email:           $("#register-email")?.value?.trim()   || "",
                    password:        $("#register-password")?.value        || "",
                    confirmPassword: $("#register-confirm")?.value         || "",
                }),
            });
            showMessage(message, "Compte créé. Tu peux maintenant te connecter.");
            setTimeout(() => { window.location.href = "/login"; }, 800);
        } catch (error) {
            showMessage(message, error.message, true);
        }
    });
}

async function loadProfile() {
    const me = await api("/api/me");
    const pseudoInput = $("#profile-pseudo");
    const emailInput  = $("#profile-email");
    const bioInput    = $("#profile-bio");

    if (pseudoInput) pseudoInput.value = me.pseudo || "";
    if (emailInput)  emailInput.value  = me.email  || "";
    if (bioInput)    bioInput.value    = me.bio    || "";

    const avatar = $("#profile-avatar");
    if (avatar) {
        if (me.avatar_url) {
            avatar.src = `${API_BASE}${me.avatar_url.replace('/api', '')}`;
        } else {
            const seed = me.pseudo || "ZukukUser";
            avatar.src = `https://api.dicebear.com/7.x/notionists/svg?seed=${seed}&backgroundColor=eef2ff`;
        }
        avatar.hidden = false;
    }

    const connectionsList = $("#connections-list");
    if (connectionsList) {
        try {
            const history = await api("/api/me/connections");
            connectionsList.innerHTML = "";
            if (history && history.length > 0) {
                history.forEach((conn, index) => {
                    const date = new Date(conn.created_at).toLocaleString('fr-FR', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }).replace('.', '');
                    let device = "Appareil inconnu";
                    if (conn.user_agent.includes("Windows")) device = "PC Windows";
                    else if (conn.user_agent.includes("Mac")) device = "Mac";
                    else if (conn.user_agent.includes("Linux")) device = "Linux";
                    else if (conn.user_agent.includes("Android")) device = "Android";
                    else if (conn.user_agent.includes("iPhone") || conn.user_agent.includes("iPad")) device = "iPhone";

                    const statusBadge = conn.status === "Réussie"
                        ? '<span class="badge badge-success" style="background:rgba(16,185,129,0.1); color:var(--success-color);">Réussie</span>'
                        : '<span class="badge badge-error" style="background:rgba(239,68,68,0.1); color:var(--danger-color);">Échouée</span>';
                    
                    const isLatest = (index === 0 && conn.status === "Réussie")
                        ? '<span style="color:var(--primary-color);font-weight:700;font-size:0.8rem;margin-left:auto;">Actuelle</span>'
                        : '';

                    // 🔴 Correction ici : on utilise --text-primary au lieu de --text-dark
                    connectionsList.innerHTML += `
                        <li class="bento-list-item">
                            <div style="display:flex;align-items:center;gap:12px;">
                                <strong style="color:var(--text-primary);">${device}</strong>
                                ${statusBadge}
                                ${isLatest}
                            </div>
                            <div style="display:flex;gap:16px;font-size:0.85rem;color:var(--text-secondary);margin-top:4px;">
                                <span style="display:flex;align-items:center;gap:4px;">
                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
                                    ${conn.ip_address}
                                </span>
                                <span style="display:flex;align-items:center;gap:4px;">
                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                                    ${date}
                                </span>
                            </div>
                        </li>`;
                });
            } else {
                connectionsList.innerHTML = `<li style="color:var(--text-secondary);">Aucun historique.</li>`;
            }
        } catch {
            connectionsList.innerHTML = `<li style="color:var(--danger-color);">Impossible de charger.</li>`;
        }
    }

    const tabPosts = $("#tab-posts");
    const tabComments = $("#tab-comments");
    const listPosts = $("#user-posts-list");
    const listComments = $("#user-comments-list");

    if (tabPosts && tabComments) {
        tabPosts.addEventListener("click", () => {
            tabPosts.classList.add("active");
            tabComments.classList.remove("active");
            listPosts.style.display = "flex";
            listComments.style.display = "none";
        });
        tabComments.addEventListener("click", () => {
            tabComments.classList.add("active");
            tabPosts.classList.remove("active");
            listComments.style.display = "flex";
            listPosts.style.display = "none";
        });
    }

    if (listPosts && listComments) {
        try {
            const activity = await api("/api/me/activity");
            listPosts.innerHTML = "";
            listComments.innerHTML = "";

            if (activity.posts && activity.posts.length > 0) {
                activity.posts.forEach(post => {
                    const date = new Date(post.created_at).toLocaleDateString('fr-FR');
                    // 🔴 Correction ici : on utilise --text-primary au lieu de --text-dark
                    listPosts.innerHTML += `
                        <li class="bento-list-item" style="flex-direction:row;justify-content:space-between;align-items:center;">
                            <div style="display:flex;flex-direction:column;gap:4px;">
                                <strong style="color:var(--text-primary);">${post.title}</strong>
                                <span style="font-size:0.8rem;color:var(--text-secondary);">Publié le ${date}</span>
                            </div>
                        </li>`;
                });
            } else {
                listPosts.innerHTML = `<li style="color:var(--text-secondary);">Aucune discussion publiée.</li>`;
            }

            if (activity.comments && activity.comments.length > 0) {
                activity.comments.forEach(comment => {
                    const date = new Date(comment.created_at).toLocaleDateString('fr-FR');
                    // 🔴 Correction ici : on utilise --text-primary au lieu de --text-dark
                    listComments.innerHTML += `
                        <li class="bento-list-item">
                            <span style="font-size:0.8rem;color:var(--primary-color);font-weight:600;">Sur : ${comment.post_title}</span>
                            <p style="font-size:0.95rem;color:var(--text-primary);margin:4px 0;">"${comment.content}"</p>
                            <span style="font-size:0.8rem;color:var(--text-secondary);">Le ${date}</span>
                        </li>`;
                });
            } else {
                listComments.innerHTML = `<li style="color:var(--text-secondary);">Aucun commentaire écrit.</li>`;
            }
        } catch {
            listPosts.innerHTML    = `<li style="color:var(--danger-color);">Impossible de charger.</li>`;
            listComments.innerHTML = `<li style="color:var(--danger-color);">Impossible de charger.</li>`;
        }
    }
    return me;
}

function initProfile() {
    const saveButton = $("#profile-save");
    if (!saveButton) return;

    loadProfile().catch(() => { window.location.href = "/login"; });

    saveButton.addEventListener("click", async () => {
        const message = $("#profile-message");
        const payload = {};

        const pseudo = $("#profile-pseudo")?.value?.trim();
        if (pseudo) payload.pseudo = pseudo;

        const email = $("#profile-email")?.value?.trim();
        if (email) payload.email = email;

        const bio = $("#profile-bio")?.value?.trim();
        if (bio !== undefined) payload.bio = bio; 

        const password = $("#profile-password")?.value;
        if (password) {
            payload.currentPassword = $("#profile-current-password")?.value || "";
            payload.password        = password;
            payload.confirmPassword = $("#profile-confirm-password")?.value || "";
        }

        try {
            if (Object.keys(payload).length > 0) {
                await api("/api/me", { method: "PUT", body: JSON.stringify(payload) });
            }

            const file = $("#profile-avatar-file")?.files?.[0];
            if (file) {
                if (file.size > 2 * 1024 * 1024) throw new Error("La photo ne doit pas dépasser 2 Mo.");
                const formData = new FormData();
                formData.append("avatar", file);
                const avatarData = await uploadFile("/api/me/avatar", formData);
                const avatarImg = $("#profile-avatar");
                if (avatarImg && avatarData.avatar_url) {
                    avatarImg.src    = `${API_BASE}${avatarData.avatar_url.replace('/api', '')}`;
                    avatarImg.hidden = false;
                }
            }

            showMessage(message, "Profil mis à jour.");
            await loadProfile();

            ["profile-current-password", "profile-password", "profile-confirm-password"].forEach(id => {
                const el = $(`#${id}`);
                if (el) el.value = "";
            });
            const fileInput = $("#profile-avatar-file");
            if (fileInput) fileInput.value = "";

        } catch (error) { showMessage(message, error.message, true); }
    });

    $("#profile-logout")?.addEventListener("click", async () => {
        await api("/api/auth/logout", { method: "POST" }).catch(() => {});
        window.location.href = "/login";
    });

    $("#profile-delete")?.addEventListener("click", () => { window.location.href = "/delete-account"; });
}

function initForgotPassword() {
    const button = $("#forgot-submit");
    if (!button) return;
    button.addEventListener("click", async () => {
        const message = $("#forgot-message");
        try {
            await api("/api/auth/forgot-password", {
                method: "POST",
                body: JSON.stringify({ email: $("#forgot-email")?.value?.trim() || "" }),
            });
            showMessage(message, "Si cet email existe, un lien a été envoyé. Vérifie tes spams.");
        } catch (error) { showMessage(message, error.message, true); }
    });
}

function initResetPassword() {
    const button = $("#reset-submit");
    if (!button) return;
    const token   = new URLSearchParams(window.location.search).get('token');
    const message = $("#reset-message");

    if (!token) {
        showMessage(message, "Le lien de réinitialisation est invalide.", true);
        button.disabled = true; button.style.opacity = "0.5"; button.style.cursor = "not-allowed";
        return;
    }

    button.addEventListener("click", async () => {
        try {
            await api("/api/auth/reset-password", {
                method: "POST",
                body: JSON.stringify({
                    token,
                    password:        $("#reset-password-input")?.value || "",
                    confirmPassword: $("#reset-confirm")?.value        || "",
                }),
            });
            showMessage(message, "Mot de passe mis à jour. Redirection...");
            setTimeout(() => { window.location.href = "/login"; }, 2500);
        } catch (error) { showMessage(message, error.message, true); }
    });
}

initHome();
initLogin(); 
initRegister(); 
initProfile(); 
initForgotPassword(); 
initResetPassword();