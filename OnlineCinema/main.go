package main

import (
	_ "github.com/lib/pq"
	"net/http"
	"techno-test_Films/OnlineCinema/config"
	"techno-test_Films/OnlineCinema/handlers/auth"
	"techno-test_Films/OnlineCinema/handlers/user"
	slogpretty "techno-test_Films/OnlineCinema/lib"
	storage2 "techno-test_Films/OnlineCinema/storage"
)

func main() {
	//загружаем конфиг
	cfg := config.MustLoad()

	//инициализируем логер
	logger := slogpretty.SetupLogger()
	logger.Info("Logger is start")

	db, err := storage2.New(cfg.DbStorage)
	if err != nil {
		logger.Error("Database service is not start: ", err.Error())
		return
	}

	//создаем все необходимые таблицы
	err = db.Init()
	if err != nil {
		logger.Error("Initialization database complete with error: ", err.Error())
	}
	logger.Info("Initialization database complete")

	//роут
	mux := http.NewServeMux()
	mux.HandleFunc("/", auth.NonPage)
	mux.HandleFunc("/GetAllUsers", auth.AdminAuth(users.GetAllUsers(db), db))
	mux.HandleFunc("/CreateUser", auth.AdminAuth(users.CreateUser(db), db))
	mux.HandleFunc("/DeleteUser", auth.AdminAuth(users.DeleteUser(db), db))
	mux.HandleFunc("/CreateActor", auth.AdminAuth(db.CreateActor, db))
	mux.HandleFunc("/UpdateActor", auth.AdminAuth(db.UpdateActor, db))
	mux.HandleFunc("/CreateFilm", auth.AdminAuth(db.CreateFilm, db))
	mux.HandleFunc("/UpdateFilm", auth.AdminAuth(db.UpdateFilm, db))
	mux.HandleFunc("/GetActors", auth.UserAuth(db.GetActors, db))
	mux.HandleFunc("/DeleteFilm", auth.AdminAuth(db.DeleteFilm, db))
	mux.HandleFunc("/DeleteActor", auth.AdminAuth(db.DeleteActor, db))
	mux.HandleFunc("/GetFilms", auth.UserAuth(db.GetFilms, db))
	//запуск сервера
	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      mux,
		ReadTimeout:  cfg.HttpServer.TimeoutRequest,
		WriteTimeout: cfg.HttpServer.IdleTimeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server does not started")
	}

}
