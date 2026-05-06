// Board JavaScript - Frontend Logic
const API_BASE = 'http://localhost:8081/api'; // Correction du port à 8081

// DOM Elements
const moodGrid = document.getElementById('moodGrid');
const statsSection = document.getElementById('statsSection');
const discussionsList = document.getElementById('discussionsList');
const quoteBanner = document.getElementById('quoteBanner');
const quoteLoader = document.getElementById('quoteLoader');
const quoteText = document.getElementById('quoteText');
const navButtons = document.querySelectorAll('.nav-item');

let currentMood = 'Calme';
let currentNav = 'Accueil';

// Initialize the board
document.addEventListener('DOMContentLoaded', () => {
  updateGreeting();
  loadBoardData();
  setupEventListeners();
  fetchRandomQuote();
});

// Options partagées pour envoyer les cookies au backend
const fetchOptions = {
    credentials: 'include'
};

// Fetch all board data from backend
async function loadBoardData() {
  try {
    const response = await fetch(`${API_BASE}/board`, fetchOptions);
    if (!response.ok) {
        if (response.status === 401) window.location.href = "/login";
        throw new Error("Erreur serveur");
    }
    const data = await response.json();
    
    renderMoods(data.moods);
    renderStats(data.stats);
    renderDiscussions(data.discussions);
  } catch (error) {
    console.error('Error loading board data:', error);
    renderMoodsLocal();
    renderStatsLocal();
    renderDiscussionsLocal();
  }
}

// Render moods from data
function renderMoods(moods) {
  moodGrid.innerHTML = '';
  moods.forEach((mood) => {
    const button = document.createElement('button');
    button.className = `mood-button ${mood.name === currentMood ? 'active' : ''}`;
    button.innerHTML = `
      <span class="mood-emoji">${mood.emoji}</span>
      <span class="mood-name">${mood.name}</span>
    `;
    
    button.addEventListener('click', () => {
      selectMood(mood.name);
      updateMoodInBackend(mood);
    });
    
    moodGrid.appendChild(button);
  });
}

// Render moods locally if API fails
function renderMoodsLocal() {
  const moods = [
    { name: 'Bien', emoji: '😊' },
    { name: 'Calme', emoji: '😌' },
    { name: 'Triste', emoji: '😢' },
    { name: 'Anxieux', emoji: '😰' },
    { name: 'Colère', emoji: '😡' },
  ];
  renderMoods(moods);
}

// Select a mood
// Fonction pour changer l'humeur ET le fond d'écran
function selectMood(moodName) {
  currentMood = moodName;
  const buttons = document.querySelectorAll('.mood-button');
  
  // Palette de couleurs apaisantes selon l'humeur
  const moodColors = {
    'Bien': 'linear-gradient(135deg, #fdf4ff 0%, #fae8ff 50%, #fce7f3 100%)',      // Rose joyeux
    'Calme': 'linear-gradient(135deg, #f0fdf4 0%, #dcfce7 50%, #e0f2fe 100%)',     // Vert/Bleu zen
    'Triste': 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 50%, #e0e7ff 100%)',    // Bleu doux réconfortant
    'Anxieux': 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 50%, #e2e8f0 100%)',   // Gris/Bleu neutre
    'Colère': 'linear-gradient(135deg, #fff7ed 0%, #ffedd5 50%, #fef08a 100%)',    // Pêche/Soleil apaisant
    'default': 'linear-gradient(135deg, #eef2ff 0%, #fae8ff 50%, #ecfeff 100%)'
  };

  // Met à jour les boutons
  buttons.forEach((btn) => {
    btn.classList.remove('active');
    if (btn.textContent.includes(moodName)) {
      btn.classList.add('active');
    }
  });

  // Animation douce du fond d'écran
  const bg = document.querySelector('.background-gradient');
  bg.style.transition = 'background 1.5s ease-in-out'; // Transition super douce
  bg.style.background = moodColors[moodName] || moodColors['default'];
}

// Update mood in backend (Route protégée)
async function updateMoodInBackend(mood) {
  try {
    await fetch(`${API_BASE}/mood`, {
      method: 'POST',
      credentials: 'include', // Envoi de l'identité
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(mood),
    });
  } catch (error) {
    console.error('Error updating mood:', error);
  }
}

// Render stats
function renderStats(stats) {
  statsSection.innerHTML = '';
  stats.forEach((stat, index) => {
    const colorMap = {
      'Posts partagés': 'blue',
      'Réponses reçues': 'red',
      'Jours actifs': 'purple',
    };
    
    const colorClass = colorMap[stat.title] || 'blue';
    const statCard = document.createElement('div');
    statCard.className = 'stat-card';
    statCard.innerHTML = `
      <div class="stat-icon ${colorClass}"></div>
      <span class="stat-title">${stat.title}</span>
    `;
    
    statsSection.appendChild(statCard);
  });
}

// Render stats locally if API fails
function renderStatsLocal() {
  const stats = [
    { title: 'Posts partagés', color: 'blue' },
    { title: 'Réponses reçues', color: 'red' },
    { title: 'Jours actifs', color: 'purple' },
  ];
  renderStats(stats);
}

// Render discussions
function renderDiscussions(discussions) {
  discussionsList.innerHTML = '';
  discussions.forEach((post) => {
    const tagColorMap = {
      'Stress': 'tag-stress',
      'Bien-être': 'tag-wellbeing',
      'Solitude': 'tag-loneliness',
    };
    
    const tagClass = tagColorMap[post.tag] || 'tag-stress';
    const discussionCard = document.createElement('div');
    discussionCard.className = 'discussion-card';
    discussionCard.innerHTML = `
      <div class="discussion-header">
        <div class="discussion-avatar">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
            <circle cx="12" cy="7" r="4"></circle>
          </svg>
        </div>
        <div class="discussion-author-info">
          <span class="discussion-author">${post.author}</span>
          <span class="discussion-time">• ${post.time}</span>
        </div>
      </div>
      <h3 class="discussion-title">${post.title}</h3>
      <div class="discussion-footer">
        <span class="discussion-tag ${tagClass}">${post.tag}</span>
        <div class="discussion-actions">
          <button class="discussion-action like" data-id="${post.id}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
            </svg>
            <span>${post.likes}</span>
          </button>
          <button class="discussion-action comment" data-id="${post.id}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
            <span>${post.comments}</span>
          </button>
        </div>
      </div>
    `;
    
    // Add event listeners for like button
    const likeBtn = discussionCard.querySelector('.like');
    likeBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      likeDiscussion(post.id, likeBtn);
    });
    
    discussionsList.appendChild(discussionCard);
  });
}

// Render discussions locally if API fails
function renderDiscussionsLocal() {
  const discussions = [
    {
      id: 1,
      author: 'Marie_123',
      time: '2h',
      title: "J'ai du mal à gérer mon stress au travail",
      tag: 'Stress',
      likes: 12,
      comments: 8,
    },
    {
      id: 2,
      author: 'Thomas_zen',
      time: '5h',
      title: "Techniques de respiration qui m'ont aidé",
      tag: 'Bien-être',
      likes: 24,
      comments: 15,
    },
    {
      id: 3,
      author: 'Sophie_22',
      time: '1 jour',
      title: "Se sentir seul à l'université",
      tag: 'Solitude',
      likes: 18,
      comments: 22,
    },
  ];
  renderDiscussions(discussions);
}

// Like a discussion (Route protégée)
async function likeDiscussion(discussionId, button) {
  try {
    const response = await fetch(`${API_BASE}/discussion/${discussionId}/like`, {
      method: 'POST',
      credentials: 'include' // Envoi de l'identité
    });
    
    if (response.ok) {
      const span = button.querySelector('span');
      const currentLikes = parseInt(span.textContent);
      span.textContent = currentLikes + 1;
      button.style.color = '#ff5722';
    }
  } catch (error) {
    console.error('Error liking discussion:', error);
  }
}

// Fetch random quote from backend
async function fetchRandomQuote() {
  try {
    const response = await fetch(`${API_BASE}/quote`, fetchOptions);
    const data = await response.json();
    displayQuote(data.quote);
  } catch (error) {
    console.error('Error fetching quote:', error);
    getRandomQuoteLocal();
  }
}

// Display quote
function displayQuote(quote) {
  quoteLoader.style.display = 'none';
  quoteText.textContent = `"${quote}"`;
  quoteText.style.display = 'block';
}

// Get random quote locally if API fails
function getRandomQuoteLocal() {
  const quotes = [
    "Tu n'es pas seul. Chaque jour est une nouvelle opportunité.",
    "Prendre soin de soi n'est pas un luxe, c'est une nécessité.",
    "Chaque petit pas vers la guérison est une victoire.",
    "Il est tout à fait normal de ne pas se sentir bien tous les jours.",
    "Ta santé mentale est une priorité absolue.",
    "Respire. Tu fais de ton mieux et c'est largement suffisant.",
    "N'oublie pas d'être aussi indulgent avec toi-même qu'avec les autres.",
  ];
  
  const randomQuote = quotes[Math.floor(Math.random() * quotes.length)];
  displayQuote(randomQuote);
}

// Setup event listeners
function setupEventListeners() {
  // Navigation buttons
  navButtons.forEach((btn) => {
    btn.addEventListener('click', () => {
      navButtons.forEach((b) => b.classList.remove('active'));
      btn.classList.add('active');
      currentNav = btn.dataset.nav;

      // Redirection si le bouton contient un attribut data-href
      if (btn.dataset.href) {
        window.location.href = btn.dataset.href;
      }
    });
  });
  
  // Header buttons
  const newPostBtn = document.querySelector('.btn-primary');
  if (newPostBtn) {
    newPostBtn.addEventListener('click', () => {
      console.log('New post button clicked');
      // Add new post functionality here
    });
  }
  
  const notificationBtn = document.querySelector('.btn-secondary');
  if (notificationBtn) {
    notificationBtn.addEventListener('click', () => {
      console.log('Notification button clicked');
      // Add notification functionality here
    });
  }
}

// Observe discussion cards for animation
const observerOptions = {
  threshold: 0.1,
  rootMargin: '0px 0px -100px 0px',
};

const observer = new IntersectionObserver((entries) => {
  entries.forEach((entry) => {
    if (entry.isIntersecting) {
      entry.target.style.opacity = '1';
      entry.target.style.transform = 'translateY(0)';
    }
  });
}, observerOptions);

document.addEventListener('DOMContentLoaded', () => {
  const cards = document.querySelectorAll('.discussion-card');
  cards.forEach((card) => {
    card.style.opacity = '0';
    card.style.transform = 'translateY(10px)';
    card.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    observer.observe(card);
  });
});

// À ajouter à la fin de board.js
async function updateGreeting() {
  const titleElement = document.querySelector('.header-title');
  const waveEmoji = '<span class="wave-emoji">👋</span>';
  
  // Déterminer Bonjour ou Bonsoir
  const hour = new Date().getHours();
  const greeting = (hour >= 5 && hour < 18) ? "Bonjour" : "Bonsoir";

  try {
    // Tenter de récupérer le pseudo depuis ton API
    const response = await fetch(`${API_BASE}/me`, fetchOptions);
    if (response.ok) {
      const user = await response.json();
      titleElement.innerHTML = `${greeting}, ${user.pseudo} ${waveEmoji}`;
      return;
    }
  } catch (e) {
    // Si échec, on laisse un message générique
  }
  titleElement.innerHTML = `${greeting} ${waveEmoji}`;
}
