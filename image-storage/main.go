package main

import (
	"net/http"
	"storage/handlers"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	// log.Info("Это информационное сообщение")
	// log.Warn("Это предупреждение")
	// log.Error("Это ошибка")
	fileHandler := handlers.NewFiles(l)

	sm := mux.NewRouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/products", fileHandler.GetProducts)

}
