package main

import (
    ZUKUKDB "zukuk/database"
    "log"
    "net/http"
)

func main() {
    ZUKUKDB.Init()


    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Forum database en ligne"))
    })

    log.Println("Serveur démarré sur http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}