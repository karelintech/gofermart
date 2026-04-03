package storage

import (
	"context"
	"database/sql"
	"gofermart/internal/auth"
	"net/http"
	"time"

	// импорт для "database/sql"
	_ "github.com/lib/pq"
)

// Storage - БД
type Storage struct {
	*sql.DB
}

// UserModel - модель пользователя
type UserModel struct {
	ID        int    `json:"id"`
	Balance   int    `json:"balance"`
	Withdrawn int    `json:"withdrawn"`
	Login     string `json:"login"`
	Password  string `json:"password"`
}

// OrderModel - модель заказа
type OrderModel struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"user_idid,omitempty"`
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_atid,omitempty"`
	Points    int       `json:"points"`
}

// WithDrawModel - модель списания баллов
type WithDrawModel struct {
	OrderNumber string    `json:"order_number"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// RegisterUser - запрос в БД на регистрацию пользователя
func (s *Storage) RegisterUser(ctx context.Context, login, password string) (status, UserID int, err error) {
	tx, err := s.Begin()
	if err != nil {
		return http.StatusInternalServerError, 0, err
	}
	defer tx.Commit()

	hexLogin := auth.HashFunc(login)
	hexPass := auth.HashFunc(password)

	var usedLogin string
	err = tx.QueryRowContext(ctx, "SELECT login FROM users WHERE login=$1", hexLogin).Scan(&usedLogin)
	if err != nil {
		if err != sql.ErrNoRows {
			return http.StatusInternalServerError, 0, err
		}
	} else {
		return http.StatusConflict, 0, err
	}

	err = tx.QueryRowContext(ctx, "INSERT INTO users(login, password) VALUES($1, $2) RETURNING id;", hexLogin, hexPass).Scan(&UserID)
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, 0, err
	}

	return http.StatusOK, UserID, nil
}

// AuthUser - запрос на авторизацию
func (s *Storage) AuthUser(ctx context.Context, login, password string) (status int, userID int, err error) {
	var user UserModel
	err = s.QueryRowContext(ctx, "SELECT id,login,password FROM users WHERE login=$1 AND password=$2;", login, password).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return http.StatusUnauthorized, 0, err
		}

		return http.StatusInternalServerError, 0, err
	}

	if user.Login == login && user.Password == password {
		return 0, user.ID, nil
	}

	return http.StatusUnauthorized, 0, nil
}

// GetOrder - запрос на получение заказа (возвращает  в т.ч bool - наличие данного заказа в БД)
func (s *Storage) GetOrder(ctx context.Context, userOrder *OrderModel) (OrderModel, bool, error) {
	var order OrderModel
	tx, err := s.Begin()
	if err != nil {
		return order, false, err
	}

	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, "SELECT * FROM orders WHERE number=$1", userOrder.Number).
		Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.CreatedAt, &order.UpdatedAt, &order.Points)
	if err == sql.ErrNoRows {
		return order, true, nil
	}
	if err != nil {
		return order, false, err
	}

	tx.Commit()
	return order, false, nil
}

// CreateOrder - запрос на создание заказа
func (s *Storage) CreateOrder(ctx context.Context, userID int, userOrder *OrderModel) error {
	tx, err := s.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO orders(user_id,number,points) VALUES($1, $2, $3);", userID, userOrder.Number, userOrder.Points)
	if err != nil {
		tx.Rollback()
		return err
	}

	var points int
	err = tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id=$1", userID).Scan(&points)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Начисляем баллы пользователю за заказ
	points += userOrder.Points
	_, err = tx.ExecContext(ctx, "UPDATE users SET balance=$1 WHERE id=$2", points, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// GetAllOrders - запрос на получение списка всех заказов
func (s *Storage) GetAllOrders(ctx context.Context, userID int) ([]OrderModel, error) {
	orders := make([]OrderModel, 0)
	rows, err := s.QueryContext(ctx, "SELECT number, status, points, created_at FROM orders WHERE user_id=$1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order OrderModel
		err = rows.Scan(&order.Number, &order.Status, &order.Points, &order.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, err
	}
	return orders, nil
}

// GetUserBalance - запрос на получение баланса пользователя
func (s *Storage) GetUserBalance(ctx context.Context, userID int) (balance, withdrawn int, err error) {
	tx, err := s.Begin()
	if err != nil {
		return -1, -1, err
	}

	err = tx.QueryRowContext(ctx, "SELECT balance, withdrawn FROM users WHERE id=$1", userID).Scan(&balance, &withdrawn)
	if err != nil {
		tx.Rollback()
		return -1, -1, err
	}
	tx.Commit()
	return
}

// CreateWithDraw - выполнение списания средств
func (s *Storage) CreateWithDraw(ctx context.Context, userID int, orderNum string, sum int) error {
	tx, err := s.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO withdrawals (user_id, order_number, sum) VALUES($1,$2,$3);", userID, orderNum, sum)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE orders SET status='PROCESSED' where number=$1;", orderNum)
	if err != nil {
		tx.Rollback()
		return err
	}

	var newBalance int
	var newWithdrawn int

	err = tx.QueryRowContext(ctx, "SELECT balance, withdrawn FROM users where id=$1;", userID).Scan(&newBalance, &newWithdrawn)
	if err != nil {
		tx.Rollback()
		return err
	}

	newBalance -= sum
	newWithdrawn += sum
	_, err = tx.ExecContext(ctx, "UPDATE users SET balance=$1, withdrawn=$2 where id=$3;", newBalance, newWithdrawn, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// GetWithDraw - запрос на запись о списании средств по заказу
func (s *Storage) GetWithDraw(ctx context.Context, userID int) ([]WithDrawModel, error) {
	var models []WithDrawModel

	tx, err := s.Begin()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT  order_number, sum, processed_at from withdrawals WHERE user_id=$1 ORDER BY processed_at DESC;", userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for rows.Next() {
		var row WithDrawModel
		err := rows.Scan(&row.OrderNumber, &row.Sum, &row.ProcessedAt)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		models = append(models, row)
	}

	if rows.Err() != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return models, nil
}
