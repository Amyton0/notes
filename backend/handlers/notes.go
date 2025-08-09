package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"notes/models"
	"notes/storage"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func CreateNote(w http.ResponseWriter, r *http.Request) {
	var note models.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := r.Context().Value("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))
	note.UserID = userID

	db := storage.GetNotesDB() // Предполагаем, что storage.GetDB() возвращает *sql.DB
	result, err := db.Exec("INSERT INTO notes (title, content, user_id) VALUES (?, ?, ?)", note.Title, note.Content, note.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	note.ID = int(id)
	note.CreatedAt = time.Now()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func GetNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	db := storage.GetNotesDB()
	var note models.Note
	err = db.QueryRow(`
		SELECT id, title, content, created_at, user_id 
		FROM notes 
		WHERE id = ?`, id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UserID)

	// 3. Обрабатываем ошибки
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := r.Context().Value("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	if note.UserID != userID {
		http.Error(w, "Access Denied", http.StatusBadRequest)
		return
	}

	// 4. Возвращаем заметку в JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем user_id из токена
	token := r.Context().Value("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// 2. Запрос к БД с фильтром по user_id
	db := storage.GetNotesDB()
	rows, err := db.Query("SELECT id, title, content FROM notes WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("DB error: %v", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 3. Формируем ответ
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content); err != nil {
			http.Error(w, `{"error":"Data parsing error"}`, http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	// 4. Явно указываем Content-Type
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notes); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	db := storage.GetNotesDB()

	var note models.Note
	err = db.QueryRow(`
		SELECT id, title, content, created_at, user_id 
		FROM notes 
		WHERE id = ?`, id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UserID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := r.Context().Value("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	if note.UserID != userID {
		http.Error(w, "Access Denied", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM notes WHERE id = ?", id) // Исправлено == на =
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var newNote models.Note
	if err := json.NewDecoder(r.Body).Decode(&newNote); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := storage.GetNotesDB()

	var note models.Note
	err = db.QueryRow(`
		SELECT id, title, content, created_at, user_id 
		FROM notes 
		WHERE id = ?`, id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UserID)

	note.Title = newNote.Title
	note.Content = newNote.Content

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := r.Context().Value("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	if note.UserID != userID {
		http.Error(w, "Access Denied", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE notes SET title = ?, content = ? WHERE id = ?", note.Title, note.Content, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
