package handler

import (
	"fmt"
	"net/http"
)

func ListMaterials(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Обработчик материалов работает")
}

func DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Удаление материалов работает")
}
