package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	// 1. Détection robuste du dossier
	baseDir := "."
	if _, err := os.Stat(filepath.Join(".", "html")); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(".", "frontend", "html")); err == nil {
			baseDir = "frontend"
		}
	}

	// 2. Fichiers statiques
	mux.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", http.FileServer(http.Dir(filepath.Join(baseDir, "css")))))
	mux.Handle("/frontend/js/", http.StripPrefix("/frontend/js/", http.FileServer(http.Dir(filepath.Join(baseDir, "js")))))

// 🔴 3. LA MAGIE : LE REVERSE PROXY 🔴
	// Toutes les requêtes qui commencent par "/api/" ou "/uploads/" sont envoyées au backend
	target, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	mux.Handle("/api/", proxy)
	mux.Handle("/uploads/", proxy) // 👈 AJOUTE CETTE LIGNE ICI !

	// 4. Pré-chargement des templates
	t := template.New("Zukuk")
	partialsPath := filepath.Join(baseDir, "html", "partials", "*.html")
	if _, err := t.ParseGlob(partialsPath); err != nil {
		log.Fatalf("Erreur Partials : %v\n", err)
	}

	pagesPath := filepath.Join(baseDir, "html", "*.html")
	if _, err := t.ParseGlob(pagesPath); err != nil {
		log.Fatalf("Erreur Pages HTML : %v\n", err)
	}

	// 5. Fonction d'affichage
	page := func(file string, active string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data := map[string]interface{}{"Active": active}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			t.ExecuteTemplate(w, file, data)
		}
	}

	// 6. Routes des pages
	mux.HandleFunc("/", page("index.html", "Accueil"))
	mux.HandleFunc("/login", page("login.html", ""))
	mux.HandleFunc("/register", page("register.html", ""))
	mux.HandleFunc("/board", page("board.html", "Accueil"))
	mux.HandleFunc("/profile", page("profile.html", "Profil"))
	mux.HandleFunc("/delete-account", page("delete-account.html", ""))
	mux.HandleFunc("/forgot-password", page("forgot-password.html", ""))
	mux.HandleFunc("/reset-password", page("reset-password.html", ""))
	mux.HandleFunc("/network", page("network.html", "Réseau"))
	mux.HandleFunc("/settings", page("settings.html", "Paramètres"))
	mux.HandleFunc("/carte", page("carte.html", "Carte"))

	mux.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") { http.NotFound(w, r); return }
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.ExecuteTemplate(w, "post.html", map[string]interface{}{"Active": ""})
	})

	mux.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") { http.NotFound(w, r); return }
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.ExecuteTemplate(w, "public_profile.html", map[string]interface{}{"Active": ""})
	})

	port := os.Getenv("FRONTEND_PORT")
	if port == "" { port = "3000" }
	fmt.Printf("🚀 Serveur Frontend Zukuk démarré sur http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}