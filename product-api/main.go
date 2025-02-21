package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	//authh "simpleMSs/auth"
	"simpleMSs/handlers"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	//"golang.org/x/crypto/nacl/auth"
)

func connectDB() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(),
		"postgres://postgres:root@localhost:5433/postgres")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return conn
}

func main() {
	l := log.New(os.Stdout, "main-logger ", log.LstdFlags)
	db := connectDB()
	productHandler := handlers.NewProducts(l, db)

	sm := mux.NewRouter()

	corsHandler := gohandlers.CORS(
		gohandlers.AllowedOrigins([]string{"*"}),
		gohandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gohandlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Accept", "Accept-Language", "Content-Language", "Origin"}),
		gohandlers.AllowCredentials(),
	)

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", productHandler.GetProducts)
	getRouter.HandleFunc("/products/{id:[0-9]+}", productHandler.GetProducts)

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/products/{id:[0-9]+}", productHandler.UpdateProduct)
	putRouter.Use(productHandler.MiddlewareValidateProduct)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/products", productHandler.AddProduct)
	postRouter.Use(productHandler.MiddlewareValidateProduct)

	deleteRouter := sm.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/products/{id:[0-9]+}", productHandler.DeleteProduct)

	// Маршруты
	//sm.HandleFunc("/auth/login", authh.LoginHandler).Methods("POST")
	//sm.HandleFunc("/auth/verify", authh.authVerifyTokenHandler).Methods("GET")

	opts := middleware.RedocOpts{SpecURL: "/swagger.json"}
	sh := middleware.Redoc(opts, nil)
	getRouter.Handle("/docs", sh)
	getRouter.Handle("/swagger.json", http.FileServer(http.Dir("./")))

	s := &http.Server{
		Addr:         "localhost:8080",
		Handler:      corsHandler(sm),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		l.Println("Starting server on port 8080")
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(tc)
}
