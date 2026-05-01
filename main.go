package main

import (
	"log"
	"net/http"
	"zukuk-forum/backend/database"
)

func main() {
	// Initialise la BDD (crée les tables, seed, lance le cleaner)
	database.Init()

	// Routeur de base — les handlers viendront sur feature/auth et feature/posts
	mux := http.NewServeMux()

	// Route de santé pour vérifier que le serveur tourne
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","db":"connected"}`))
	})

	// Servir les fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Serveur démarré sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Erreur serveur :", err)
	}
}