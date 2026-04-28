package zukuk
import (
    "log"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var ZUKUKDB *sql.DB

func Init() {
    var err error
    ZUKUKDB, err = sql.Open("sqlite3", "./zukuk.db")
    if err != nil {
        log.Fatal("Erreur ouverture BDD :", err)
    }

    if err = ZUKUKDB.Ping(); err != nil {
        log.Fatal("Erreur connexion BDD :", err)
    }

    createTables()
    log.Println("Base de données initialisée")
}

func createTables() {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            First_name TEXT    NOT NULL,
            Last_name  TEXT    NOT NULL,
            username   TEXT    NOT NULL UNIQUE,
            email      TEXT    NOT NULL UNIQUE,
            password   TEXT    NOT NULL,           -- bcrypt hash
            avatar_url TEXT    DEFAULT '',
            bio        TEXT    DEFAULT '',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,

        `CREATE TABLE IF NOT EXISTS sessions (
            id         TEXT    PRIMARY KEY,        -- UUID généré côté Go
            user_id    INTEGER NOT NULL,
            expires_at DATETIME NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,

        
        `CREATE TABLE IF NOT EXISTS post_categories (
            id   INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT    NOT NULL UNIQUE            -- ex: "Suicide", "Addictions", "Anxiété"
        );`,

        
        `CREATE TABLE IF NOT EXISTS posts (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id     INTEGER NOT NULL,
            category_id INTEGER,
            title       TEXT    NOT NULL,
            content     TEXT    NOT NULL,
            likes       INTEGER DEFAULT 0,
            created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id)     REFERENCES users(id)      ON DELETE CASCADE,
            FOREIGN KEY (category_id) REFERENCES post_categories(id) ON DELETE SET NULL
        );`,

        `CREATE TABLE IF NOT EXISTS comments (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id    INTEGER NOT NULL,
            user_id    INTEGER NOT NULL,
            content    TEXT    NOT NULL,
            likes      INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES posts(id)    ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id)    ON DELETE CASCADE
        );`,

        `CREATE TABLE IF NOT EXISTS likes (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id    INTEGER NOT NULL,
            post_id    INTEGER,
            comment_id INTEGER,
            UNIQUE(user_id, post_id),
            UNIQUE(user_id, comment_id),
            FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE,
            FOREIGN KEY (post_id)    REFERENCES posts(id)    ON DELETE CASCADE,
            FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
        );`,

        `CREATE TABLE IF NOT EXISTS activities_categories (
            id   INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT    NOT NULL UNIQUE
        );`,

        `CREATE TABLE IF NOT EXISTS activities (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            title       TEXT    NOT NULL,
            description TEXT    NOT NULL,
            date        DATETIME,
            max_places  INTEGER DEFAULT 0,          
            created_by  INTEGER NOT NULL,
            category_id INTEGER,
            created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
            FOREIGN KEY (category_id) REFERENCES activities_categories(id) ON DELETE SET NULL
        );`,

        
        `CREATE TABLE IF NOT EXISTS activity_participants (
            activity_id INTEGER NOT NULL,
            user_id     INTEGER NOT NULL,
            joined_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (activity_id, user_id),
            FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
            FOREIGN KEY (user_id)     REFERENCES users(id)      ON DELETE CASCADE
        );`,

        
        `INSERT OR IGNORE INTO post_categories (name) VALUES
            ('Stress'),
            ('Suicide'),
            ('Solitude'),
            ('Dépression'), 
            ('Bien-être'),
            ('Addictions'),
            ('Anxiété'),
            ('Etudes'),
            ('Autre');`,

        `INSERT OR IGNORE INTO activities_categories (name) VALUES
            ('sport'),
            ('relaxation et bien-être'),
            ('groupe de parole'),
            ('Activités créatives'),
            ('activités nature'),
            ('lieux sociaux'),
            ('Autre');`,
    }

    for _, q := range queries {
        if _, err := ZUKUKDB.Exec(q); err != nil {
            log.Fatal("Erreur création table :\n", q, "\n→", err)
        }
    }
}
