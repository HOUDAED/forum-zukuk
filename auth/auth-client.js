const API_BASE = 'http://localhost:3000';

const loginForm = document.getElementById('login-form');
const otpForm = document.getElementById('otp-form');
const errorMsg = document.getElementById('error-message');
const successMsg = document.getElementById('success-message');
const otpEmailDisplay = document.getElementById('otp-email');
const retryBtn = document.getElementById('btn-retry');

let currentEmail = '';

function showError(message) {
    errorMsg.textContent = message;
    errorMsg.classList.add('show');
    successMsg.classList.remove('show');
    setTimeout(() => errorMsg.classList.remove('show'), 5000);
}

function showSuccess(message) {
    successMsg.textContent = message;
    successMsg.classList.add('show');
    errorMsg.classList.remove('show');
}

function showForm(formType) {
    if (formType === 'otp') {
        loginForm.style.display = 'none';
        otpForm.style.display = 'block';
    } else {
        loginForm.style.display = 'block';
        otpForm.style.display = 'none';
    }
}

function setButtonLoading(button, isLoading) {
    if (isLoading) {
        button.disabled = true;
        button.innerHTML = '<span class="spinner"></span>Chargement...';
    } else {
        button.disabled = false;
        button.innerHTML = button.id === 'btn-login' ? 'Continuer' : 'Vérifier';
    }
}

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;

    if (!email || !password) {
        showError('Email et mot de passe requis.');
        return;
    }

    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
        showError('Email invalide.');
        return;
    }

    const btnLogin = document.getElementById('btn-login');
    setButtonLoading(btnLogin, true);
    errorMsg.classList.remove('show');

    try {
        const response = await fetch(`${API_BASE}/api/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (!response.ok) {
            showError(data.error || 'Erreur lors de la connexion.');
            return;
        }

        currentEmail = email;
        otpEmailDisplay.textContent = email;
        showSuccess('Code OTP envoyé par email.');
        showForm('otp');
        document.getElementById('otp-code').focus();

    } catch (error) {
        console.error('Error:', error);
        showError('Erreur réseau. Vérifiez votre connexion.');
    } finally {
        setButtonLoading(btnLogin, false);
    }
});

otpForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const code = document.getElementById('otp-code').value.trim();

    if (!code || code.length !== 6) {
        showError('Veuillez entrer un code à 6 chiffres.');
        return;
    }

    if (!/^\d{6}$/.test(code)) {
        showError('Le code doit contenir uniquement des chiffres.');
        return;
    }

    const btnVerify = document.getElementById('btn-verify');
    setButtonLoading(btnVerify, true);
    errorMsg.classList.remove('show');

    try {
        const response = await fetch(`${API_BASE}/api/auth/verify-otp`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email: currentEmail, code })
        });

        const data = await response.json();

        if (!response.ok) {
            showError(data.error || 'Vérification échouée.');
            return;
        }

        localStorage.setItem('authToken', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
        
        showSuccess('Connexion réussie! Redirection...');
        setTimeout(() => {
            window.location.href = '/dashboard'; // À adapter selon ton app
        }, 1500);

    } catch (error) {
        console.error('Error:', error);
        showError('Erreur réseau. Réessayez.');
    } finally {
        setButtonLoading(btnVerify, false);
    }
});

retryBtn.addEventListener('click', () => {
    loginForm.reset();
    showForm('login');
    document.getElementById('email').focus();
});

function loginWithGoogle() {
    // Pour une intégration complète:
    // 1. Installer: npm install passport passport-google-oauth20
    // 2. Configurer credentials Google OAuth dans .env
    // 3. Utiliser: window.location.href = `${API_BASE}/api/auth/google`;
    
    alert('Google OAuth non configuré.\n\nPour activer:\n1. Générer une clé Google OAuth\n2. Installer passport\n3. Configurer .env\n\nDocumentation: https://developers.google.com/identity/protocols/oauth2');
}

function loginWithSchool() {
    // Pour intégrer avec ton portail scolaire:
    // Adapter selon le système utilisé (EasySchool, Pronote, etc.)
    
    alert('Portail École non configuré.\n\nContact: À implémenter avec votre portail scolaire');
}

// Vérifier si déjà connecté
window.addEventListener('load', () => {
    const token = localStorage.getItem('authToken');
    if (token) {
        // Déjà connecté, rediriger
        window.location.href = '/dashboard';
    }
});
