package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	startCleaner()

	log.Println("Base de données initialisée avec succès")
}

func createTables() {
	queries := []string{
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
			paused_until   DATETIME DEFAULT NULL,
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
		`CREATE TABLE IF NOT EXISTS password_resets (
			token      TEXT     PRIMARY KEY,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS password_history (
			id            INTEGER  PRIMARY KEY AUTOINCREMENT,
			user_id       INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			password_hash TEXT     NOT NULL,
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_daily_mood ON mood_history(user_id, date(recorded_at))`,
		`CREATE TABLE IF NOT EXISTS activity_categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT    NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS connection_history (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			ip_address TEXT     NOT NULL DEFAULT '',
			user_agent TEXT     NOT NULL DEFAULT '',
			status     TEXT     NOT NULL DEFAULT 'Réussie',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_history_user    ON connection_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_history_created ON connection_history(created_at DESC)`,
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
				'restore_post','restore_comment','restore_user','ban_user'
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

// ═════════════════════════════════════════════════════════════════════════════
// SUPPRESSION — trois niveaux cohérents
//
//  1. SoftDeleteUser   → deleted_at + snapshot dans deletion_log (30 j)
//  2. SoftDeletePost   → idem pour les posts
//  3. SoftDeleteComment→ idem pour les commentaires
//
// Le cleaner (startCleaner) exécute la purge définitive après purge_after.
// DeleteUserAndData appelle SoftDeleteUser (plus de hard DELETE immédiat).
// ═════════════════════════════════════════════════════════════════════════════

// SoftDeleteUser masque le compte, invalide les sessions et enregistre
// un snapshot dans deletion_log. La purge définitive interviendra après 30 jours.
func SoftDeleteUser(userID int, deletedBy int, reason string) error {
	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Vérifier que l'utilisateur existe et n'est pas déjà supprimé
	var pseudo, email string
	if err := tx.QueryRow(
		`SELECT pseudo, email FROM users WHERE id = ? AND deleted_at IS NULL`, userID,
	).Scan(&pseudo, &email); err != nil {
		return fmt.Errorf("utilisateur introuvable ou déjà supprimé")
	}

	// 1. Soft delete : marquer deleted_at
	if _, err := tx.Exec(
		`UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, userID,
	); err != nil {
		return fmt.Errorf("erreur soft delete user : %w", err)
	}

	// 2. Invalider toutes les sessions immédiatement
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("erreur suppression sessions : %w", err)
	}

	// 3. Invalider les tokens de réinitialisation de mot de passe
	tx.Exec(`DELETE FROM password_resets WHERE user_id = ?`, userID)

	// 4. Enregistrer un snapshot dans deletion_log (purge après 30 jours)
	snapshot := fmt.Sprintf(`{"pseudo":%q,"email":%q,"user_id":%d}`, pseudo, email, userID)
	if _, err := tx.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('users', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		userID, deletedBy, reason, snapshot,
	); err != nil {
		return fmt.Errorf("erreur deletion_log : %w", err)
	}

	return tx.Commit()
}

// DeleteUserAndData est appelé depuis le handler HTTP.
// Il utilise SoftDeleteUser : le compte est masqué immédiatement,
// la purge définitive intervient après 30 jours via le cleaner.
func DeleteUserAndData(userID int) error {
	return SoftDeleteUser(userID, userID, "user_request")
}

// SoftDeletePost soft-delete un post et tous ses commentaires,
// avec un snapshot dans deletion_log.
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

	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`UPDATE posts    SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, postID)
	tx.Exec(`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE post_id = ?`, postID)
	tx.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('posts', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		postID, deletedBy, reason, snapshot)

	return tx.Commit()
}

// SoftDeleteComment soft-delete un commentaire avec snapshot.
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

	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, commentID)
	tx.Exec(`
		INSERT INTO deletion_log (table_name, record_id, deleted_by, reason, snapshot, purge_after)
		VALUES ('comments', ?, ?, ?, ?, datetime('now', '+30 days'))`,
		commentID, deletedBy, reason, snapshot)

	return tx.Commit()
}

// ═════════════════════════════════════════════════════════════════════════════
// MOT DE PASSE
// ═════════════════════════════════════════════════════════════════════════════

func CreateResetToken(email string, token string) error {
	var userID int
	err := ZUKUKDB.QueryRow(`SELECT id FROM users WHERE email = ? AND deleted_at IS NULL`, email).Scan(&userID)
	if err != nil {
		return err
	}
	ZUKUKDB.Exec(`DELETE FROM password_resets WHERE user_id = ?`, userID)
	_, err = ZUKUKDB.Exec(`
		INSERT INTO password_resets (token, user_id, expires_at)
		VALUES (?, ?, datetime('now', '+15 minutes'))`,
		token, userID,
	)
	return err
}

func ValidateResetToken(token string) (int, error) {
	var userID int
	err := ZUKUKDB.QueryRow(`
		SELECT user_id FROM password_resets
		WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`,
		token,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("token invalide ou expiré")
	}
	return userID, nil
}

// IsPasswordInHistory vérifie si le mot de passe en clair correspond à l'un
// des 5 derniers hashes stockés (bcrypt). Corrige le bug original qui ne
// comparait jamais réellement les hashes.
func IsPasswordInHistory(userID int, plainPassword string) (bool, error) {
	rows, err := ZUKUKDB.Query(
		`SELECT password_hash FROM password_history
		 WHERE user_id = ?
		 ORDER BY created_at DESC LIMIT 5`,
		userID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			continue
		}
		// Comparaison bcrypt réelle (corrige le bug : avant, _ = hash ne comparait rien)
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)) == nil {
			return true, nil // mot de passe déjà utilisé
		}
	}
	return false, nil
}

func UpdatePassword(userID int, newPasswordHash string) error {
	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		newPasswordHash, userID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`INSERT INTO password_history (user_id, password_hash) VALUES (?, ?)`,
		userID, newPasswordHash,
	); err != nil {
		return err
	}
	tx.Exec(`DELETE FROM password_resets WHERE user_id = ?`, userID)
	tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)

	return tx.Commit()
}

// ═════════════════════════════════════════════════════════════════════════════
// LIKES
// ═════════════════════════════════════════════════════════════════════════════

func TogglePostLike(postID, userID int) (liked bool, count int, err error) {
	var exists int
	if err := ZUKUKDB.QueryRow(`
		SELECT COUNT(*) FROM posts p
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
		SELECT COUNT(*) FROM post_likes pl
		JOIN posts p ON p.id = pl.post_id
		WHERE pl.post_id = ? AND p.deleted_at IS NULL`,
		postID,
	).Scan(&count)

	return liked, count, nil
}

// ═════════════════════════════════════════════════════════════════════════════
// CLEANER — tâche de fond toutes les heures
// ═════════════════════════════════════════════════════════════════════════════

func startCleaner() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			cleanExpiredSessions()
			purgeExpiredDeletions()
		}
	}()
	log.Println("Cleaner démarré (intervalle : 1h)")
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

// purgeExpiredDeletions supprime définitivement les enregistrements dont
// purge_after est dépassé (soft-deleted depuis plus de 30 jours).
//
// Pour les users : ON DELETE CASCADE supprime automatiquement sessions,
// posts, comments, likes, mood_history, connection_history, etc.
func purgeExpiredDeletions() {
	rows, err := ZUKUKDB.Query(`
		SELECT table_name, record_id FROM deletion_log
		WHERE purge_after < CURRENT_TIMESTAMP`)
	if err != nil {
		log.Println("Erreur lecture deletion_log :", err)
		return
	}
	defer rows.Close()

	// Tables autorisées (whitelist pour éviter les injections SQL)
	allowed := map[string]bool{
		"users":    true,
		"posts":    true,
		"comments": true,
	}

	var toDelete []struct {
		table string
		id    int
	}
	for rows.Next() {
		var tableName string
		var recordID int
		if err := rows.Scan(&tableName, &recordID); err != nil {
			continue
		}
		if !allowed[tableName] {
			log.Printf("Purge refusée pour table non autorisée : %s", tableName)
			continue
		}
		toDelete = append(toDelete, struct {
			table string
			id    int
		}{tableName, recordID})
	}
	rows.Close() // fermer avant d'exécuter des modifications

	for _, item := range toDelete {
		_, err := ZUKUKDB.Exec(
			fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, item.table), item.id,
		)
		if err != nil {
			log.Printf("Erreur purge %s#%d : %v", item.table, item.id, err)
			continue
		}
		ZUKUKDB.Exec(`
			DELETE FROM deletion_log WHERE table_name = ? AND record_id = ?`,
			item.table, item.id,
		)
		log.Printf("Purgé définitivement : %s#%d", item.table, item.id)
	}
}

// logModerationAction enregistre une action admin dans moderation_log.
func logModerationAction(adminID int, action, targetType string, targetID int, reason string) {
	ZUKUKDB.Exec(`
		INSERT INTO moderation_log (admin_id, action, target_type, target_id, reason)
		VALUES (?, ?, ?, ?, ?)`,
		adminID, action, targetType, targetID, reason,
	)
}

// PauseUserAccount met le compte en pause.
// Tu peux définir une durée, ou simplement mettre CURRENT_TIMESTAMP pour indiquer que c'est en pause jusqu'à réactivation.
func PauseUserAccount(userID int) error {
	tx, err := ZUKUKDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Mettre à jour la colonne paused_until (ici on met la date actuelle comme marqueur de pause)
	if _, err := tx.Exec(
		`UPDATE users SET paused_until = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, 
		userID,
	); err != nil {
		return fmt.Errorf("erreur lors de la mise en pause : %w", err)
	}

	// 2. Supprimer toutes les sessions actives pour forcer la déconnexion
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("erreur suppression sessions : %w", err)
	}

	return tx.Commit()
}

// ResumeUserAccount annule la pause du compte
func ResumeUserAccount(userID int) error {
	_, err := ZUKUKDB.Exec(
		`UPDATE users SET paused_until = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, 
		userID,
	)
	if err != nil {
		return fmt.Errorf("erreur lors de la réactivation du compte : %w", err)
	}
	return nil
}