package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var ctx = context.Background()
var rdb *redis.Client

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // адрес Redis сервера
		Password: "",               // пароль (если есть)
		DB:       0,                // номер базы данных
	})
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/main", mainHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		userData, err := rdb.Get(ctx, username).Result()
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		var user User
		if err := json.Unmarshal([]byte(userData), &user); err != nil {
			http.Error(w, "Error parsing user data", http.StatusInternalServerError)
			return
		}

		if user.Password != password {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    username,
			HttpOnly: true,
			Secure:   false,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/main", http.StatusSeeOther)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user := User{
			Username: username,
			Password: password,
		}

		userData, err := json.Marshal(user)
		if err != nil {
			http.Error(w, "Error marshaling user data", http.StatusInternalServerError)
			return
		}

		err = rdb.Set(ctx, username, userData, 0).Err()
		if err != nil {
			http.Error(w, "Error saving user data", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value

	tmpl := template.Must(template.ParseFiles("templates/main.html"))
	tmpl.Execute(w, map[string]string{"Username": username})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value

	// Удаляем пользователя из Redis
	err = rdb.Del(ctx, username).Err()
	if err != nil {
		http.Error(w, "Error deleting user from database", http.StatusInternalServerError)
		return
	}

	// Очищаем куки
	expiration := time.Now().Add(-time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		Expires:  expiration,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
