document.addEventListener('DOMContentLoaded', () => {
    
    const API_BASE = 'http://localhost:8081';
    let currentSettings = {};

    // ─── 1. GESTION DES ONGLETS ──────────────────────────────────────────
    // On ne sélectionne QUE les onglets qui ont un attribut data-target (ce qui exclut le bouton Profil)
    const tabs = document.querySelectorAll('.tab-link[data-target]');
    const contents = document.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            // Désactiver tous les onglets
            tabs.forEach(t => t.classList.remove('active'));
            contents.forEach(c => c.classList.remove('active'));
            
            // Activer l'onglet cliqué
            e.currentTarget.classList.add('active');
            const targetId = e.currentTarget.getAttribute('data-target');
            document.getElementById(targetId).classList.add('active');

            // Si on ouvre la sécurité, on charge les sessions
            if(targetId === 'security') {
                loadSessions();
            }
        });
    });

    // ─── 2. CHARGEMENT DE L'INTERFACE ────────────────────────────────────
    async function loadSettings() {
        try {
            const res = await fetch(`${API_BASE}/api/settings/`, { credentials: 'include' });
            if (res.ok) {
                currentSettings = await res.json();
                
                // Met à jour la case Anonyme
                const anonCheck = document.getElementById('pref-anon');
                if(anonCheck) anonCheck.checked = currentSettings.anon_by_default;
                
                // Met à jour les boutons de Thème
                updateThemeButtons(currentSettings.theme);
                
                // Optionnel : Applique visuellement le thème au body (si tu as du CSS pour ça)
                document.body.setAttribute('data-theme', currentSettings.theme);
            }
        } catch (e) {
            console.error("Erreur de chargement des paramètres :", e);
        }
    }

    // ─── 3. SAUVEGARDE DE L'INTERFACE ────────────────────────────────────
    async function updateSetting(newValues) {
        const payload = { ...currentSettings, ...newValues };
        try {
            const res = await fetch(`${API_BASE}/api/settings/`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
                credentials: 'include'
            });
            if(res.ok) {
                currentSettings = payload;
            }
        } catch (e) { 
            console.error(e); 
        }
    }

    // Gère le design des boutons de thème
    function updateThemeButtons(activeTheme) {
        document.querySelectorAll('.theme-btn').forEach(btn => {
            if(btn.getAttribute('data-theme') === activeTheme) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        });
    }

    // Écouteur pour la case anonymat
    const anonToggle = document.getElementById('pref-anon');
    if(anonToggle) {
        anonToggle.addEventListener('change', (e) => {
            updateSetting({ anon_by_default: e.target.checked });
        });
    }

// Écouteur pour les boutons de thème (dans settings.js)
    document.querySelectorAll('.theme-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const selectedTheme = e.currentTarget.getAttribute('data-theme');
            
            updateThemeButtons(selectedTheme);
            updateSetting({ theme: selectedTheme });
            
            // Applique instantanément le thème sur la page actuelle
            document.body.setAttribute('data-theme', selectedTheme);
            
            // 🔥 NOUVEAU : On sauvegarde dans le navigateur pour les autres pages
            localStorage.setItem('zukuk_theme', selectedTheme);
        });
    });
    // ─── 4. SÉCURITÉ : GESTION DES SESSIONS ──────────────────────────────
    async function loadSessions() {
        const container = document.getElementById('sessions-list');
        try {
            const res = await fetch(`${API_BASE}/api/settings/sessions`, { credentials: 'include' });
            if (res.ok) {
                const data = await res.json();
                if(data.sessions && data.sessions.length > 0) {
                    container.innerHTML = data.sessions.map(s => `
                        <div class="session-item" style="padding: 12px 0; border-bottom: 1px solid #e2e8f0; display: flex; justify-content: space-between; align-items: center;">
                            <div>
                                <strong style="color: #1e293b; font-family: monospace;">ID: ${s.token}</strong> <br>
                                <span style="font-size:0.85rem; color:#64748b">Connecté le : ${new Date(s.created_at).toLocaleDateString('fr-FR')}</span>
                            </div>
                            ${s.is_current ? '<span style="color: #10b981; font-weight: bold; font-size: 0.85rem; background: #d1fae5; padding: 4px 8px; border-radius: 6px;">Appareil actuel</span>' : ''}
                        </div>
                    `).join('');
                } else {
                    container.innerHTML = '<div style="color: #64748b;">Aucune session active trouvée.</div>';
                }
            }
        } catch(e) { 
            container.innerHTML = '<div style="color: #ef4444;">Erreur de chargement des sessions.</div>'; 
        }
    }

    const btnRevoke = document.getElementById('btn-revoke-sessions');
    if(btnRevoke) {
        btnRevoke.addEventListener('click', async () => {
            if(confirm("Veux-tu vraiment déconnecter tous tes autres appareils ?")) {
                try {
                    await fetch(`${API_BASE}/api/settings/sessions/revoke`, { method: 'POST', credentials: 'include' });
                    loadSessions(); // Recharge la liste pour montrer qu'il n'en reste qu'une
                    alert("Tous les autres appareils ont été déconnectés !");
                } catch(e) {
                    alert("Erreur lors de la déconnexion.");
                }
            }
        });
    }

    // ─── 5. DONNÉES (EXPORT) ─────────────────────────────────────────────
    const btnExport = document.getElementById('btn-export');
    if(btnExport) {
        btnExport.addEventListener('click', () => {
            window.location.href = `${API_BASE}/api/settings/export`;
        });
    }

    // Lancement de l'initialisation
    loadSettings();
});