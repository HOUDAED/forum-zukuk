const express = require('express');
const cors = require('cors');
const bcrypt = require('bcrypt');
const jwt = require('jsonwebtoken');
const nodemailer = require('nodemailer');
const rateLimit = require('express-rate-limit');
const dotenv = require('dotenv');
const path = require('path');

dotenv.config();

const app = express();

app.use(express.json());
app.use(cors());

app.use('/auth', express.static(path.join(__dirname, 'auth')));
app.use('/sign', express.static(path.join(__dirname, 'sign')));
app.use(express.static(path.join(__dirname, 'public')));

const loginLimiter = rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 5,
    message: 'Trop de tentatives de connexion, réessayez plus tard.'
});

const registerLimiter = rateLimit({
    windowMs: 60 * 60 * 1000,
    max: 3
});

const transporter = nodemailer.createTransport({
    service: process.env.EMAIL_SERVICE || 'gmail',
    auth: {
        user: process.env.EMAIL_USER,
        pass: process.env.EMAIL_PASSWORD
    }
});

const users = [];
const tempOTP = {};

function validateEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validatePassword(password) {
    return password.length >= 8;
}

app.post('/api/auth/login', loginLimiter, async (req, res) => {
    try {
        const { email, password } = req.body;

        if (!email || !password) {
            return res.status(400).json({ error: "Email et mot de passe requis." });
        }

        if (!validateEmail(email)) {
            return res.status(400).json({ error: "Email invalide." });
        }

        const user = users.find(u => u.email === email);

        if (!user || !(await bcrypt.compare(password, user.password))) {
            return res.status(401).json({ error: "Identifiants incorrects." });
        }

        const otp = Math.floor(100000 + Math.random() * 900000).toString();
        tempOTP[email] = { code: otp, expires: Date.now() + 300000 };

        await sendOTPEmail(email, otp);
        
        return res.status(200).json({ message: "OTP_SENT", email });
    } catch (error) {
        console.error('Login error:', error);
        res.status(500).json({ error: "Erreur serveur." });
    }
});

app.post('/api/auth/verify-otp', (req, res) => {
    try {
        const { email, code } = req.body;

        if (!email || !code) {
            return res.status(400).json({ error: "Email et code requis." });
        }

        const record = tempOTP[email];

        if (!record || record.code !== code || record.expires < Date.now()) {
            return res.status(400).json({ error: "Code invalide ou expiré." });
        }

        delete tempOTP[email];

        const user = users.find(u => u.email === email);
        const token = jwt.sign(
            { id: user.id, email: user.email, pseudo: user.pseudo },
            process.env.JWT_SECRET || 'your_secret_key',
            { expiresIn: '7d' }
        );

        return res.status(200).json({ message: "SUCCESS", token, user: { id: user.id, pseudo: user.pseudo, email: user.email } });
    } catch (error) {
        console.error('OTP verification error:', error);
        res.status(500).json({ error: "Erreur serveur." });
    }
});

app.post('/api/auth/register', registerLimiter, async (req, res) => {
    try {
        const { pseudo, email, password, confirmPassword } = req.body;

        if (!pseudo || !email || !password || !confirmPassword) {
            return res.status(400).json({ error: "Tous les champs sont requis." });
        }

        if (!validateEmail(email)) {
            return res.status(400).json({ error: "Email invalide." });
        }

        if (!validatePassword(password)) {
            return res.status(400).json({ error: "Le mot de passe doit contenir au moins 8 caractères." });
        }

        if (password !== confirmPassword) {
            return res.status(400).json({ error: "Les mots de passe ne correspondent pas." });
        }

        if (pseudo.length < 3 || pseudo.length > 20) {
            return res.status(400).json({ error: "Le pseudo doit contenir entre 3 et 20 caractères." });
        }

        const existingUser = users.find(u => u.email === email);
        if (existingUser) {
            return res.status(400).json({ error: "Cet email est déjà utilisé." });
        }

        const existingPseudo = users.find(u => u.pseudo === pseudo);
        if (existingPseudo) {
            return res.status(400).json({ error: "Ce pseudo est déjà pris." });
        }

        const hashedPassword = await bcrypt.hash(password, 10);

        const newUser = {
            id: Date.now(),
            pseudo,
            email,
            password: hashedPassword,
            isVerified: false,
            createdAt: new Date()
        };
        users.push(newUser);

        const otp = Math.floor(100000 + Math.random() * 900000).toString();
        tempOTP[email] = { code: otp, expires: Date.now() + 600000 };

        await sendWelcomeEmail(pseudo, email, otp);
        
        res.status(201).json({ message: "Utilisateur créé, OTP envoyé." });
    } catch (error) {
        console.error('Register error:', error);
        res.status(500).json({ error: "Erreur serveur." });
    }
});

async function sendOTPEmail(email, code) {
    try {
        await transporter.sendMail({
            from: process.env.EMAIL_USER,
            to: email,
            subject: 'Votre code de vérification',
            html: `
                <h2>Code de vérification</h2>
                <p>Votre code est: <strong>${code}</strong></p>
                <p>Ce code expire dans 5 minutes.</p>
            `
        });
        console.log(`[MAIL] Code OTP envoyé à ${email}`);
    } catch (error) {
        console.error('Error sending OTP email:', error);
    }
}

async function sendWelcomeEmail(pseudo, email, code) {
    try {
        await transporter.sendMail({
            from: process.env.EMAIL_USER,
            to: email,
            subject: 'Bienvenue sur Forum Zukuk',
            html: `
                <h2>Bienvenue ${pseudo} ! 👋</h2>
                <p>Votre code de vérification est: <strong>${code}</strong></p>
                <p>Ce code expire dans 10 minutes.</p>
                <p>Merci d'avoir rejoint notre communauté !</p>
            `
        });
        console.log(`[MAIL] Email de bienvenue envoyé à ${email}`);
    } catch (error) {
        console.error('Error sending welcome email:', error);
    }
}

app.get('/api/auth/google', (req, res) => {
    res.json({ 
        error: 'OAuth Google non configuré',
        message: 'Pour intégrer Google OAuth, installer: npm install passport passport-google-oauth20',
        docs: 'https://developers.google.com/identity/protocols/oauth2'
    });
});

app.post('/api/auth/google-callback', async (req, res) => {
    try {
        const { googleId, email, pseudo } = req.body;
        
        let user = users.find(u => u.googleId === googleId);
        
        if (!user) {
            user = {
                id: Date.now(),
                googleId,
                email,
                pseudo: pseudo || email.split('@')[0],
                password: null,
                isVerified: true,
                provider: 'google',
                createdAt: new Date()
            };
            users.push(user);
        }
        
        const token = jwt.sign(
            { id: user.id, email: user.email, pseudo: user.pseudo },
            process.env.JWT_SECRET || 'your_secret_key',
            { expiresIn: '7d' }
        );
        
        res.json({ message: 'SUCCESS', token, user });
    } catch (error) {
        res.status(500).json({ error: 'Erreur Google OAuth' });
    }
});

// Routes OAuth (Portail Écoles)
app.get('/api/auth/school', (req, res) => {
    res.json({ 
        error: 'Portail École non configuré',
        message: 'À intégrer avec votre système de portail scolaire',
    });
});

app.post('/api/auth/school-callback', async (req, res) => {
    try {
        const { schoolId, email, pseudo, schoolName } = req.body;
        
        let user = users.find(u => u.schoolId === schoolId);
        
        if (!user) {
            user = {
                id: Date.now(),
                schoolId,
                email,
                pseudo: pseudo || email.split('@')[0],
                password: null,
                isVerified: true,
                provider: 'school',
                schoolName,
                createdAt: new Date()
            };
            users.push(user);
        }
        
        const token = jwt.sign(
            { id: user.id, email: user.email, pseudo: user.pseudo },
            process.env.JWT_SECRET || 'your_secret_key',
            { expiresIn: '7d' }
        );
        
        res.json({ message: 'SUCCESS', token, user });
    } catch (error) {
        res.status(500).json({ error: 'Erreur Portail École' });
    }
});

app.get('/api/health', (req, res) => {
    res.json({ status: 'OK' });
});

app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'auth', 'auth.html'));
});

app.get('/auth.html', (req, res) => {
    res.redirect('/auth/auth.html');
});

app.get('/sign.html', (req, res) => {
    res.redirect('/sign/sign.html');
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Serveur lancé sur le port ${PORT}`);
    console.log(`📝 Accès: http://localhost:${PORT}/auth.html`);
    console.log(`📝 Accès: http://localhost:${PORT}/sign.html`);
});

module.exports = app;
