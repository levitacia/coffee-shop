package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// User представляет структуру пользователя
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Credentials представляет структуру для входящих данных аутентификации
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse представляет структуру ответа с токеном
type TokenResponse struct {
	Token string `json:"token"`
}

var users = []User{
	{
		ID:       1,
		Username: "test@example.com",
		Password: "$2a$10$h.dl5J86rGH7I8bD9bZeZeci0pDt0.VwFTGujlnEaZXPf/q7vM5wO", // "password"
	},
}

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func main() {
	router := mux.NewRouter()

	// Маршруты
	router.HandleFunc("/auth/login", LoginHandler).Methods("POST")
	router.HandleFunc("/auth/verify", VerifyTokenHandler).Methods("GET")

	// Запуск сервера
	log.Printf("Starting auth service on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// LoginHandler обрабатывает запросы на аутентификацию
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Поиск пользователя
	var user *User
	for _, u := range users {
		if u.Username == creds.Username {
			user = &u
			break
		}
	}

	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Создание JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   user.Username,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Отправка токена
	json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
}

// VerifyTokenHandler проверяет валидность токена
func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Удаление префикса "Bearer "
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
