package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ZUKUKDB *sql.DB

func Init() {
	var err error
	ZUKUKDB, err = sql.Open("sqlite3", "./database/zukuk.db?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		log.Fatal("Erreur ouverture BDD :", err)
	}
	if err = ZUKUKDB.Ping(); err != nil {
		log.Fatal("Erreur connexion BDD :", err)
	}
	ZUKUKDB.SetMaxOpenConns(25)
	ZUKUKDB.SetMaxIdleConns(5)
	ZUKUKDB.SetConnMaxLifetime(5 * time.Minute)

	createTables()
	seedCategories()
	migrateActivitiesTable()
	seedActivities()
	startCleaner()

	log.Println("Base de données initialisée avec succès")
}

func createTables() {
	queries := []string{

		// ── USERS ─────────────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS users (
			id             INTEGER  PRIMARY KEY AUTOINCREMENT,
			pseudo         TEXT     NOT NULL UNIQUE CHECK(length(pseudo) BETWEEN 3 AND 20),
			email          TEXT     NOT NULL UNIQUE,
			password_hash  TEXT     NOT NULL,
			avatar_url     TEXT     NOT NULL DEFAULT '',
			is_admin       INTEGER  NOT NULL DEFAULT 0,
			is_anonymous   INTEGER  NOT NULL DEFAULT 0,
			email_verified INTEGER  NOT NULL DEFAULT 0,
			verify_token   TEXT     NOT NULL DEFAULT '',
			verify_expires DATETIME,
			created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at     DATETIME DEFAULT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT     PRIMARY KEY,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS post_categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT    NOT NULL UNIQUE
		)`,

		`CREATE TABLE IF NOT EXISTS posts (
			id           INTEGER  PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			category_id  INTEGER  REFERENCES post_categories(id) ON DELETE SET NULL,
			title        TEXT     NOT NULL,
			content      TEXT     NOT NULL,
			is_anonymous INTEGER  NOT NULL DEFAULT 1,
			latitude     REAL     DEFAULT NULL,
			longitude    REAL     DEFAULT NULL,
			created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at   DATETIME DEFAULT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS comments (
			id           INTEGER  PRIMARY KEY AUTOINCREMENT,
			post_id      INTEGER  NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			user_id      INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content      TEXT     NOT NULL,
			is_anonymous INTEGER  NOT NULL DEFAULT 1,
			created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at   DATETIME DEFAULT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS post_likes (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			post_id    INTEGER  NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(post_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS comment_likes (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			comment_id INTEGER  NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(comment_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS mood_history (
			id          INTEGER  PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			mood        TEXT     NOT NULL CHECK(mood IN ('Bien','Calme','Triste','Anxieux','Colère')),
			recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_daily_mood
		 ON mood_history(user_id, date(recorded_at))`,

		`CREATE TABLE IF NOT EXISTS activity_categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT    NOT NULL UNIQUE
		)`,

		`CREATE TABLE IF NOT EXISTS activities (
			id          INTEGER  PRIMARY KEY AUTOINCREMENT,
			created_by  INTEGER  REFERENCES users(id) ON DELETE SET NULL,
			category_id INTEGER  REFERENCES activity_categories(id) ON DELETE SET NULL,
			name        TEXT     NOT NULL,
			description TEXT     NOT NULL DEFAULT '',
			address     TEXT     NOT NULL DEFAULT '',
			latitude    REAL     NOT NULL,
			longitude   REAL     NOT NULL,
			schedule    TEXT     NOT NULL DEFAULT '',
			rating      REAL     NOT NULL DEFAULT 0,
			max_places  INTEGER  NOT NULL DEFAULT 0,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS activity_participants (
			activity_id INTEGER  NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
			user_id     INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			joined_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (activity_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS deletion_log (
			id          INTEGER  PRIMARY KEY AUTOINCREMENT,
			table_name  TEXT     NOT NULL,
			record_id   INTEGER  NOT NULL,
			deleted_by  INTEGER  REFERENCES users(id) ON DELETE SET NULL,
			reason      TEXT     NOT NULL DEFAULT 'user_request',
			snapshot    TEXT     NOT NULL,
			deleted_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			purge_after DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS moderation_log (
			id          INTEGER  PRIMARY KEY AUTOINCREMENT,
			admin_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			action      TEXT     NOT NULL CHECK(action IN (
				'delete_post','delete_comment','delete_user',
				'restore_post','restore_comment','restore_user',
				'ban_user'
			)),
			target_type TEXT     NOT NULL,
			target_id   INTEGER  NOT NULL,
			reason      TEXT     DEFAULT '',
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_posts_user       ON posts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_category   ON posts(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created    ON posts(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_deleted    ON posts(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post    ON comments(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user    ON comments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_deleted ON comments(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_post  ON post_likes(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_user  ON post_likes(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comment_likes    ON comment_likes(comment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expiry  ON sessions(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user    ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_pseudo     ON users(pseudo)`,
		`CREATE INDEX IF NOT EXISTS idx_users_deleted    ON users(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_mood_user        ON mood_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activities_geo   ON activities(latitude, longitude)`,
		`CREATE INDEX IF NOT EXISTS idx_deletion_log     ON deletion_log(purge_after)`,
		`CREATE INDEX IF NOT EXISTS idx_moderation_log   ON moderation_log(admin_id)`,
	}

	for _, q := range queries {
		if _, err := ZUKUKDB.Exec(q); err != nil {
			log.Fatalf("Erreur création table :\n%s\n→ %v", q, err)
		}
	}
	log.Println("Tables créées avec succès")
}

func seedCategories() {
	for _, c := range []string{
		"Stress", "Solitude", "Études", "Anxiété", "Dépression",
		"Travail", "Relations", "Santé mentale", "Bien-être", "Sport", "Autre",
	} {
		ZUKUKDB.Exec(`INSERT OR IGNORE INTO post_categories (name) VALUES (?)`, c)
	}
	for _, c := range []string{
		"Activités sportives", "Relaxation & bien-être",
		"Groupes de parole", "Activités créatives",
		"Lieux sociaux", "Activités nature",
	} {
		ZUKUKDB.Exec(`INSERT OR IGNORE INTO activity_categories (name) VALUES (?)`, c)
	}
	log.Println("Catégories initialisées")
}

func migrateActivitiesTable() {
	rows, err := ZUKUKDB.Query(`PRAGMA table_info(activities)`)
	if err != nil {
		log.Println("Erreur migration activities:", err)
		return
	}
	defer rows.Close()

	existing := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			continue
		}
		existing[name] = true
	}

	if !existing["mood"] {
		if _, err := ZUKUKDB.Exec(`ALTER TABLE activities ADD COLUMN mood TEXT NOT NULL DEFAULT 'relaxant'`); err != nil {
			log.Println("Erreur ajout colonne mood:", err)
		}
	}
	if !existing["photo"] {
		if _, err := ZUKUKDB.Exec(`ALTER TABLE activities ADD COLUMN photo TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Println("Erreur ajout colonne photo:", err)
		}
	}
}

func seedActivities() {
	var count int
	if err := ZUKUKDB.QueryRow(`SELECT COUNT(*) FROM activities`).Scan(&count); err != nil {
		return
	}
	if count > 0 {
		return
	}

	type activitySeed struct {
		Name      string
		Category  string
		Mood      string
		Desc      string
		Address   string
		Lat       float64
		Lng       float64
		MaxPlaces int
	}

	seeds := []activitySeed{
		{"Cours de yoga en plein air", "Relaxation & bien-être", "relaxant", "Séance de yoga douce au parc, accessible à tous.", "Jardin du Luxembourg, Paris", 48.8462, 2.3372, 15},
		{"Atelier peinture", "Activités créatives", "relaxant", "Peinture en groupe avec matériel fourni.", "Vieux Port, Marseille", 43.2965, 5.3698, 12},
		{"Randonnée en forêt", "Activités nature", "actif", "Balade guidée dans la forêt de Fontainebleau.", "Forêt de Fontainebleau, France", 48.4000, 2.7000, 18},
		{"Cercle de parole bien-être", "Groupes de parole", "social", "Échange bienveillant autour du stress et de l'anxiété.", "Lyon 2e, Lyon", 45.7485, 4.8320, 20},
		{"Course collective", "Activités sportives", "actif", "Jogging en groupe le long du canal.", "Île de Nantes, Nantes", 47.2114, -1.5536, 14},
		{"Pause café zen", "Lieux sociaux", "social", "Rencontre chaleureuse pour parler de bien-être.", "Vieux Lille, Lille", 50.6373, 3.0635, 10},
		{"Atelier création musicale", "Activités créatives", "social", "Session de musique collective et partage d'idées.", "Le Suquet, Cannes", 43.5528, 7.0174, 12},
		{"Randonnée côtière", "Activités nature", "actif", "Marche face à la mer pour se ressourcer.", "Calanques, Marseille", 43.2120, 5.4387, 16},
		{"Séance méditation", "Relaxation & bien-être", "relaxant", "Méditation guidée dans un lieu calme.", "Parc de la Tête d'Or, Lyon", 45.7772, 4.8525, 18},
		{"Atelier photo urbaine", "Activités créatives", "faible-énergie", "Découverte de la photographie de rue.", "Quartier historique, Strasbourg", 48.5839, 7.7455, 12},
		{"Sortie vélo découverte", "Activités sportives", "actif", "Balade cycliste sur des pistes sécurisées.", "Bassin de la Villette, Paris", 48.8920, 2.3786, 14},
		{"Café discussion bien-être", "Lieux sociaux", "social", "Rencontre conviviale pour partager ses ressources.", "Place du Capitole, Toulouse", 43.6045, 1.4440, 18},
		{"Atelier écriture créative", "Activités créatives", "relaxant", "Écriture libre pour libérer ses émotions.", "Quartier Latin, Paris", 48.8497, 2.3449, 12},
		{"Randonnée nature", "Activités nature", "actif", "Exploration dans les vignobles bordelais.", "Bordeaux, France", 44.8378, -0.5792, 20},
		{"Séance de stretching", "Relaxation & bien-être", "relaxant", "Étirements doux pour relâcher les tensions.", "Parc Monceau, Paris", 48.8790, 2.3086, 12},
		{"Atelier cuisine saine", "Activités créatives", "social", "Cuisine collective simple et gourmande.", "Part-Dieu, Lyon", 45.7600, 4.8599, 14},
		{"Marche du soir", "Activités sportives", "actif", "Course douce au bord de la Garonne.", "Quais de Bordeaux, Bordeaux", 44.8410, -0.5720, 16},
		{"Groupe de parole parentale", "Groupes de parole", "social", "Échanges autour de la parentalité.", "Centre-ville, Rennes", 48.1173, -1.6778, 15},
		{"Yoga du matin", "Relaxation & bien-être", "relaxant", "Séance de yoga pour bien commencer la journée.", "Plage de Nice", 43.7009, 7.2684, 20},
		{"Atelier jardinage", "Activités nature", "faible-énergie", "Jardinage collectif en plein air.", "Parc de la Tête d'Or, Lyon", 45.7797, 4.8550, 12},
		{"Café littéraire", "Lieux sociaux", "social", "Lecture partagée autour de thèmes positifs.", "Rue de la République, Lyon", 45.7580, 4.8357, 14},
		{"Sortie randonnée urbaine", "Activités sportives", "actif", "Balade culturelle et sportive.", "Centre-ville, Annecy", 45.8992, 6.1294, 14},
		{"Atelier relaxation", "Relaxation & bien-être", "relaxant", "Exercices de respiration en petit groupe.", "Parc de la Ciutadella, Perpignan", 42.6887, 2.8940, 12},
		{"Cercle de parole créativité", "Groupes de parole", "social", "Partage autour de projets créatifs.", "Médiathèque, Nantes", 47.2184, -1.5536, 16},
		{"Marche en bord de mer", "Activités nature", "actif", "Promenade revitalisante sur la plage.", "Plage du Prado, Marseille", 43.2624, 5.3840, 18},
		{"Atelier poterie", "Activités créatives", "relaxant", "Modelage et détente en groupe.", "Croix-Rousse, Lyon", 45.7769, 4.8266, 10},
		{"Randonnée verte", "Activités nature", "actif", "Balade dans la forêt vosgienne.", "Nancy, France", 48.6936, 6.1845, 16},
		{"Café mindfulness", "Lieux sociaux", "social", "Pause calme et bienveillance en groupe.", "Place de la Comédie, Montpellier", 43.6108, 3.8767, 14},
		{"Atelier dessin nature", "Activités créatives", "relaxant", "Croquis en plein air.", "Saint-Malo, France", 48.6438, -2.0258, 12},
		{"Jogging du matin", "Activités sportives", "actif", "Course collective en bord de rivière.", "Quais de Seine, Paris", 48.8566, 2.3522, 20},
		{"Groupe de parole art-thérapie", "Groupes de parole", "social", "Échange autour des émotions par l'art.", "Strasbourg, France", 48.5734, 7.7521, 12},
		{"Atelier relaxation musicale", "Relaxation & bien-être", "relaxant", "Écoute et détente guidée.", "Parc Jean-Moulin, Toulouse", 43.5617, 1.4532, 12},
		{"Atelier tricot solidaire", "Lieux sociaux", "social", "Tricot collectif et discussions bienveillantes.", "Café associatif, Marseille", 43.2965, 5.3698, 10},
		{"Promenade nature", "Activités nature", "faible-énergie", "Balade douce au bord du lac.", "Lac Annecy, Annecy", 45.9000, 6.1290, 16},
		{"Atelier mosaïque", "Activités créatives", "relaxant", "Création artistique pas à pas.", "Vieux Nice, Nice", 43.7102, 7.2620, 12},
		{"Randonnée citadine", "Activités sportives", "actif", "Marche sportive dans la vieille ville.", "Montpellier, France", 43.6108, 3.8767, 18},
		{"Séance de méditation pleine conscience", "Relaxation & bien-être", "relaxant", "Méditation guidée au jardin.", "Parc Borély, Marseille", 43.2626, 5.3994, 14},
		{"Atelier slam", "Activités créatives", "social", "Écriture et partage de textes.", "Belleville, Paris", 48.8728, 2.3820, 12},
		{"Marche santé", "Activités sportives", "actif", "Marche collective pour se remettre en mouvement.", "Quais de la Loire, Orléans", 47.9029, 1.9093, 16},
		{"Groupe de parole anxiété", "Groupes de parole", "social", "Écoute et soutien entre participants.", "Centre-ville, Rouen", 49.4431, 1.0993, 14},
		{"Atelier bien-être créatif", "Activités créatives", "relaxant", "Création d'objets zen.", "Parc de la Colombière, Dijon", 47.3220, 5.0417, 12},
		{"Relaxation au lac", "Activités nature", "relaxant", "Séance face à l'eau pour se détendre.", "Lac du Bourget, Aix-les-Bains", 45.6860, 5.9078, 12},
	}

	for _, seed := range seeds {
		var categoryID int
		if err := ZUKUKDB.QueryRow(`SELECT id FROM activity_categories WHERE name = ?`, seed.Category).Scan(&categoryID); err != nil {
			continue
		}
		ZUKUKDB.Exec(`
			INSERT INTO activities (
				category_id, name, description, address, latitude, longitude,
				schedule, rating, max_places, mood, photo
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			categoryID, seed.Name, seed.Desc, seed.Address, seed.Lat, seed.Lng,
			"", 0, seed.MaxPlaces, seed.Mood, "",
		)
	}
}
func SoftDeletePost(postID, deletedBy int, reason string) error {
	var title, content string
	var userID int
	err := ZUKUKDB.QueryRow(
		`SELECT title, content, user_id FROM posts WHERE id = ? AND deleted_at IS NULL`, postID,
	).Scan(&title, &content, &userID)
	if err != nil {
		return fmt.Errorf("post introuvable : %w", err)
	}
	snapshot := fmt.Sprintf(`{"title":%q,"content":%q,"user_id":%d}`, title, content, userID)
	if _, err = ZUKUKDB.Exec(
		`UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, postID,
	); err != nil {
		return err
	}
	_, err = ZUKUKDB.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('posts', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		postID, deletedBy, reason, snapshot,
	)
	return err
}

func SoftDeleteComment(commentID, deletedBy int, reason string) error {
	var content string
	var postID, userID int
	err := ZUKUKDB.QueryRow(
		`SELECT content, post_id, user_id FROM comments WHERE id = ? AND deleted_at IS NULL`, commentID,
	).Scan(&content, &postID, &userID)
	if err != nil {
		return fmt.Errorf("commentaire introuvable : %w", err)
	}
	snapshot := fmt.Sprintf(`{"content":%q,"post_id":%d,"user_id":%d}`, content, postID, userID)
	ZUKUKDB.Exec(`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, commentID)
	_, err = ZUKUKDB.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('comments', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		commentID, deletedBy, reason, snapshot,
	)
	return err
}

func SoftDeleteUser(userID, deletedBy int, reason string) error {
	var pseudo, email string
	err := ZUKUKDB.QueryRow(
		`SELECT pseudo, email FROM users WHERE id = ? AND deleted_at IS NULL`, userID,
	).Scan(&pseudo, &email)
	if err != nil {
		return fmt.Errorf("utilisateur introuvable : %w", err)
	}
	snapshot := fmt.Sprintf(`{"pseudo":%q,"email":%q}`, pseudo, email)
	ZUKUKDB.Exec(`UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, userID)
	ZUKUKDB.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	_, err = ZUKUKDB.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('users', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		userID, deletedBy, reason, snapshot,
	)
	return err
}

func DeleteUserAndData(userID int) error {
	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("utilisateur introuvable")
	}

	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func RestorePost(postID, adminID int) error {
	_, err := ZUKUKDB.Exec(`UPDATE posts SET deleted_at = NULL WHERE id = ?`, postID)
	if err != nil {
		return err
	}
	logModerationAction(adminID, "restore_post", "post", postID, "restauré par admin")
	return nil
}

func RestoreComment(commentID, adminID int) error {
	_, err := ZUKUKDB.Exec(`UPDATE comments SET deleted_at = NULL WHERE id = ?`, commentID)
	if err != nil {
		return err
	}
	logModerationAction(adminID, "restore_comment", "comment", commentID, "restauré par admin")
	return nil
}

func RestoreUser(userID, adminID int) error {
	_, err := ZUKUKDB.Exec(`UPDATE users SET deleted_at = NULL WHERE id = ?`, userID)
	if err != nil {
		return err
	}
	logModerationAction(adminID, "restore_user", "user", userID, "restauré par admin")
	return nil
}

func logModerationAction(adminID int, action, targetType string, targetID int, reason string) {
	ZUKUKDB.Exec(`
		INSERT INTO moderation_log (admin_id, action, target_type, target_id, reason)
		VALUES (?, ?, ?, ?, ?)`,
		adminID, action, targetType, targetID, reason,
	)
}

func TogglePostLike(postID, userID int) (liked bool, count int, err error) {
	var exists int
	if err := ZUKUKDB.QueryRow(`
		SELECT COUNT(*)
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = ? AND p.deleted_at IS NULL AND u.deleted_at IS NULL`,
		postID,
	).Scan(&exists); err != nil {
		return false, 0, err
	}
	if exists == 0 {
		return false, 0, fmt.Errorf("publication introuvable")
	}

	res, execErr := ZUKUKDB.Exec(
		`INSERT OR IGNORE INTO post_likes (post_id, user_id) VALUES (?, ?)`, postID, userID,
	)
	if execErr != nil {
		return false, 0, execErr
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		ZUKUKDB.Exec(`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID)
		liked = false
	} else {
		liked = true
	}
	ZUKUKDB.QueryRow(`
		SELECT COUNT(*)
		FROM post_likes pl
		JOIN posts p ON p.id = pl.post_id
		WHERE pl.post_id = ? AND p.deleted_at IS NULL`,
		postID,
	).Scan(&count)
	return liked, count, nil
}

func ToggleCommentLike(commentID, userID int) (liked bool, count int, err error) {
	var exists int
	if err := ZUKUKDB.QueryRow(`
		SELECT COUNT(*)
		FROM comments c
		JOIN posts p ON p.id = c.post_id
		JOIN users u ON u.id = c.user_id
		WHERE c.id = ? AND c.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL`,
		commentID,
	).Scan(&exists); err != nil {
		return false, 0, err
	}
	if exists == 0 {
		return false, 0, fmt.Errorf("commentaire introuvable")
	}

	res, execErr := ZUKUKDB.Exec(
		`INSERT OR IGNORE INTO comment_likes (comment_id, user_id) VALUES (?, ?)`, commentID, userID,
	)
	if execErr != nil {
		return false, 0, execErr
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		ZUKUKDB.Exec(`DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		liked = false
	} else {
		liked = true
	}
	ZUKUKDB.QueryRow(`
		SELECT COUNT(*)
		FROM comment_likes cl
		JOIN comments c ON c.id = cl.comment_id
		WHERE cl.comment_id = ? AND c.deleted_at IS NULL`,
		commentID,
	).Scan(&count)
	return liked, count, nil
}

func startCleaner() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			cleanExpiredSessions()
			purgeExpiredDeletions()
		}
	}()
	log.Println("🧹 Cleaner démarré (intervalle : 1h)")
}

func cleanExpiredSessions() {
	res, err := ZUKUKDB.Exec(`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`)
	if err != nil {
		log.Println("Erreur nettoyage sessions :", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("Sessions expirées supprimées : %d", n)
	}
}

func purgeExpiredDeletions() {
	rows, err := ZUKUKDB.Query(`
		SELECT table_name, record_id FROM deletion_log
		WHERE purge_after < CURRENT_TIMESTAMP`)
	if err != nil {
		log.Println("Erreur lecture deletion_log :", err)
		return
	}
	defer rows.Close()
	allowed := map[string]bool{"posts": true, "comments": true, "users": true}
	count := 0
	for rows.Next() {
		var tableName string
		var recordID int
		if err := rows.Scan(&tableName, &recordID); err != nil {
			continue
		}
		if !allowed[tableName] {
			log.Printf("Nom de table non autorisé : %s", tableName)
			continue
		}
		ZUKUKDB.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, tableName), recordID)
		ZUKUKDB.Exec(`DELETE FROM deletion_log WHERE table_name = ? AND record_id = ?`, tableName, recordID)
		count++
	}
	if count > 0 {
		log.Printf("Suppressions définitives : %d", count)
	}
}
