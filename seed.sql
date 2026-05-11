-- ==============================================================================
-- 🚀 ZUKUK - FICHIER DE SEEDING MASSIF (100 Users, 2000+ Posts, 5000+ Commentaires)
-- Mot de passe universel : Zukuk123!
-- ==============================================================================

PRAGMA foreign_keys = OFF;

-- 🧹 1. NETTOYAGE COMPLET DE LA BASE
DELETE FROM activity_participants;
DELETE FROM activities;
DELETE FROM mood_history;
DELETE FROM comment_likes;
DELETE FROM post_likes;
DELETE FROM comments;
DELETE FROM posts;
DELETE FROM users;
DELETE FROM sqlite_sequence;

-- 👤 2. CRÉATION DES 10 UTILISATEURS "RÉALISTES" (Admin & Piliers)
INSERT INTO users (pseudo, email, password_hash, bio, avatar_url, is_admin, created_at) VALUES 
('Edvige_Admin', 'edvige@zukuk.fr', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Créatrice de Zukuk. Bienvenue dans notre espace safe ! 💙', 'https://api.dicebear.com/7.x/notionists/svg?seed=Edvige&backgroundColor=eef2ff', 1, datetime('now', '-365 days')),
('Lumina', 'lumina@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Passionnée de psycho, toujours là pour écouter.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Lumina&backgroundColor=dbeafe', 0, datetime('now', '-300 days')),
('Alex_Lyon', 'alex@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Étudiant en data. Cherche un équilibre vie pro/perso.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Alex&backgroundColor=dcfce7', 0, datetime('now', '-250 days')),
('SarahZen', 'sarah@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Le yoga me sauve la vie ! 🧘‍♀️', 'https://api.dicebear.com/7.x/notionists/svg?seed=SarahZ&backgroundColor=fce7f3', 0, datetime('now', '-200 days')),
('Marcus_99', 'marcus@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Nouveau à Lyon. Sport en plein air.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Marcus&backgroundColor=fef3c7', 0, datetime('now', '-150 days')),
('ElodieT', 'elodie@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Introvertie qui sort de sa zone de confort.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Elodie&backgroundColor=e0f2fe', 0, datetime('now', '-100 days')),
('Tom_Dev', 'tom@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Trop d''écrans, pas assez de sommeil.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Tom&backgroundColor=fef9c3', 0, datetime('now', '-90 days')),
('Chloe_M', 'chloe@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Je vis au jour le jour.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Chloe&backgroundColor=fce7f3', 0, datetime('now', '-80 days')),
('Nico_B', 'nico@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Photographe amateur.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Nico&backgroundColor=f3e8ff', 0, datetime('now', '-60 days')),
('Amina', 'amina@email.com', '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u', 'Cherche toujours le positif.', 'https://api.dicebear.com/7.x/notionists/svg?seed=Amina&backgroundColor=dbeafe', 0, datetime('now', '-30 days'));

-- 👤 3. CRÉATION DE 90 UTILISATEURS GÉNÉRÉS ALÉATOIREMENT (Pour arriver à 100)
INSERT INTO users (pseudo, email, password_hash, bio, avatar_url, created_at)
WITH RECURSIVE cnt(x) AS (SELECT 11 UNION ALL SELECT x+1 FROM cnt WHERE x<100)
SELECT
  'ZukukUser_' || x,
  'user' || x || '@random.com',
  '$2a$12$NqL.O91b.8.rR2D/H1i8U.Z9g6P7Q9uQ4K9.z9v8u9.z9v8u9.z9v8u',
  'Membre de Zukuk depuis la vague ' || (ABS(RANDOM()) % 5 + 1) || '.',
  'https://api.dicebear.com/7.x/notionists/svg?seed=ZukukUser_' || x || '&backgroundColor=f1f5f9',
  datetime('now', '-' || (ABS(RANDOM()) % 300) || ' days', '-' || (ABS(RANDOM()) % 24) || ' hours')
FROM cnt;

-- 📝 4. CRÉATION DE 2000 POSTS ALÉATOIRES SUR 1 AN
-- Catégories BDD: 1=Stress, 2=Solitude, 3=Études, 4=Anxiété, 5=Dépression, 6=Travail, 7=Relations, 8=Santé, 9=Bien-être, 10=Sport, 11=Autre
INSERT INTO posts (user_id, category_id, title, content, created_at, updated_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<2000)
SELECT
  (ABS(RANDOM()) % 100) + 1, -- ID User entre 1 et 100
  (ABS(RANDOM()) % 11) + 1,  -- ID Categorie entre 1 et 11
  CASE (ABS(RANDOM()) % 10)
    WHEN 0 THEN 'Besoin de vider mon sac...'
    WHEN 1 THEN 'Je n''arrive plus à dormir à cause du stress'
    WHEN 2 THEN 'Petite victoire du jour 🎉'
    WHEN 3 THEN 'Quelqu''un se sent seul(e) en télétravail ?'
    WHEN 4 THEN 'Astuces pour calmer une crise d''angoisse'
    WHEN 5 THEN 'Les études me bouffent la vie'
    WHEN 6 THEN 'J''ai décidé de me reprendre en main !'
    WHEN 7 THEN 'Comment se faire des amis à Lyon ?'
    WHEN 8 THEN 'Je suis épuisé(e) mentalement aujourd''hui'
    ELSE 'Routine matinale : vos conseils ?'
  END,
  CASE (ABS(RANDOM()) % 4)
    WHEN 0 THEN 'Salut tout le monde. Je n''ai pas trop le moral aujourd''hui. J''ai l''impression que tout s''enchaîne mal et je perds le contrôle. Est-ce que certains ressentent ça parfois ? J''aimerais juste en parler.'
    WHEN 1 THEN 'Bonjour ! Je voulais partager une note positive. J''ai réussi à sortir me promener au parc pendant 1h sans regarder mon téléphone. Ça m''a fait un bien fou ! Prenez soin de vous.'
    WHEN 2 THEN 'Je suis étudiant et la pression des examens est ingérable. Je fais des insomnies et j''ai la boule au ventre en permanence. Si vous avez des applications ou des tisanes à conseiller, je prends.'
    ELSE 'Je suis nouveau sur le forum. Le concept de Zukuk est vraiment génial, merci pour cet espace safe. J''espère pouvoir échanger et faire de belles rencontres virtuelles (ou réelles) avec vous !'
  END || ' (Post #' || x || ')',
  datetime('now', '-' || (ABS(RANDOM()) % 365) || ' days', '-' || (ABS(RANDOM()) % 60) || ' minutes'),
  datetime('now', '-' || (ABS(RANDOM()) % 365) || ' days')
FROM cnt;

-- 💬 5. CRÉATION DE 5000 COMMENTAIRES ALÉATOIRES
INSERT INTO comments (post_id, user_id, content, created_at, updated_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<5000)
SELECT
  (ABS(RANDOM()) % 2000) + 1, -- Post ID au hasard
  (ABS(RANDOM()) % 100) + 1,  -- User ID au hasard
  CASE (ABS(RANDOM()) % 8)
    WHEN 0 THEN 'Je comprends totalement ce que tu ressens. Courage ! 💪'
    WHEN 1 THEN 'Essaie la cohérence cardiaque, ça m''aide énormément personnellement.'
    WHEN 2 THEN 'Bravo pour cette belle étape ! C''est super inspirant.'
    WHEN 3 THEN 'Tu n''es pas seul(e). Si tu veux discuter en privé, n''hésite pas.'
    WHEN 4 THEN 'Totalement d''accord avec toi. Le télétravail isole beaucoup trop.'
    WHEN 5 THEN 'As-tu pensé à en parler à un professionnel ? Ça m''a sauvé la vie.'
    WHEN 6 THEN 'Bienvenue sur Zukuk ! Tu vas voir, la commu est super bienveillante.'
    ELSE 'Merci pour le partage, je vais tester cette méthode dès ce soir.'
  END,
  datetime('now', '-' || (ABS(RANDOM()) % 300) || ' days'),
  datetime('now', '-' || (ABS(RANDOM()) % 300) || ' days')
FROM cnt;

-- ❤️ 6. CRÉATION DE 10 000 LIKES SUR LES POSTS
INSERT OR IGNORE INTO post_likes (post_id, user_id, created_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<10000)
SELECT
  (ABS(RANDOM()) % 2000) + 1,
  (ABS(RANDOM()) % 100) + 1,
  datetime('now', '-' || (ABS(RANDOM()) % 300) || ' days')
FROM cnt;

-- 📍 7. CRÉATION DE 40 ACTIVITÉS SUR LA CARTE DE LYON
-- (20 Passées, 20 Futures)
INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, max_places, created_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<40)
SELECT
  (ABS(RANDOM()) % 100) + 1,
  (ABS(RANDOM()) % 6) + 1,
  CASE (ABS(RANDOM()) % 6)
    WHEN 0 THEN 'Footing Détente Tête d''Or'
    WHEN 1 THEN 'Café Papote Bellecour'
    WHEN 2 THEN 'Session Yoga & Respiration'
    WHEN 3 THEN 'Groupe de parole : Gérer le stress'
    WHEN 4 THEN 'Balade photo au Parc de Parilly'
    ELSE 'Atelier Dessin en plein air'
  END,
  'Venez nombreux pour décompresser et faire des rencontres sympas, sans jugement.',
  CASE (ABS(RANDOM()) % 5)
    WHEN 0 THEN 'Parc de la Tête d''Or, Lyon'
    WHEN 1 THEN 'Place Bellecour, Lyon'
    WHEN 2 THEN 'Parc de Parilly, Bron'
    WHEN 3 THEN 'Quais du Rhône, Lyon'
    ELSE 'Parc de Gerland, Lyon'
  END,
  45.75 + (RANDOM() * 0.000000000000005), -- Légère variation de latitude autour de Lyon
  4.83 + (RANDOM() * 0.000000000000005),  -- Légère variation de longitude
  -- Moitié passé (historique), Moitié futur (sur la carte aujourd'hui)
  CASE WHEN x < 20 THEN strftime('%Y-%m-%dT%H:%M', datetime('now', '-' || (ABS(RANDOM()) % 30) || ' days'))
                   ELSE strftime('%Y-%m-%dT%H:%M', datetime('now', '+' || (ABS(RANDOM()) % 20) || ' days')) END,
  (ABS(RANDOM()) % 15) + 5, -- Places entre 5 et 20
  datetime('now', '-30 days')
FROM cnt;

-- 👥 8. CRÉATION DE 300 INSCRIPTIONS AUX ACTIVITÉS
INSERT OR IGNORE INTO activity_participants (activity_id, user_id, joined_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<300)
SELECT
  (ABS(RANDOM()) % 40) + 1,
  (ABS(RANDOM()) % 100) + 1,
  datetime('now', '-' || (ABS(RANDOM()) % 20) || ' days')
FROM cnt;

-- 🎭 9. CRÉATION D'HISTORIQUES D'HUMEUR
INSERT INTO mood_history (user_id, mood, recorded_at)
WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x<500)
SELECT
  (ABS(RANDOM()) % 100) + 1,
  CASE (ABS(RANDOM()) % 5)
    WHEN 0 THEN 'Bien'
    WHEN 1 THEN 'Calme'
    WHEN 2 THEN 'Triste'
    WHEN 3 THEN 'Anxieux'
    ELSE 'Colère'
  END,
  datetime('now', '-' || (ABS(RANDOM()) % 30) || ' days')
FROM cnt;

PRAGMA foreign_keys = ON;