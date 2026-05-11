package api

// Dans le fichier main.go du backend (API)
r := mux.NewRouter()
api := r.PathPrefix("/api").Subrouter()

// Routes pour la carte et les activités
api.HandleFunc("/activities", handlers.GetActivitiesHandler).Methods("GET")
api.HandleFunc("/activities", handlers.CreateActivityHandler).Methods("POST")
api.HandleFunc("/activities/{id:[0-9]+}/join", handlers.ToggleJoinHandler).Methods("POST")