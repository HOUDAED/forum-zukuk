package main

import (
	"html/template"
    "net/http"
)

func main() {
    // Dans ton main() existant, ajoute simplement cette route :

http.HandleFunc("/carte", func(w http.ResponseWriter, r *http.Request) {
    // 1. On charge les fichiers nécessaires
    tmpl, err := template.ParseFiles(
        "./frontend/html/carte.html",
        "./frontend/html/sidebar.html",
    )
    if err != nil {
        http.Error(w, "Fichiers HTML introuvables : "+err.Error(), 500)
        return
    }

    // 2. On exécute spécifiquement le bloc 'content' défini dans carte.html
    // C'est cette ligne qui permet au navigateur d'afficher l'interface Zukuk
    err = tmpl.ExecuteTemplate(w, "content", nil)
    if err != nil {
        http.Error(w, "Erreur de rendu du template : "+err.Error(), 500)
    }
})

// Assure-toi que la gestion des fichiers statiques est bien présente une seule fois :
fs := http.FileServer(http.Dir("./frontend"))
http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

}