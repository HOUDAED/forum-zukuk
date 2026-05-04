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
    const response = await fetch(path, {
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
    const response = await fetch(path, {
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
            window.location.href = "/profile";
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

    const subtitle = $("#profile-subtitle");
    if (subtitle) subtitle.textContent = `Connecté en tant que ${me.pseudo}`;

    const avatar = $("#profile-avatar");
    if (avatar) {
        if (me.avatar_url) {
            avatar.src    = me.avatar_url;
            avatar.hidden = false;
        } else {
            avatar.hidden = true;
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
                    avatarImg.src    = avatarData.avatar_url;
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



initHome();
initLogin();
initRegister();
initProfile();