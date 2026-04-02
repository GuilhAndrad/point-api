package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GuilhAndrad/point-api/internal/model"
	"github.com/GuilhAndrad/point-api/internal/service"
)

type AuthHandler struct {
    svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

type registerRequest struct {
    Name     string     `json:"name"`
    Email    string     `json:"email"`
    Password string     `json:"password"`
    Role     model.Role `json:"role"` // "employee" ou "admin"
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "payload inválido")
        return
    }

    user, err := h.svc.Register(r.Context(), req.Name, req.Email, req.Password, req.Role)
    if err != nil {
        switch err {
        case service.ErrEmailAlreadyExists:
            respondError(w, http.StatusConflict, err.Error())
        default:
            respondError(w, http.StatusInternalServerError, "erro ao criar usuário")
        }
        return
    }

    respondJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "payload inválido")
        return
    }

    token, user, err := h.svc.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        respondError(w, http.StatusUnauthorized, "credenciais inválidas")
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "token": token,
        "user":  user,
    })
}