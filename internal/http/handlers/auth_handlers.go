package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/superlogarifm/goCalc-v3/internal/auth"
	"github.com/superlogarifm/goCalc-v3/internal/models"
	"github.com/superlogarifm/goCalc-v3/internal/storage"
)

type AuthHandlers struct {
	AuthService    *auth.AuthService
	UserRepository storage.UserRepository
}

func NewAuthHandlers(authService *auth.AuthService, userRepo storage.UserRepository) *AuthHandlers {
	return &AuthHandlers{
		AuthService:    authService,
		UserRepository: userRepo,
	}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	// Хешируем пароль
	hashedPassword, err := h.AuthService.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal server error (hashing)", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Login:        req.Login,
		PasswordHash: hashedPassword,
	}

	err = h.UserRepository.CreateUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			http.Error(w, "User with this login already exists", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error (db)", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "User registered successfully", "user_id": user.ID})
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepository.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error (db)", http.StatusInternalServerError)
		}
		return
	}

	if !h.AuthService.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := h.AuthService.GenerateToken(user.ID, user.Login)
	if err != nil {
		http.Error(w, "Internal server error (token gen)", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})

	/*
		//
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt_token",
			Value:    token,
			Expires:  time.Now().Add(h.AuthService.tokenDuration), // Устанавливаем время жизни cookie
			HttpOnly: true, // Запрещает доступ к cookie из JavaScript
			Secure:   r.TLS != nil, // Отправлять cookie только по HTTPS, если соединение защищено
			Path:     "/", // Cookie доступен для всех путей
			// SameSite: http.SameSiteLaxMode, // Защита от CSRF
		})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
	*/
}
