package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "github.com/gorilla/mux"
)

const (
    defaultDBPath = "./zukuk.db"
    apiPrefix     = "/api"
)

var db *sql.DB

// Activity matches EXACTLY the columns of the activities table
type Activity struct {
    ID          int64    `json:"id,omitempty"`
    CreatedBy   *int64   `json:"created_by"`   // nullable
    CategoryID  int      `json:"category_id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Address     string   `json:"address"`
    Latitude    float64  `json:"latitude"`
    Longitude   float64  `json:"longitude"`
    Schedule    string   `json:"schedule"`
    Rating      float64  `json:"rating"`
    MaxPlaces   int      `json:"max_places"`
    CreatedAt   string   `json:"created_at,omitempty"`
    // Additional fields returned by API
    ParticipantsCount int  `json:"participants_count,omitempty"`
    IsParticipating   bool `json:"is_participating,omitempty"`
}

func main() {
    var err error
    dbPath := os.Getenv("ZUKUK_DB")
    if dbPath == "" {
        dbPath = defaultDBPath
    }

    db, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    if err := initSchema(db); err != nil {
        log.Fatalf("failed to init schema: %v", err)
    }

    r := mux.NewRouter()
    api := r.PathPrefix(apiPrefix).Subrouter()

    api.HandleFunc("/me", handleMe).Methods("GET")
    api.HandleFunc("/activities", handleGetActivities).Methods("GET")
    api.HandleFunc("/activities", handleCreateActivity).Methods("POST")
    api.HandleFunc("/activities/{id:[0-9]+}/join", handleJoinActivity).Methods("POST")

    // CORS middleware to allow credentials (client uses credentials: 'include')
    handler := corsMiddleware(r)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    addr := ":" + port
    log.Printf("Server listening on %s (DB: %s)", addr, dbPath)
    if err := http.ListenAndServe(addr, handler); err != nil {
        log.Fatalf("server error: %v", err)
    }
}

/* ---------- Schema ---------- */
func initSchema(db *sql.DB) error {
    schema := `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS activities (
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
);

CREATE TABLE IF NOT EXISTS activity_participants (
    activity_id INTEGER  NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id     INTEGER  NOT NULL,
    joined_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, user_id)
);
`
    _, err := db.Exec(schema)
    return err
}

/* ---------- Helpers ---------- */

// getCurrentUserID reads cookie "user_id" and returns user id or 0 if not present
func getCurrentUserID(r *http.Request) (int64, error) {
    c, err := r.Cookie("user_id")
    if err != nil {
        return 0, errors.New("no user cookie")
    }
    id, err := strconv.ParseInt(c.Value, 10, 64)
    if err != nil {
        return 0, errors.New("invalid user id cookie")
    }
    return id, nil
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
    writeJSON(w, code, map[string]string{"error": msg})
}

/* ---------- Handlers ---------- */

// GET /api/me
// Returns minimal user info based on cookie. If no cookie, returns 401 with {}
func handleMe(w http.ResponseWriter, r *http.Request) {
    id, err := getCurrentUserID(r)
    if err != nil || id == 0 {
        // Not authenticated
        w.WriteHeader(http.StatusUnauthorized)
        _, _ = w.Write([]byte(`{}`))
        return
    }
    // Minimal user object
    user := map[string]interface{}{
        "id": id,
    }
    writeJSON(w, http.StatusOK, user)
}

// GET /api/activities
// Returns activities with participants_count and is_participating for current user
func handleGetActivities(w http.ResponseWriter, r *http.Request) {
    // optional: allow query filters (category, text) from client later
    rows, err := db.Query(`SELECT id, created_by, category_id, name, description, address, latitude, longitude, schedule, rating, max_places, created_at FROM activities ORDER BY created_at DESC`)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to query activities")
        return
    }
    defer rows.Close()

    var result []Activity
    var currentUserID int64
    if uid, err := getCurrentUserID(r); err == nil {
        currentUserID = uid
    }

    for rows.Next() {
        var a Activity
        var createdBy sql.NullInt64
        var createdAt string
        if err := rows.Scan(&a.ID, &createdBy, &a.CategoryID, &a.Name, &a.Description, &a.Address, &a.Latitude, &a.Longitude, &a.Schedule, &a.Rating, &a.MaxPlaces, &createdAt); err != nil {
            log.Println("scan err:", err)
            continue
        }
        if createdBy.Valid {
            v := createdBy.Int64
            a.CreatedBy = &v
        } else {
            a.CreatedBy = nil
        }
        a.CreatedAt = createdAt

        // participants_count
        var cnt int
        if err := db.QueryRow(`SELECT COUNT(*) FROM activity_participants WHERE activity_id = ?`, a.ID).Scan(&cnt); err != nil {
            log.Println("count err:", err)
            cnt = 0
        }
        a.ParticipantsCount = cnt

        // is_participating
        if currentUserID > 0 {
            var exists int
            if err := db.QueryRow(`SELECT 1 FROM activity_participants WHERE activity_id = ? AND user_id = ? LIMIT 1`, a.ID, currentUserID).Scan(&exists); err == nil {
                a.IsParticipating = true
            } else {
                a.IsParticipating = false
            }
        } else {
            a.IsParticipating = false
        }

        result = append(result, a)
    }

    writeJSON(w, http.StatusOK, result)
}

// POST /api/activities
// Payload JSON must correspond EXACTLY to the columns of activities table.
// We accept created_at if provided (ISO8601), otherwise DB default will be used.
func handleCreateActivity(w http.ResponseWriter, r *http.Request) {
    var payload Activity
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json payload")
        return
    }

    // Validate required fields
    if strings.TrimSpace(payload.Name) == "" {
        writeError(w, http.StatusBadRequest, "name is required")
        return
    }
    // latitude/longitude must be present
    if payload.Latitude == 0 && payload.Longitude == 0 {
        writeError(w, http.StatusBadRequest, "latitude and longitude are required")
        return
    }

    // Insert with explicit columns to match schema
    tx, err := db.Begin()
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to begin tx")
        return
    }
    defer tx.Rollback()

    stmt := `INSERT INTO activities (created_by, category_id, name, description, address, latitude, longitude, schedule, rating, max_places, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))`
    res, err := tx.Exec(stmt,
        payload.CreatedBy,
        payload.CategoryID,
        payload.Name,
        payload.Description,
        payload.Address,
        payload.Latitude,
        payload.Longitude,
        payload.Schedule,
        payload.Rating,
        payload.MaxPlaces,
        nullString(payload.CreatedAt),
    )
    if err != nil {
        log.Println("insert err:", err)
        writeError(w, http.StatusInternalServerError, "failed to create activity")
        return
    }
    lastID, err := res.LastInsertId()
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to retrieve id")
        return
    }
    if err := tx.Commit(); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to commit")
        return
    }

    // Return created activity (fresh from DB)
    created, err := fetchActivityByID(lastID, 0) // no current user context
    if err != nil {
        writeError(w, http.StatusInternalServerError, "created but failed to fetch")
        return
    }
    writeJSON(w, http.StatusCreated, created)
}

// POST /api/activities/{id}/join
// Toggles participation for current user: if not participating -> insert; if participating -> delete.
// Returns updated activity (with participants_count and is_participating).
func handleJoinActivity(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    aid, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid activity id")
        return
    }

    userID, err := getCurrentUserID(r)
    if err != nil || userID == 0 {
        writeError(w, http.StatusUnauthorized, "authentication required")
        return
    }

    tx, err := db.Begin()
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to begin tx")
        return
    }
    defer tx.Rollback()

    // Check if activity exists
    var exists int
    if err := tx.QueryRow(`SELECT 1 FROM activities WHERE id = ? LIMIT 1`, aid).Scan(&exists); err != nil {
        writeError(w, http.StatusNotFound, "activity not found")
        return
    }

    // Check if already participating
    var present int
    err = tx.QueryRow(`SELECT 1 FROM activity_participants WHERE activity_id = ? AND user_id = ? LIMIT 1`, aid, userID).Scan(&present)
    if err == sql.ErrNoRows {
        // Not participating -> insert, but check capacity first
        var participants int
        var maxPlaces int
        if err := tx.QueryRow(`SELECT (SELECT COUNT(*) FROM activity_participants WHERE activity_id = ?) as cnt, max_places FROM activities WHERE id = ?`, aid, aid).Scan(&participants, &maxPlaces); err != nil {
            writeError(w, http.StatusInternalServerError, "failed to check capacity")
            return
        }
        if maxPlaces > 0 && participants >= maxPlaces {
            writeError(w, http.StatusConflict, "activity is full")
            return
        }
        _, err := tx.Exec(`INSERT INTO activity_participants (activity_id, user_id, joined_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, aid, userID)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to join")
            return
        }
    } else if err == nil {
        // Already participating -> remove (unjoin)
        _, err := tx.Exec(`DELETE FROM activity_participants WHERE activity_id = ? AND user_id = ?`, aid, userID)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to unjoin")
            return
        }
    } else {
        writeError(w, http.StatusInternalServerError, "failed to check participation")
        return
    }

    if err := tx.Commit(); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to commit")
        return
    }

    // Return updated activity with counts and is_participating
    updated, err := fetchActivityByID(aid, userID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to fetch updated activity")
        return
    }
    writeJSON(w, http.StatusOK, updated)
}

/* ---------- Utility DB fetch ---------- */

// fetchActivityByID returns activity with participants_count and is_participating for given userID (0 = anonymous)
func fetchActivityByID(id int64, userID int64) (Activity, error) {
    var a Activity
    var createdBy sql.NullInt64
    var createdAt string
    err := db.QueryRow(`SELECT id, created_by, category_id, name, description, address, latitude, longitude, schedule, rating, max_places, created_at FROM activities WHERE id = ? LIMIT 1`, id).
        Scan(&a.ID, &createdBy, &a.CategoryID, &a.Name, &a.Description, &a.Address, &a.Latitude, &a.Longitude, &a.Schedule, &a.Rating, &a.MaxPlaces, &createdAt)
    if err != nil {
        return a, err
    }
    if createdBy.Valid {
        v := createdBy.Int64
        a.CreatedBy = &v
    } else {
        a.CreatedBy = nil
    }
    a.CreatedAt = createdAt

    var cnt int
    if err := db.QueryRow(`SELECT COUNT(*) FROM activity_participants WHERE activity_id = ?`, id).Scan(&cnt); err != nil {
        cnt = 0
    }
    a.ParticipantsCount = cnt

    if userID > 0 {
        var present int
        if err := db.QueryRow(`SELECT 1 FROM activity_participants WHERE activity_id = ? AND user_id = ? LIMIT 1`, id, userID).Scan(&present); err == nil {
            a.IsParticipating = true
        } else {
            a.IsParticipating = false
        }
    } else {
        a.IsParticipating = false
    }

    return a, nil
}

/* ---------- Helpers ---------- */

func nullString(s string) interface{} {
    if strings.TrimSpace(s) == "" {
        return nil
    }
    // try to parse time to ensure valid format; if invalid, return raw string
    if _, err := time.Parse(time.RFC3339, s); err == nil {
        return s
    }
    // Accept other formats as-is
    return s
}

/* ---------- CORS Middleware ---------- */

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Adjust allowed origin in production
        origin := r.Header.Get("Origin")
        if origin == "" {
            origin = "*"
        }
        // Allow credentials and necessary headers
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Vary", "Origin")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}