package handlers

import (
	"database/sql"
	"gofermart/internal/auth"
	"gofermart/internal/storage"
	"io"

	// Используем pgx
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
)

var endpoint string = "/api/user"

type logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	SetOutput(io.Writer)
}

// BaseController - роутер с логгером и БД
type BaseController struct {
	Logger  logger
	Storage storage.Storage
}

// NewBaseController - ссылка на новый роутер
func NewBaseController() *BaseController {
	return &BaseController{}
}

// DBInit - инициализация БД
func (bc *BaseController) DBInit(credentials string) error {
	db, err := sql.Open("pgx", credentials)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	bc.Storage.DB = db
	return nil
}

// BuildHandlers - все хандлеры роутера
func (bc *BaseController) BuildHandlers() *chi.Mux {
	router := chi.NewRouter()
	router.Route(endpoint, func(r chi.Router) {
		r.Post("/register", bc.RegisterUser)
		r.Post("/login", bc.LoginUser)
	})

	router.Group(func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Post(endpoint+"/orders", bc.CreateOrder)
		r.Get(endpoint+"/orders", bc.GetOrders)
		r.Get(endpoint+"/orders/{number}", bc.GetOrderInfo)

		r.Get(endpoint+"/balance", bc.GetBalance)
		r.Post(endpoint+"/balance/withdraw", bc.CreateWithdraw)
		r.Get(endpoint+"/withdrawals", bc.GetWithdraw)
	})

	return router
}
