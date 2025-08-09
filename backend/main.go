package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"notes/handlers"
	"notes/middleware"
	"notes/storage"
)

func main() {
	notesDb, err := storage.InitNotesDB()
	if err != nil {
		log.Fatal(err)
	}
	usersDb, err := storage.InitUsersDB()
	if err != nil {
		log.Fatal(err)
	}
	defer notesDb.Close()
	defer usersDb.Close()

	r := mux.NewRouter()

	// 2. Публичные маршруты
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/register", handlers.Register).Methods("POST")

	// 3. Защищенные маршруты
	authRouter := r.PathPrefix("/api").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)
	authRouter.HandleFunc("/notes", handlers.CreateNote).Methods("POST")
	authRouter.HandleFunc("/notes/{id}", handlers.GetNote).Methods("GET")
	authRouter.HandleFunc("/notes", handlers.GetNotes).Methods("GET")
	authRouter.HandleFunc("/notes/{id}", handlers.DeleteNote).Methods("DELETE")
	authRouter.HandleFunc("/notes/{id}", handlers.UpdateNote).Methods("PUT") // Исправлено UPDATE на PUT

	// 1. Статические файлы через mux
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))
	
	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", r)) // Убрали 0.0.0.0 для единообразия
}
