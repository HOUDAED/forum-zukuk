const API_BASE = 'http://localhost:3000';

const signupForm = document.getElementById('signup-form');
const errorMsg = document.getElementById('error-message');
const successMsg = document.getElementById('success-message');
const passwordError = document.getElementById('password-error');
const tabLogin = document.getElementById('tab-login');
const tabSignup = document.getElementById('tab-signup');

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

function setButtonLoading(button, isLoading) {
    if (isLoading) {
        button.disabled = true;
        button.innerHTML = '<span class="spinner"></span>Création en cours...';
    } else {
        button.disabled = false;
        button.innerHTML = 'Créer mon compte';
    }
}

function validatePassword(password, confirmPassword) {
    if (password !== confirmPassword) {
        passwordError.style.display = 'block';
        return false;
    }
    passwordError.style.display = 'none';
    return true;
}

function validateEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validatePseudo(pseudo) {
    return pseudo.length >= 3 && pseudo.length <= 20;
}

function validatePassword2(password) {
    return password.length >= 8;
}

document.getElementById('confirm').addEventListener('change', () => {
    const password = document.getElementById('password').value;
    const confirm = document.getElementById('confirm').value;
    if (confirm) {
        validatePassword(password, confirm);
    }
});

document.getElementById('pseudo').addEventListener('input', (e) => {
    if (e.target.value.length < 3 || e.target.value.length > 20) {
        e.target.style.borderColor = '#f56565';
    } else {
        e.target.style.borderColor = '#e2e8f0';
    }
});

document.getElementById('password').addEventListener('input', (e) => {
    if (e.target.value.length < 8) {
        e.target.style.borderColor = '#f56565';
    } else {
        e.target.style.borderColor = '#e2e8f0';
    }
});

signupForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const pseudo = document.getElementById('pseudo').value.trim();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm').value;

    // Validations
    if (!pseudo || !email || !password || !confirmPassword) {
        showError('Tous les champs sont requis.');
        return;
    }

    if (!validatePseudo(pseudo)) {
        showError('Le pseudo doit contenir entre 3 et 20 caractères.');
        return;
    }

    if (!validateEmail(email)) {
        showError('Email invalide.');
        return;
    }

    if (!validatePassword2(password)) {
        showError('Le mot de passe doit contenir au moins 8 caractères.');
        return;
    }

    if (!validatePassword(password, confirmPassword)) {
        showError('Les mots de passe ne correspondent pas.');
        return;
    }

    const btnRegister = document.getElementById('btn-register');
    setButtonLoading(btnRegister, true);
    errorMsg.classList.remove('show');

    try {
        const response = await fetch(`${API_BASE}/api/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ 
                pseudo, 
                email, 
                password,
                confirmPassword 
            })
        });

        const data = await response.json();

        if (!response.ok) {
            showError(data.error || 'Erreur lors de l\'inscription.');
            return;
        }

        // Succès
        showSuccess('Compte créé! Un code OTP a été envoyé par email. Redirection...');
        
        // Sauvegarder l'email et rediriger vers la page de login
        localStorage.setItem('pendingEmail', email);
        
        setTimeout(() => {
            window.location.href = '/auth.html'; // ou un formulaire OTP spécifique
        }, 2000);

    } catch (error) {
        console.error('Error:', error);
        showError('Erreur réseau. Vérifiez votre connexion.');
    } finally {
        setButtonLoading(btnRegister, false);
    }
});

// Navigation entre tabs
function showForm(formType) {
    if (formType === 'login') {
        window.location.href = '/auth.html';
    } else {
        // Rester sur signup
        tabLogin.classList.remove('active');
        tabSignup.classList.add('active');
    }
}

// Vérifier si déjà connecté
window.addEventListener('load', () => {
    const token = localStorage.getItem('authToken');
    if (token) {
        // Déjà connecté, rediriger
        window.location.href = '/dashboard';
    }
});
