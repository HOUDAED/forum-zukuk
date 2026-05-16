document.addEventListener('DOMContentLoaded', () => {
    
   const API_BASE = '/api';
    let currentSettings = {};

    // ─── TOAST UNIVERSEL (Polymorphe) ───────────────────
    function showToast(msg, isError = false) {
        const existing = document.getElementById('settings-toast');
        if (existing) existing.remove();

        const toast = document.createElement('div');
        toast.id = 'settings-toast';
        toast.textContent = msg;
        
        toast.style.cssText = `position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%) translateY(100px); background: var(--bg-element); color: var(--text-primary); border: 1px solid var(--border-color); padding: 14px 28px; border-radius: 16px; font-weight: 600; font-size: 0.95rem; box-shadow: var(--shadow-md); z-index: 9999; transition: transform 0.4s cubic-bezier(0.2, 0.9, 0.2, 1);`;
        
        if (isError) {
             toast.style.color = 'var(--danger-color)';
             toast.style.borderColor = 'var(--danger-color)';
        } else {
             toast.style.color = 'var(--primary-color)';
             toast.style.borderColor = 'var(--primary-color)';
        }

        document.body.appendChild(toast);
        requestAnimationFrame(() => toast.style.transform = 'translateX(-50%) translateY(0)');
        setTimeout(() => {
            toast.style.transform = 'translateX(-50%) translateY(100px)';
            setTimeout(() => toast.remove(), 400);
        }, 3500);
    }

    // ─── MODALE DE CONFIRMATION (Polymorphe) ────────
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
            background: var(--bg-element); border-radius: 24px; padding: 32px; width: 100%; max-width: 400px;
            box-shadow: var(--shadow-md); border: 1px solid var(--border-color); transform: translateY(20px) scale(0.95);
            transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1); text-align: center;
        `;

        modal.innerHTML = `
            <div style="width: 64px; height: 64px; background: rgba(239, 68, 68, 0.1); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 20px;">
                <svg viewBox="0 0 24 24" fill="none" stroke="var(--danger-color)" stroke-width="2" style="width: 32px; height: 32px;">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
            </div>
            <h3 style="font-size: 1.25rem; font-weight: 700; color: var(--text-primary); margin-bottom: 12px; font-family: 'Merriweather', serif;">${title}</h3>
            <p style="color: var(--text-secondary); font-size: 0.95rem; margin-bottom: 28px; line-height: 1.5;">${message}</p>
            <div style="display: flex; gap: 12px;">
                <button id="cancel-confirm" style="flex: 1; padding: 12px; border: 2px solid var(--border-color); border-radius: 14px; background: transparent; color: var(--text-secondary); font-weight: 600; cursor: pointer; transition: all 0.2s;">Annuler</button>
                <button id="accept-confirm" style="flex: 1; padding: 12px; border: none; border-radius: 14px; background: var(--danger-color); color: white; font-weight: 600; cursor: pointer; box-shadow: 0 4px 14px rgba(239, 68, 68, 0.3); transition: all 0.2s;">Déconnecter</button>
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
        
        document.getElementById('accept-confirm').onclick = () => {
            close();
            onConfirm();
        };

        // Effets de survol
        document.getElementById('cancel-confirm').onmouseover = function() { this.style.background = 'var(--bg-element-hover)'; this.style.color = 'var(--text-primary)'; };
        document.getElementById('cancel-confirm').onmouseout = function() { this.style.background = 'transparent'; this.style.color = 'var(--text-secondary)'; };
        document.getElementById('accept-confirm').onmouseover = function() { this.style.filter = 'brightness(1.1)'; this.style.transform = 'scale(1.05)'; };
        document.getElementById('accept-confirm').onmouseout = function() { this.style.filter = 'none'; this.style.transform = 'none'; };
    }

    // ─── 1. GESTION DES ONGLETS ──────────────────────────────────────────
    const tabs = document.querySelectorAll('.tab-link[data-target]');
    const contents = document.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            tabs.forEach(t => t.classList.remove('active'));
            contents.forEach(c => c.classList.remove('active'));
            
            e.currentTarget.classList.add('active');
            const targetId = e.currentTarget.getAttribute('data-target');
            document.getElementById(targetId).classList.add('active');

            if(targetId === 'security') {
                loadSessions();
            }
        });
    });

    // ─── 2. CHARGEMENT DE L'INTERFACE ────────────────────────────────────
    async function loadSettings() {
        try {
            // 1. Charge les paramètres de thème (Clair/Sombre/Glass)
            const res = await fetch(`${API_BASE}/settings/`, { credentials: 'include' });
            if (res.ok) {
                currentSettings = await res.json();
                
                const anonCheck = document.getElementById('pref-anon');
                if(anonCheck) anonCheck.checked = currentSettings.anon_by_default;
                
                updateThemeButtons(currentSettings.theme);
                document.documentElement.setAttribute('data-theme', currentSettings.theme);
            }

            // 🔴 2. LA NOUVEAUTÉ : Demande l'humeur à la base de données !
            const moodRes = await fetch(`${API_BASE}/me/mood-history`, { credentials: 'include' });
            if (moodRes.ok) {
                const moodData = await moodRes.json();
                if (moodData.history && moodData.history.length > 0) {
                    const currentMood = moodData.history[0].mood.toLowerCase().replace('è', 'e');
                    document.documentElement.setAttribute('data-mood', currentMood);
                    localStorage.setItem('zukuk_mood', currentMood);
                }
            }

        } catch (e) {
            showToast("Erreur de chargement des paramètres.", true);
        }
    }

    // ─── 3. SAUVEGARDE DE L'INTERFACE ────────────────────────────────────
    async function updateSetting(newValues) {
        const payload = { ...currentSettings, ...newValues };
        try {
            const res = await fetch(`${API_BASE}/settings/`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
                credentials: 'include'
            });
            if(res.ok) {
                currentSettings = payload;
                showToast("✨ Paramètres sauvegardés !", false);
            } else {
                showToast("Erreur lors de la sauvegarde.", true);
            }
        } catch (e) { 
            showToast("Impossible de joindre le serveur.", true);
        }
    }

    function updateThemeButtons(activeTheme) {
        document.querySelectorAll('.theme-btn').forEach(btn => {
            if(btn.getAttribute('data-theme') === activeTheme) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        });
    }

    const anonToggle = document.getElementById('pref-anon');
    if(anonToggle) {
        anonToggle.addEventListener('change', (e) => {
            updateSetting({ anon_by_default: e.target.checked });
        });
    }

    document.querySelectorAll('.theme-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const selectedTheme = e.currentTarget.getAttribute('data-theme');
            
            updateThemeButtons(selectedTheme);
            updateSetting({ theme: selectedTheme });
            
            // Applique sur document.documentElement
            document.documentElement.setAttribute('data-theme', selectedTheme);
            localStorage.setItem('zukuk_theme', selectedTheme);
        });
    });

    // ─── 4. SÉCURITÉ : GESTION DES SESSIONS ──────────────────────────────
    async function loadSessions() {
        const container = document.getElementById('sessions-list');
        try {
            const res = await fetch(`${API_BASE}/settings/sessions`, { credentials: 'include' });
            if (res.ok) {
                const data = await res.json();
                if(data.sessions && data.sessions.length > 0) {
                    container.innerHTML = data.sessions.map(s => `
                        <div class="session-item">
                            <div>
                                <strong>ID: ${s.token.substring(0,8)}...</strong> <br>
                                <span style="font-size:0.85rem;">Connecté le : ${new Date(s.created_at).toLocaleDateString('fr-FR', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' })}</span>
                            </div>
                            ${s.is_current ? '<span style="color: white; font-weight: bold; font-size: 0.85rem; background: var(--success-color); padding: 4px 8px; border-radius: 6px;">Appareil actuel</span>' : ''}
                        </div>
                    `).join('');
                } else {
                    container.innerHTML = '<div style="color: var(--text-secondary);">Aucune session active trouvée.</div>';
                }
            }
        } catch(e) { 
            container.innerHTML = '<div style="color: var(--danger-color);">Erreur de chargement des sessions.</div>'; 
        }
    }

    const btnRevoke = document.getElementById('btn-revoke-sessions');
    if(btnRevoke) {
        btnRevoke.addEventListener('click', () => {
            showConfirmModal(
                "Sécurité des sessions",
                "Es-tu sûr de vouloir forcer la déconnexion de tous tes autres appareils (téléphones, ordinateurs) ? Tu devras te reconnecter.",
                async () => {
                    const originalText = btnRevoke.textContent;
                    btnRevoke.textContent = "Déconnexion en cours...";
                    btnRevoke.disabled = true;

                    try {
                        const res = await fetch(`${API_BASE}/settings/sessions/revoke`, { method: 'POST', credentials: 'include' });
                        if (res.ok) {
                            loadSessions(); 
                            showToast("🛡️ Tous les autres appareils ont été déconnectés !", false);
                        } else {
                            showToast("Erreur lors de la déconnexion.", true);
                        }
                    } catch(e) {
                        showToast("Serveur injoignable.", true);
                    } finally {
                        btnRevoke.textContent = originalText;
                        btnRevoke.disabled = false;
                    }
                }
            );
        });
    }

    // ─── 5. DONNÉES (EXPORT) ─────────────────────────────────────────────
    const btnExport = document.getElementById('btn-export');
    if(btnExport) {
        btnExport.addEventListener('click', () => {
            window.location.href = `${API_BASE}/settings/export`;
        });
    }

    // Lancement de l'initialisation
    loadSettings();
});