const API_BASE = 'http://localhost:8081';
const $ = (selector) => document.querySelector(selector);

const dynamicInputs = document.querySelectorAll(".dynamic-width");
if (dynamicInputs.length > 0) {
    const canvasContext = document.createElement("canvas").getContext("2d");
    canvasContext.font = "600 1.4rem 'Outfit', sans-serif";
    dynamicInputs.forEach((input) => {
        const adjustWidth = function () {
            let text = this.value || this.placeholder;
            if (this.type === "password" && this.value) {
                text = "•".repeat(this.value.length);
            }
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
    element.style.borderColor = isError ? "#fecaca" : "#bbf7d0";
    element.style.background  = isError ? "#fef2f2" : "#f0fdf4";
    element.style.color       = isError ? "#991b1b" : "#166534";
}

async function api(path, options = {}) {
    const response = await fetch(`${API_BASE}${path}`, {
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
            ...(options.headers || {}),
        },
        ...options,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Erreur serveur.");
    return data;
}

async function uploadFile(path, formData) {
    const response = await fetch(`${API_BASE}${path}`, {
        method: "POST",
        credentials: "include",
        body: formData,
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Erreur upload.");
    return data;
}

function initHome() {
    $("#home-login")?.addEventListener("click", () => {
        window.location.href = "/login";
    });
    $("#home-join")?.addEventListener("click", () => {
        window.location.href = "/register";
    });
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

    $("#login-password")?.addEventListener("keydown", (e) => {
        if (e.key === "Enter") submit();
    });
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

    if (pseudoInput) pseudoInput.value = me.pseudo || "";
    if (emailInput)  emailInput.value  = me.email  || "";

    const avatar = $("#profile-avatar");
    if (avatar) {
        if (me.avatar_url) {
            avatar.src    = `${API_BASE}${me.avatar_url}`;
            avatar.hidden = false;
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
                    const date = new Date(conn.created_at).toLocaleString('fr-FR', {
                        day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit'
                    }).replace('.', '');
                    
                    let device = "Appareil inconnu";
                    if (conn.user_agent.includes("Windows")) device = "PC Windows";
                    else if (conn.user_agent.includes("Mac")) device = "Mac";
                    else if (conn.user_agent.includes("Linux")) device = "Linux";
                    else if (conn.user_agent.includes("Android")) device = "Android";
                    else if (conn.user_agent.includes("iPhone") || conn.user_agent.includes("iPad")) device = "iPhone";

                    const statusBadge = conn.status === "Réussie"
                        ? '<span class="badge badge-success">Réussie</span>'
                        : '<span class="badge badge-error">Échouée</span>';

                    const isLatest = (index === 0 && conn.status === "Réussie") 
                        ? '<span style="color: var(--primary); font-weight: 700; font-size: 0.8rem; margin-left: auto;">Actuelle</span>' 
                        : '';

                    connectionsList.innerHTML += `
                        <li class="bento-list-item">
                            <div style="display: flex; align-items: center; gap: 12px;">
                                <strong style="color: var(--text-dark);">${device}</strong> 
                                ${statusBadge}
                                ${isLatest}
                            </div>
                            <div style="display: flex; gap: 16px; font-size: 0.85rem; color: var(--text-gray); margin-top: 4px;">
                                <span style="display: flex; align-items: center; gap: 4px;">
                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#ec4899" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                                    ${conn.ip_address}
                                </span>
                                <span style="display: flex; align-items: center; gap: 4px;">
                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#64748b" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                                    ${date}
                                </span>
                            </div>
                        </li>
                    `;
                });
            } else {
                connectionsList.innerHTML = `<li style="color: var(--text-gray);">Aucun historique.</li>`;
            }
        } catch (error) {
            connectionsList.innerHTML = `<li style="color: #991b1b;">Impossible de charger.</li>`;
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
                    listPosts.innerHTML += `
                        <li class="bento-list-item" style="flex-direction: row; justify-content: space-between; align-items: center;">
                            <div style="display: flex; flex-direction: column; gap: 4px;">
                                <strong style="color: var(--text-dark);">${post.title}</strong>
                                <span style="font-size: 0.8rem; color: var(--text-gray);">Publié le ${date}</span>
                            </div>
                        </li>
                    `;
                });
            } else {
                listPosts.innerHTML = `<li style="color: var(--text-gray);">Aucune discussion publiée.</li>`;
            }

            if (activity.comments && activity.comments.length > 0) {
                activity.comments.forEach(comment => {
                    const date = new Date(comment.created_at).toLocaleDateString('fr-FR');
                    listComments.innerHTML += `
                        <li class="bento-list-item">
                            <span style="font-size: 0.8rem; color: var(--primary); font-weight: 600;">Sur : ${comment.post_title}</span>
                            <p style="font-size: 0.95rem; color: var(--text-dark); margin: 0;">"${comment.content}"</p>
                            <span style="font-size: 0.8rem; color: var(--text-gray);">Le ${date}</span>
                        </li>
                    `;
                });
            } else {
                listComments.innerHTML = `<li style="color: var(--text-gray);">Aucun commentaire écrit.</li>`;
            }
        } catch (error) {
            listPosts.innerHTML = `<li style="color: #991b1b;">Impossible de charger.</li>`;
            listComments.innerHTML = `<li style="color: #991b1b;">Impossible de charger.</li>`;
        }
    }

    return me;
}

function initProfile() {
    const saveButton = $("#profile-save");
    if (!saveButton) return;

    loadProfile().catch(() => {
        window.location.href = "/login";
    });

    saveButton.addEventListener("click", async () => {
        const message = $("#profile-message");
        const payload = {};

        const pseudo = $("#profile-pseudo")?.value?.trim();
        if (pseudo) payload.pseudo = pseudo;

        const email = $("#profile-email")?.value?.trim();
        if (email) payload.email = email;

        const password = $("#profile-password")?.value;
        if (password) {
            payload.currentPassword = $("#profile-current-password")?.value || "";
            payload.password        = password;
            payload.confirmPassword = $("#profile-confirm-password")?.value || "";
        }

        try {
            if (Object.keys(payload).length > 0) {
                await api("/api/me", {
                    method: "PUT",
                    body: JSON.stringify(payload),
                });
            }

            const file = $("#profile-avatar-file")?.files?.[0];
            if (file) {
                if (file.size > 2 * 1024 * 1024) {
                    throw new Error("La photo ne doit pas dépasser 2 Mo.");
                }

                const formData = new FormData();
                formData.append("avatar", file);

                const avatarData = await uploadFile("/api/me/avatar", formData);

                const avatarImg = $("#profile-avatar");
                if (avatarImg && avatarData.avatar_url) {
                    avatarImg.src    = `${API_BASE}${avatarData.avatar_url}`;
                    avatarImg.hidden = false;
                }
            }

            showMessage(message, "Profil mis à jour.");
            await loadProfile();

            const pwdFields = ["profile-current-password", "profile-password", "profile-confirm-password"];
            pwdFields.forEach(id => {
                const el = $(`#${id}`);
                if (el) el.value = "";
            });
            const fileInput = $("#profile-avatar-file");
            if (fileInput) fileInput.value = "";

        } catch (error) {
            showMessage(message, error.message, true);
        }
    });

    $("#profile-logout")?.addEventListener("click", async () => {
        await api("/api/auth/logout", { method: "POST" }).catch(() => {});
        window.location.href = "/login";
    });

    $("#profile-delete")?.addEventListener("click", async () => {
        if (!confirm("Supprimer définitivement ton compte ? Cette action est irréversible.")) return;
        const message = $("#profile-message");
        try {
            await api("/api/me", { method: "DELETE" });
            window.location.href = "/";
        } catch (error) {
            showMessage(message, error.message, true);
        }
    });
}

function initForgotPassword() {
    const button = $("#forgot-submit");
    if (!button) return;

    button.addEventListener("click", async () => {
        const message = $("#forgot-message");
        try {
            await api("/api/auth/forgot-password", {
                method: "POST",
                body: JSON.stringify({
                    email: $("#forgot-email")?.value?.trim() || "",
                }),
            });
            showMessage(message, "Si cet email existe, un lien de réinitialisation a été envoyé. Vérifie tes spams.");
        } catch (error) {
            showMessage(message, error.message, true);
        }
    });
}

function initResetPassword() {
    const button = $("#reset-submit");
    if (!button) return;

    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    const message = $("#reset-message");

    if (!token) {
        showMessage(message, "Le lien de réinitialisation est invalide ou introuvable.", true);
        button.disabled = true;
        button.style.opacity = "0.5";
        button.style.cursor = "not-allowed";
        return;
    }

    button.addEventListener("click", async () => {
        try {
            await api("/api/auth/reset-password", {
                method: "POST",
                body: JSON.stringify({
                    token: token,
                    password: $("#reset-password-input")?.value || "",
                    confirmPassword: $("#reset-confirm")?.value || "",
                }),
            });
            showMessage(message, "Mot de passe mis à jour. Tu vas être redirigé vers la connexion.");
            setTimeout(() => { window.location.href = "/login"; }, 2500);
        } catch (error) {
            showMessage(message, error.message, true);
        }
    });
}

initHome();
initLogin();
initRegister();
initProfile();
initForgotPassword();
initResetPassword();