package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	mux := http.NewServeMux()

	staticDir := filepath.Join(".", "frontend")

	// CORRECTION ICI : On ajoute "/frontend" dans les routes pour correspondre à tes fichiers HTML
	mux.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", http.FileServer(http.Dir(filepath.Join(staticDir, "css")))))
	mux.Handle("/frontend/js/", http.StripPrefix("/frontend/js/", http.FileServer(http.Dir(filepath.Join(staticDir, "js")))))

	pages := map[string]string{
		"/":                "index.html",
		"/login":           "login.html",
		"/register":        "register.html",
		"/board":           "board.html",
		"/profile":         "profile.html",
		"/forgot-password": "forgot-password.html",
		"/reset-password":  "reset-password.html",
	}

	for route, file := range pages {
		f := file
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "html", f))
		})
	}

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Serveur Frontend Zukuk démarré sur http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erreur démarrage serveur frontend : %v", err)
	}
}
