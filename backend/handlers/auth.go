package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"notes/storage"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"notes/models"
)

var JwtKey = []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30") // Замените на случайную строку!

func Login(w http.ResponseWriter, r *http.Request) {
	// Логируем заголовки
	log.Printf("Headers: %v", r.Header)

	// Читаем тело запроса
	body, _ := io.ReadAll(r.Body)
	log.Printf("Raw body: %s", string(body))

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		log.Println("Invalid Content-Type")
		http.Error(w, `{"error":"Требуется application/json"}`, http.StatusBadRequest)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(body, &creds); err != nil {
		log.Printf("JSON parse error: %v", err)
		http.Error(w, `{"error":"Неверный JSON"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// Получаем данные из формы
	username := creds.Username
	password := creds.Password

	// Здесь должна быть проверка пользователя в БД
	user, err := getUserFromDB(username)
	if err != nil || user.CheckPassword(password) != nil {
		JSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создаем JWT-токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		JSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Отправляем токен клиенту
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"token":"` + tokenString + `"}`))
}

func getUserFromDB(username string) (models.User, error) {
	db := storage.GetUsersDB()
	var user models.User
	err := db.QueryRow(`
		SELECT id, username, password
		FROM users 
		WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.Password)

	return user, err
}

func Register(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем данные из запроса
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 2. Проверяем обязательные поля
	if username == "" || password == "" {
		JSONError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// 3. Создаем пользователя и хешируем пароль
	user := models.User{
		Username: username,
		Password: password, // Пароль будет захеширован
	}

	if err := user.HashPassword(); err != nil {
		JSONError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// 4. Сохраняем в БД (исправлена ошибка в параметрах)
	db := storage.GetUsersDB()
	result, err := db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		user.Username, // Используем username, а не password дважды
		user.Password, // Здесь уже хешированный пароль
	)

	if err != nil {
		// Проверяем на дубликат username
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			JSONError(w, "Username already exists", http.StatusConflict)
		} else {
			JSONError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// 5. Возвращаем ответ без пароля
	id, _ := result.LastInsertId()
	response := struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}{
		ID:       int(id),
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
