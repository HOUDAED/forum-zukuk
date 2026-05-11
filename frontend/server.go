package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	// Fichiers statiques (Déjà parfaitement gérés par ton HEAD)
	mux.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", http.FileServer(http.Dir(filepath.Join(baseDir, "css")))))
	mux.Handle("/frontend/js/", http.StripPrefix("/frontend/js/", http.FileServer(http.Dir(filepath.Join(baseDir, "js")))))

	// 2. Pré-chargement ULTRA STRICT des templates
	t := template.New("Zukuk")

	// 🔴 ON EXIGE LES PARTIALS
	partialsPath := filepath.Join(baseDir, "html", "partials", "*.html")
	if _, err := t.ParseGlob(partialsPath); err != nil {
		log.Fatalf("\n🚨 ERREUR FATALE (DÉMARRAGE STOPPÉ) 🚨\nJe ne trouve aucun fichier dans : %s\n👉 Vérifie que le dossier 'partials' existe bien.\nErreur technique : %v\n\n", partialsPath, err)
	}

	// 🔴 ON EXIGE LES PAGES (carte.html sera chargée automatiquement ici !)
	pagesPath := filepath.Join(baseDir, "html", "*.html")
	if _, err := t.ParseGlob(pagesPath); err != nil {
		log.Fatalf("\n🚨 ERREUR FATALE (DÉMARRAGE STOPPÉ) 🚨\nJe ne trouve aucune page HTML dans : %s\nErreur technique : %v\n\n", pagesPath, err)
	}

	// 3. Confirmation visuelle dans le terminal
	fmt.Println("---------------------------------------------------")
	fmt.Println("✅ SUCCÈS ! Les templates suivants sont chargés :")
	for _, tmpl := range t.Templates() {
		fmt.Println("   -", tmpl.Name())
	}
	fmt.Println("---------------------------------------------------")

	// 4. Fonction pour afficher les pages
	page := func(file string, active string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data := map[string]interface{}{"Active": active}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := t.ExecuteTemplate(w, file, data); err != nil {
				http.Error(w, "Erreur d'affichage : "+err.Error(), http.StatusInternalServerError)
			}
		}
	}

	// 5. Routes
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

	// 📍 LA NOUVELLE ROUTE DE LA CARTE EST INTÉGRÉE ICI 📍
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