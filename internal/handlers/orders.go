package handlers

import (
	"database/sql"
	"encoding/json"
	"gofermart/internal/auth"
	"gofermart/internal/storage"
	"net/http"
)

// CreateOrder - создание заказа
func (bc *BaseController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("POST /api/user/orders")

	userOrder := storage.OrderModel{}
	err := json.NewDecoder(r.Body).Decode(&userOrder)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		bc.Logger.Error(err)
		return
	}
	defer r.Body.Close()

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	order, isEmpty, err := bc.Storage.GetOrder(r.Context(), &userOrder)

	if isEmpty {
		bc.Storage.CreateOrder(r.Context(), cookieUserID, &userOrder)
		w.WriteHeader(http.StatusAccepted)
		bc.Logger.Info("Order was created with number - ", userOrder.Number)
		return
	}

	if order.UserID != cookieUserID {
		w.WriteHeader(http.StatusConflict)
		bc.Logger.Warn("Froad")
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetOrders - получение списка всех заказов пользователя
func (bc *BaseController) GetOrders(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("GET /api/user/orders")

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	orders, err := bc.Storage.GetAllOrders(r.Context(), cookieUserID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "No Content", http.StatusNoContent)
		bc.Logger.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}
}

// GetOrderInfo - информация о заказе
func (bc *BaseController) GetOrderInfo(w http.ResponseWriter, r *http.Request) {
	bc.Logger.Info("GET /api/user/orders/", r.PathValue("number"))

	cookieUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		bc.Logger.Error(err)
		return
	}

	var userOrder storage.OrderModel
	userOrder.Number = r.PathValue("number")
	DBOrder, isEmpty, err := bc.Storage.GetOrder(r.Context(), &userOrder)

	if DBOrder.UserID != cookieUserID {
		w.WriteHeader(http.StatusConflict)
		bc.Logger.Warn("Froad")
		return
	}

	if isEmpty {
		bc.Storage.CreateOrder(r.Context(), cookieUserID, &userOrder)
		w.WriteHeader(http.StatusNoContent)
		bc.Logger.Info("No Content")
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bc.Logger.Error(err)
		return
	}

	responce := struct {
		Number string `json:"order"`
		Status string `json:"status"`
		Points int    `json:"points"`
	}{}
	responce.Number = DBOrder.Number
	responce.Status = DBOrder.Status
	responce.Points = DBOrder.Points

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(responce); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		bc.Logger.Error(err)
	}
}
