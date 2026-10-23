package handler

import (
	"fmt"
	"net/http"
)

func ListRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Обработчик запросов работает")
}
