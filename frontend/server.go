package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	staticDir := filepath.Join(".", "frontend")

	// ── Static assets ────────────────────────────────────────────────────────
	mux.Handle("/frontend/css/", http.StripPrefix("/frontend/css/",
		http.FileServer(http.Dir(filepath.Join(staticDir, "css")))))
	mux.Handle("/frontend/js/", http.StripPrefix("/frontend/js/",
		http.FileServer(http.Dir(filepath.Join(staticDir, "js")))))

	// ── Page helper ──────────────────────────────────────────────────────────
	page := func(file string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "html", file))
		}
	}

	// ── Static pages (exact match) ───────────────────────────────────────────
	mux.HandleFunc("/",                page("index.html"))
	mux.HandleFunc("/login",           page("login.html"))
	mux.HandleFunc("/register",        page("register.html"))
	mux.HandleFunc("/board",           page("board.html"))
	mux.HandleFunc("/profile",         page("profile.html"))
	mux.HandleFunc("/forgot-password", page("forgot-password.html"))
	mux.HandleFunc("/reset-password",  page("reset-password.html"))
	mux.HandleFunc("/network",         page("network.html"))

	// ── Dynamic pages (prefix match) ─────────────────────────────────────────
	// /post/123  → post.html (ID récupéré côté JS via URL)
	mux.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		// Refuse les tentatives de path-traversal
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "html", "post.html"))
	})

	// /profile/123 → public_profile.html
	mux.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "html", "public_profile.html"))
	})

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Serveur Frontend Zukuk démarré sur http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erreur démarrage serveur frontend : %v", err)
	}
}