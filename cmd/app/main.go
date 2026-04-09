package main

import (
	"flag"
	"gofermart/internal/handlers"
	"gofermart/internal/storage"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var addr string

func main() {
	addr := os.Getenv("RUN_ADDRESS")
	if addr == "" {
		flag.StringVar(&addr, "a", "localhost:8080", "Addres and a port")
		flag.Parse()
	}

	router := handlers.NewBaseController()
	router.Logger = logrus.New()

	file, err := os.OpenFile("data.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		router.Logger.Fatal(err)
	}
	defer file.Close()

	router.Logger.SetOutput(file)
	if err := godotenv.Load(); err != nil {
		router.Logger.Fatal(err)
	}

	DBURL := os.Getenv("DBURL")
	if err = storage.RunMigrations(DBURL, ""); err != nil {
		router.Logger.Fatal(err)
	}

	credentials := os.Getenv("credentials")
	err = router.DBInit(credentials)
	if err != nil {
		router.Logger.Fatal(err)
	}

	router.Logger.Info("Run ", addr)
	router.Logger.Fatal(http.ListenAndServe(addr, router.BuildHandlers()))
}
