package main

import (
	"log"
	"net/http"

	"portal/internal/handler"
)

func main() {
	r := handler.NewRouter()

	addr := ":8080"
	log.Println("Сервер запущен на http://localhost" + addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
