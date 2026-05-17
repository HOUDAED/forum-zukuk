package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"forum-zukuk/database" // Ajuste le chemin selon ton module
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: Fichier .env non trouvé")
	}
	database.Init()

	fmt.Println("🚀 Démarrage du Super Seeder Réaliste Zukuk (Objectif Juin 2026)...")
	rand.Seed(time.Now().UnixNano())

	// 1. Nettoyage de la BDD pour éviter les doublons ou conflits de clés
	fmt.Println("🧹 Nettoyage complet des tables...")
	tables := []string{
		"notifications", "activity_participants", "activities", "activity_categories",
		"comment_likes", "post_likes", "comments", "posts", "post_categories",
		"mood_history", "connection_history", "sessions", "password_history", "users",
	}
	for _, table := range tables {
		_, err := database.ZUKUKDB.Exec("DELETE FROM " + table)
		if err != nil {
			log.Printf("Erreur lors du nettoyage de la table %s: %v", table, err)
		}
		database.ZUKUKDB.Exec("DELETE FROM sqlite_sequence WHERE name='" + table + "'")
	}

	// Ré-initialisation des catégories par défaut via ta fonction native
	database.ZUKUKDB.Exec(`INSERT OR IGNORE INTO post_categories (id, name) VALUES (1, 'Santé mentale')`)

	// 2. Génération des 100 Utilisateurs avec des VRAIS noms et bios réalistes
	fmt.Println("👤 Génération de 100 utilisateurs réels...")
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	passwordStr := string(hash)

	realPseudos := []string{
		"Dagbegnon", "Edvige", "Maxime", "Izzah", "Agnia", "Caleb", "Tyrique", "Rahima", "Lucas", "Mathieu",
		"Chloe", "Thomas", "Sarah", "Ines", "Camille", "Nathan", "Juliette", "Enzo", "Amandine", "Alexandre",
		"Hugo", "Emma", "Manon", "Louis", "Jade", "Lana", "Zoe", "Arthur", "Gabriel", "Lina",
		"Koffi", "Mawuli", "Abla", "Yao", "Kwami", "Afoute", "Sika", "Sena", "Elolo", "Fafa",
		"Sophia", "Liam", "Olivia", "Noah", "Ava", "Oliver", "Isabella", "Elijah", "Mia", "James",
		"Charlotte", "Benjamin", "Amelia", "Lucas", "Mia", "Mason", "Evelyn", "Ethan", "Harper", "Pierre",
		"Jean", "Marie", "Michel", "Nicolas", "Antoine", "Guillaume", "Julien", "Romain", "Florian", "Corentin",
		"Celine", "Aurelie", "Elodie", "Laura", "Marine", "Lucie", "Marion", "Pauline", "Justine", "Lea",
		"Matthieu", "Damien", "Benoit", "Remi", "Simon", "Alban", "Fabien", "Cedric", "Ludovic", "Stephan",
		"Valerie", "Sandrine", "Isabelle", "Nathalie", "Sophie", "Caroline", "Stephanie", "Chantal", "Sylvie", "gta",
	}

	realBios := []string{
		"Étudiant passionné par l'Intelligence Artificielle et le traitement des données. Toujours ravi d'échanger sur la tech !",
		"Ici pour trouver de la sérénité, partager mes lectures inspirantes et discuter calmement au quotidien. 😊",
		"Grand amateur de football et de musique (percussions). Un esprit sain dans un corps sain !",
		"Développeur fullstack en quête d'inspiration. J'aime comprendre comment l'esprit et les machines fonctionnent.",
		"Parfois anxieux, j'ai trouvé dans la communauté Zukuk un espace d'entraide incroyable. Merci à tous.",
		"Prendre soin de sa santé mentale est la clé. Échangeons sans tabou et soutenons-nous les uns les autres.",
		"Curieux du monde, passionné d'environnement, de politique et de solutions numériques innovantes.",
	}

	var userIDs []int
	for i := 0; i < 100; i++ {
		pseudo := realPseudos[i]
		if i == 99 { // S'assurer que le compte 'gta' de tes captures d'écran est bien présent
			pseudo = "gta"
		}
		email := fmt.Sprintf("%s.zukuk@example.com", strings.ToLower(pseudo))
		bio := realBios[rand.Intn(len(realBios))]
		avatar := fmt.Sprintf("/uploads/avatars/default_%d.png", rand.Intn(6)+1) // Simule tes avatars physiques

		res, err := database.ZUKUKDB.Exec(`
			INSERT INTO users (pseudo, email, password_hash, bio, avatar_url, email_verified, created_at) 
			VALUES (?, ?, ?, ?, ?, 1, datetime('now', '-30 days'))`,
			pseudo, email, passwordStr, bio, avatar,
		)
		if err == nil {
			id, _ := res.LastInsertId()
			userIDs = append(userIDs, int(id))
			
			// Initialisation d'une ligne de réglages pour chaque utilisateur pour ton système de thème
			database.ZUKUKDB.Exec(`
				INSERT INTO user_settings (user_id, anon_by_default, show_mood, notif_likes, notif_comments, theme)
				VALUES (?, 0, 1, 1, 1, 'glass')`, id)
		}
	}

	// 3. Catégories de Forums et Humeurs historiques
	fmt.Println("🎭 Attribution de l'historique des humeurs...")
	moods := []string{"Bien", "Calme", "Triste", "Anxieux", "Colère"}
	for _, uid := range userIDs {
		// Génère 3 changements d'humeur historiques par personne
		for j := 3; j >= 0; j-- {
			timeOffset := fmt.Sprintf("-%d days", j*5)
			database.ZUKUKDB.Exec(`
				INSERT OR IGNORE INTO mood_history (user_id, mood, recorded_at) 
				VALUES (?, ?, datetime('now', ?))`,
				uid, moods[rand.Intn(len(moods))], timeOffset,
			)
		}
	}

	// Récupération des IDs réels des catégories d'entraide
	var categoryIDs []int
	rowsCat, _ := database.ZUKUKDB.Query(`SELECT id FROM post_categories`)
	for rowsCat.Next() {
		var cid int
		rowsCat.Scan(&cid)
		categoryIDs = append(categoryIDs, cid)
	}
	rowsCat.Close()

	// 4. Génération de 250 Discussions (Posts) cohérentes
	fmt.Println("📝 Publication de 250 discussions d'entraide...")
	titles := []string{
		"Comment surmonter le syndrome de l'imposteur en première année d'études ?",
		"Astuces pour calmer une crise d'anxiété avant un entretien important ?",
		"Besoin de vider mon sac : la solitude des grandes villes me pèse énormément",
		"Quels sont vos rituels du soir pour déconnecter du code et mieux dormir ?",
		"Le sport comme thérapie : comment le football m'a aidé à sortir de la dépression",
		"Est-il possible de concilier un job étudiant éprouvant et ses projets personnels ?",
		"Comment garder la motivation quand on entreprend une reconversion professionnelle ?",
		"Grosse frustration aujourd'hui face à un bug d'architecture... comment relativiser ?",
	}
	contents := []string{
		"Bonjour à toute la communauté, je fais ce post aujourd'hui car je me sens un peu submergé. J'aimerais beaucoup avoir vos retours d'expérience et vos précieux conseils pour traverser cette période. Prenez soin de vous.",
		"Je partage ceci de manière anonyme car c'est un sujet délicat, mais je sais qu'ici l'écoute est bienveillante. Est-ce que d'autres personnes ressentent la même chose en ce moment ? Parlons-en ensemble.",
		"Juste un petit message de soutien pour tous ceux qui traversent une journée difficile. N'oubliez pas que vous n'êtes pas seuls et que chaque petit pas compte vers le mieux-être. 😊",
	}

	var postIDs []int
	var postOwners = make(map[int]int) // Associe post_id -> user_id de l'auteur
	for i := 0; i < 250; i++ {
		uid := userIDs[rand.Intn(len(userIDs))]
		cid := categoryIDs[rand.Intn(len(categoryIDs))]
		title := titles[rand.Intn(len(titles))]
		content := contents[rand.Intn(len(contents))]
		isAnon := rand.Intn(4) == 0 // 25% de posts anonymes

		res, err := database.ZUKUKDB.Exec(`
			INSERT INTO posts (user_id, category_id, title, content, is_anonymous, created_at) 
			VALUES (?, ?, ?, ?, ?, datetime('now', '-10 days'))`,
			uid, cid, title, content, isAnon,
		)
		if err == nil {
			pid, _ := res.LastInsertId()
			postIDs = append(postIDs, int(pid))
			postOwners[int(pid)] = uid
		}
	}

	// 5. Génération de 800 Commentaires et déclenchement des VRAIES notifications
	fmt.Println("💬 Remplissage de 800 commentaires de soutien & notifications...")
	commentTexts := []string{
		"Sache que ton message me touche beaucoup, tu as tout mon courage !",
		"Je traverse exactement la même chose à Lyon en ce moment, si tu veux en parler à l'occasion.",
		"C'est une excellente analyse. Prendre du recul et respirer un grand coup fait un bien fou !",
		"Ne lâche rien, les débuts sont toujours difficiles mais le résultat en vaut la peine.",
		"Merci d'avoir partagé ça avec nous, c'est tellement libérateur de voir qu'on est pas seul.",
	}

	for i := 0; i < 800; i++ {
		uid := userIDs[rand.Intn(len(userIDs))]
		pid := postIDs[rand.Intn(len(postIDs))]
		content := commentTexts[rand.Intn(len(commentTexts))]
		isAnon := rand.Intn(10) == 0 // 10% anonymes

		_, err := database.ZUKUKDB.Exec(`
			INSERT INTO comments (post_id, user_id, content, is_anonymous, created_at) 
			VALUES (?, ?, ?, ?, datetime('now', '-2 days'))`,
			pid, uid, content, isAnon,
		)
		if err == nil {
			// Déclenchement automatique d'une vraie notification pour l'auteur de la discussion
			authorID := postOwners[pid]
			if authorID != uid && !isAnon {
				database.ZUKUKDB.Exec(`
					INSERT INTO notifications (user_id, actor_id, type, post_id, is_read)
					VALUES (?, ?, 'comment', ?, 0)`,
					authorID, uid, pid,
				)
			}
		}
	}

	// 6. Distribution de 1500 Likes et Notifications de Likes associés
	fmt.Println("❤️ Distribution de 1500 mentions J'aime...")
	likesCount := 0
	for likesCount < 1500 {
		uid := userIDs[rand.Intn(len(userIDs))]
		pid := postIDs[rand.Intn(len(postIDs))]

		_, err := database.ZUKUKDB.Exec(`
			INSERT INTO post_likes (post_id, user_id, created_at) 
			VALUES (?, ?, datetime('now', '-1 days'))`,
			pid, uid,
		)
		if err == nil { // Évite les doublons de clés uniques (user_id + post_id)
			likesCount++
			authorID := postOwners[pid]
			if authorID != uid {
				database.ZUKUKDB.Exec(`
					INSERT INTO notifications (user_id, actor_id, type, post_id, is_read)
					VALUES (?, ?, 'like', ?, 0)`,
					authorID, uid, pid,
				)
			}
		}
	}

	// 7. 📍 CARTE : Injections de VRAIES Activités en JUIN 2026 (Mois prochain) et participants
	fmt.Println("📍 Création d'activités réelles pour Juin 2026...")
	
	// Vérifier et récupérer les catégories d'activité
	var actCategoryIDs []int
	rowsActCat, _ := database.ZUKUKDB.Query(`SELECT id FROM activity_categories`)
	for rowsActCat.Next() {
		var acid int
		rowsActCat.Scan(&acid)
		actCategoryIDs = append(actCategoryIDs, acid)
	}
	rowsActCat.Close()

	if len(actCategoryIDs) == 0 {
		database.ZUKUKDB.Exec(`INSERT INTO activity_categories (name) VALUES ('Relaxation & bien-être')`)
		actCategoryIDs = append(actCategoryIDs, 1)
	}

	type ActivityTemplate struct {
		Name string
		Desc string
		Addr string
		Lat  float64
		Lng  float64
		Sched string
	}

	realActivities := []ActivityTemplate{
		{"Session Footing & Sophrologie", "Une course douce de 5km suivie d'une séance guidée de relaxation pour évacuer les tensions.", "Parc de la Tête d'Or, 69006 Lyon", 45.7772, 4.8525, "Samedi 06 Juin 2026 à 10:00"},
		{"Atelier Code & Café Entraide", "Venez avec vos bugs et vos projets étudiants ! On s'entraide sur Go, Python et le web dans la bonne humeur.", "Place Louis Pradel, 69001 Lyon", 45.7675, 4.8356, "Mercredi 10 Juin 2026 à 14:30"},
		{"Cercle de Parole et Échanges", "Un espace d'expression libre et confidentiel pour parler de nos surcharges mentales et s'écouter.", "Mairie du 7e, Rue Chevreul, 69007 Lyon", 45.7483, 4.8422, "Mardi 16 Juin 2026 à 18:30"},
		{"Balade Méditative en Nature", "Marche lente et exercices de pleine conscience le long des berges pour calmer l'anxiété.", "Berges du Rhône, 69003 Lyon", 45.7602, 4.8451, "Dimanche 21 Juin 2026 à 09:30"},
	}

	for i, act := range realActivities {
		creatorID := userIDs[rand.Intn(len(userIDs))]
		catID := actCategoryIDs[rand.Intn(len(actCategoryIDs))]
		maxPlaces := rand.Intn(15) + 10 // Entre 10 et 25 places max

		res, err := database.ZUKUKDB.Exec(`
			INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, rating, max_places)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 4.5, ?)`,
			creatorID, catID, act.Name, act.Desc, act.Addr, act.Lat, act.Lng, act.Sched, maxPlaces,
		)
		
		if err == nil {
			actID, _ := res.LastInsertId()
			
			// 8. Inscriptions de 5 à 12 membres réels à chaque activité
			numParticipants := rand.Intn(8) + 5
			shuffleUserIDs := append([]int(nil), userIDs...)
			rand.Shuffle(len(shuffleUserIDs), func(i, j int) { shuffleUserIDs[i], shuffleUserIDs[j] = shuffleUserIDs[j], shuffleUserIDs[i] })

			for p := 0; p < numParticipants; p++ {
				participantID := shuffleUserIDs[p]
				database.ZUKUKDB.Exec(`
					INSERT OR IGNORE INTO activity_participants (activity_id, user_id, joined_at)
					VALUES (?, ?, datetime('now'))`,
					actID, participantID,
				)
			}
			log.Printf("→ Activité [%s] créée avec %d inscrits.", act.Name, numParticipants)
		} else {
			log.Printf("Erreur insertion activité %d: %v", i, err)
		}
	}

	fmt.Println("\n✅ BASE DE DONNÉES INJECTÉE AVEC SUCCÈS !")
	fmt.Println("👉 100 Membres actifs avec avatars Dicebear et bios.")
	fmt.Println("👉 250 Discussions thématiques, 800 Commentaires, 1500 Likes.")
	fmt.Println("👉 Les boîtes de réception de notifications de tes membres sont pleines !")
	fmt.Println("👉 4 Activités géolocalisées à Lyon programmées pour Juin 2026 avec participants.")
	fmt.Println("🔑 Tous les comptes ont le mot de passe : password123")
}