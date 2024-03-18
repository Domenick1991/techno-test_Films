package main

import (
	_ "/techno-test_Films/docs"
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
	"net/http"
	"techno-test_Films/config"
	"techno-test_Films/handlers/actor"
	"techno-test_Films/handlers/auth"
	"techno-test_Films/handlers/film"
	users "techno-test_Films/handlers/user"
	slogpretty "techno-test_Films/lib"
	storage2 "techno-test_Films/storage"
)

// @title Фильмотека API
// @version 1.0
// @description Фильмотека
// @host localhost:8080
// @securitydefinitions.basic BasicAuth
// @in header
// @name Authorization
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
	mux.HandleFunc("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))
	mux.HandleFunc("/GetAllUsers", auth.AdminAuth(users.GetAllUsers(db, logger), db))
	mux.HandleFunc("/CreateUser", auth.AdminAuth(users.CreateUser(db, logger), db))
	mux.HandleFunc("/DeleteUser", auth.AdminAuth(users.DeleteUser(db, logger), db))
	mux.HandleFunc("/CreateActor", auth.AdminAuth(actor.CreateActor(db, logger), db))
	mux.HandleFunc("/UpdateActor", auth.AdminAuth(actor.UpdateActor(db, logger), db))
	mux.HandleFunc("/CreateFilm", auth.AdminAuth(film.CreateFilm(db, logger), db))
	mux.HandleFunc("/UpdateFilm", auth.AdminAuth(film.UpdateFilm(db, logger), db))
	mux.HandleFunc("/GetActors", auth.UserAuth(actor.GetActors(db, logger), db))
	mux.HandleFunc("/DeleteFilm", auth.AdminAuth(film.DeleteFilm(db, logger), db))
	mux.HandleFunc("/DeleteActor", auth.AdminAuth(actor.DeleteActor(db, logger), db))
	mux.HandleFunc("/GetFilms", auth.UserAuth(film.GetFilms(db, logger), db))
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
