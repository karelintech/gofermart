package handlers

import (
	"encoding/json"
	"gofermart/internal/auth"
	"gofermart/internal/storage"
	"net/http"
	"strconv"
	"time"
)

// RegisterUser - регистрация нового пользователя
func (bc *BaseController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("POST /api/user/register")

	var user storage.UserModel

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		bc.Logger.Error(err)
		return
	}

	status, userID, err := bc.Storage.RegisterUser(r.Context(), user.Login, user.Password)
	if err != nil {
		w.WriteHeader(status)
		bc.Logger.Error(err)
		return
	}

	if status == http.StatusConflict {
		w.WriteHeader(http.StatusConflict)
		bc.Logger.Warn("Login is already used!")
		return
	}

	token, err := auth.GetJWTToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	stringUserID := strconv.Itoa(userID)
	http.SetCookie(w,
		&http.Cookie{
			Name:     "Token",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 1),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	http.SetCookie(w, &http.Cookie{
		Name:     "UserID",
		Value:    stringUserID,
		Expires:  time.Now().Add(time.Hour * 1),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

// LoginUser - авторизация пользователя
func (bc *BaseController) LoginUser(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("POST /api/user/login")

	var user storage.UserModel
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		bc.Logger.Error(err)
		return
	}

	login := auth.HashFunc(user.Login)
	password := auth.HashFunc(user.Password)

	status, userID, err := bc.Storage.AuthUser(r.Context(), login, password)
	if err != nil {
		w.WriteHeader(status)
		bc.Logger.Error(err)
		return
	}

	if status == http.StatusUnauthorized {
		w.WriteHeader(status)
		bc.Logger.Warn("Wrong login and/or password!")
		return
	}

	stringUserID := strconv.Itoa(userID)

	token, err := auth.GetJWTToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:     "Token",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 1),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	http.SetCookie(w, &http.Cookie{
		Name:     "UserID",
		Value:    stringUserID,
		Expires:  time.Now().Add(time.Hour * 1),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
