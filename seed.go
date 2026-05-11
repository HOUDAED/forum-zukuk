package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

// randomDate génère une date aléatoire entre une date de début et aujourd'hui
func randomDate(start time.Time) time.Time {
	delta := time.Since(start)
	randomDuration := time.Duration(rand.Int63n(int64(delta)))
	return start.Add(randomDuration)
}

func main() {
	// Initialiser le générateur aléatoire
	rand.Seed(time.Now().UnixNano())
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	db, err := sql.Open("sqlite3", "./database/zukuk.db")
	if err != nil {
		log.Fatal("Erreur BDD :", err)
	}
	defer db.Close()

	fmt.Println("🚀 Démarrage de la génération massive de données Zukuk (depuis 2024)...")

	// On utilise une transaction pour insérer des milliers de lignes très rapidement
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	// 1. CATÉGORIES (On s'assure qu'elles existent)
	categoriesPost := []string{"Stress", "Solitude", "Études", "Anxiété", "Dépression", "Travail", "Relations", "Santé mentale", "Bien-être", "Sport", "Autre"}
	for _, c := range categoriesPost {
		tx.Exec(`INSERT OR IGNORE INTO post_categories (name) VALUES (?)`, c)
	}
	categoriesAct := []string{"Activités sportives", "Relaxation & bien-être", "Groupes de parole", "Activités créatives", "Lieux sociaux", "Activités nature"}
	for _, c := range categoriesAct {
		tx.Exec(`INSERT OR IGNORE INTO activity_categories (name) VALUES (?)`, c)
	}

	// 2. UTILISATEURS (Génération de 50 utilisateurs)
	fmt.Println("👤 Génération de 50 utilisateurs...")
	commonPassword := hashPassword("zukuk123")
	pseudos := []string{"Edvige", "Maxime", "Agnia", "Caleb", "Tyrique", "Rahima", "Izzah", "DataNinja", "LyonDev", "CodeurFou", "TamTamMaster", "FootFan", "ZenStudent", "IA_Lover", "TechBro"}
	
	// On ajoute 35 utilisateurs génériques
	for i := 1; i <= 35; i++ {
		pseudos = append(pseudos, fmt.Sprintf("User%d", i))
	}

	for i, pseudo := range pseudos {
		date := randomDate(startDate)
		email := fmt.Sprintf("%s@zukuk.com", pseudo)
		bio := "Passionné(e) par l'informatique, les données et la vie lyonnaise."
		
		_, err := tx.Exec(`
			INSERT INTO users (pseudo, email, password_hash, bio, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			pseudo, email, commonPassword, bio, date, date,
		)
		if err != nil {
			log.Printf("Erreur user %s: %v", pseudo, err)
		}

		// Paramètres pour chaque utilisateur
		tx.Exec(`INSERT INTO user_settings (user_id, theme, updated_at) VALUES (?, ?, ?)`, i+1, "glass", date)
		
		// Historique de connexion (3 par utilisateur)
		for c := 0; c < 3; c++ {
			tx.Exec(`INSERT INTO connection_history (user_id, ip_address, user_agent, created_at) VALUES (?, '192.168.1.50', 'Mozilla/5.0 Windows', ?)`, i+1, randomDate(date))
		}
	}

	// 3. POSTS / DISCUSSIONS (Génération de 200 posts)
	fmt.Println("📝 Génération de 200 discussions...")
	postTitles := []string{
		"Des avis sur le Bachelor IA & Data ?", "Gros stress avant les partiels d'informatique", 
		"Cherche des gens pour un foot ce weekend", "Comment gérer l'anxiété liée aux deadlines ?",
		"Meilleurs spots pour bosser à Lyon ?", "Besoin de faire une pause des écrans (Digital Detox)",
		"Je me sens seul(e) depuis mon déménagement", "Quelqu'un joue du tam-tam ici ?",
		"Comment trouver une alternance sans expérience ?", "Syndrome de l'imposteur en code...",
	}

	for i := 1; i <= 200; i++ {
		userID := rand.Intn(50) + 1
		catID := rand.Intn(11) + 1
		title := postTitles[rand.Intn(len(postTitles))]
		content := fmt.Sprintf("Ceci est le contenu généré automatiquement pour la discussion %d. J'aimerais beaucoup avoir vos retours et échanger avec vous sur ce sujet !", i)
		date := randomDate(startDate)

		tx.Exec(`
			INSERT INTO posts (user_id, category_id, title, content, is_anonymous, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			userID, catID, title, content, rand.Intn(2), date, date,
		)
	}

	// 4. COMMENTAIRES (Génération de 600 commentaires)
	fmt.Println("💬 Génération de 600 commentaires...")
	commentsContent := []string{
		"Totalement d'accord avec toi !", "Je vis exactement la même chose en ce moment...",
		"Courage, c'est une mauvaise passe, ça va aller mieux.", "Si tu veux on peut aller boire un café pour en parler.",
		"As-tu essayé la méthode Pomodoro pour réviser ?", "Super initiative !", 
		"Merci pour le partage, ça aide de ne pas se sentir seul.", "Grave, le réseau de la ville est galère parfois.",
	}

	for i := 1; i <= 600; i++ {
		postID := rand.Intn(200) + 1
		userID := rand.Intn(50) + 1
		content := commentsContent[rand.Intn(len(commentsContent))]
		date := randomDate(startDate)

		tx.Exec(`
			INSERT INTO comments (post_id, user_id, content, is_anonymous, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			postID, userID, content, rand.Intn(2), date, date,
		)
		
		// Notifications pour ces commentaires (si pas son propre post)
		tx.Exec(`
			INSERT INTO notifications (user_id, actor_id, type, post_id, created_at)
			SELECT user_id, ?, 'comment', id, ? FROM posts WHERE id = ? AND user_id != ?`,
			userID, date, postID, userID,
		)
	}

	// 5. LIKES (Génération de 1000 likes sur les posts)
	fmt.Println("❤️ Génération de 1000 likes...")
	for i := 1; i <= 1000; i++ {
		postID := rand.Intn(200) + 1
		userID := rand.Intn(50) + 1
		date := randomDate(startDate)
		
		// INSERT OR IGNORE gère les doublons uniques (post_id, user_id)
		tx.Exec(`INSERT OR IGNORE INTO post_likes (post_id, user_id, created_at) VALUES (?, ?, ?)`, postID, userID, date)
	}

	// 6. HUMEURS (Génération de 800 entrées d'humeur)
	fmt.Println("🎭 Génération de 800 historiques d'humeur...")
	moods := []string{"Bien", "Calme", "Triste", "Anxieux", "Colère"}
	for i := 1; i <= 800; i++ {
		userID := rand.Intn(50) + 1
		mood := moods[rand.Intn(len(moods))]
		date := randomDate(startDate)
		// Eviter les doublons le même jour
		tx.Exec(`INSERT OR IGNORE INTO mood_history (user_id, mood, recorded_at) VALUES (?, ?, ?)`, userID, mood, date)
	}

	// 7. ACTIVITÉS (Génération de 15 activités locales)
	fmt.Println("📍 Génération de 15 activités...")
	activities := []struct{ name, address string; lat, lng float64 }{
		{"Match de foot détente", "Parc de la Tête d'Or, Lyon", 45.7789, 4.8528},
		{"Session révisions IA", "Campus YNOV, Lyon", 45.7461, 4.8384},
		{"Atelier Percussions & Tam-tam", "Place Bellecour, Lyon", 45.7578, 4.8320},
		{"Groupe de parole : Gérer le stress", "Bibliothèque Part-Dieu", 45.7621, 4.8569},
	}

	for i := 1; i <= 15; i++ {
		act := activities[rand.Intn(len(activities))]
		creator := rand.Intn(10) + 1
		cat := rand.Intn(6) + 1
		date := randomDate(startDate)
		
		res, _ := tx.Exec(`
			INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, max_places, created_at) 
			VALUES (?, ?, ?, 'Venez nombreux pour décompresser et faire des rencontres sympas !', ?, ?, ?, 15, ?)`,
			creator, cat, act.name, act.address, act.lat, act.lng, date,
		)
		
		// Inscription de 5 participants aléatoires par activité
		actID, _ := res.LastInsertId()
		for p := 1; p <= 5; p++ {
			participant := rand.Intn(50) + 1
			tx.Exec(`INSERT OR IGNORE INTO activity_participants (activity_id, user_id, joined_at) VALUES (?, ?, ?)`, actID, participant, date)
		}
	}

	// Validation de la transaction
	if err := tx.Commit(); err != nil {
		log.Fatal("Erreur lors du Commit :", err)
	}

	fmt.Println("✅ SUCCÈS TOTAL ! La base de données contient maintenant des milliers d'interactions.")
	fmt.Println("🔑 Tous les mots de passe sont : zukuk123")
	fmt.Println("📧 Exemples de comptes : Edvige@zukuk.com, Maxime@zukuk.com, Agnia@zukuk.com")
}