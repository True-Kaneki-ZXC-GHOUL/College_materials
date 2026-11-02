package handler

import (
	"encoding/json"
	"net/http"
)

// decodeJSON читает тело запроса в v.
// Вынесено в отдельный файл, т.к. используется несколькими хендлерами.
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}