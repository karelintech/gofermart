package handlers

import (
	"encoding/json"
	"gofermart/internal/auth"
	"gofermart/internal/storage"
	"net/http"
)

// GetBalance - проверка баланса с баллами
func (bc *BaseController) GetBalance(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("GET /api/user/balance")

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	response := &struct {
		Current   int `json:"current"`
		Withdrawn int `json:"withdrawn"`
	}{}

	response.Current, response.Withdrawn, err = bc.Storage.GetUserBalance(r.Context(), cookieUserID)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}
}

// CreateWithdraw - запрос на списание средств
func (bc *BaseController) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("POST /api/user/balance/withdraw")

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	request := struct {
		OrderNum string `json:"order"`
		Sum      int    `json:"sum"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		bc.Logger.Error(err)
		return
	}

	if request.Sum <= 0 {
		w.WriteHeader(http.StatusPaymentRequired)
		bc.Logger.Warn("Payment Required")
		return
	}

	order, isEmpty, err := bc.Storage.GetOrder(r.Context(), &storage.OrderModel{
		Number: request.OrderNum,
	})

	if order.UserID != cookieUserID {
		w.WriteHeader(http.StatusConflict)
		bc.Logger.Warn("Froad")
		return
	}

	if isEmpty {
		w.WriteHeader(http.StatusUnprocessableEntity)
		bc.Logger.Error(err)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	balance, _, err := bc.Storage.GetUserBalance(r.Context(), order.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	if request.Sum > balance {
		w.WriteHeader(http.StatusPaymentRequired)
		bc.Logger.Warn("Payment Required")
		return
	}

	err = bc.Storage.CreateWithDraw(r.Context(), cookieUserID, request.OrderNum, request.Sum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetWithdraw  - получение информации о выводе баллов
func (bc *BaseController) GetWithdraw(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("POST /api/user/balance/withdraw")

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	rows, err := bc.Storage.GetWithDraw(r.Context(), cookieUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	if rows == nil {
		w.WriteHeader(http.StatusNoContent)
		bc.Logger.Info("No Content")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(rows); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		bc.Logger.Error(err)
	}
}
