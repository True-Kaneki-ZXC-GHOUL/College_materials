package handler

import (
	"encoding/json"
	"github.com/yourorg/portal/internal/handler/response"
	"github.com/yourorg/portal/internal/usecase"
	"net/http"
)

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}

	token, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"token": token})
}
