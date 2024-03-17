package main

import (
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
	"net/http"
	"techno-test_Films/OnlineCinema/config"
	_ "techno-test_Films/OnlineCinema/docs"
	"techno-test_Films/OnlineCinema/handlers/actor"
	"techno-test_Films/OnlineCinema/handlers/auth"
	"techno-test_Films/OnlineCinema/handlers/film"
	"techno-test_Films/OnlineCinema/handlers/user"
	slogpretty "techno-test_Films/OnlineCinema/lib"
	storage2 "techno-test_Films/OnlineCinema/storage"
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
	mux.HandleFunc("/GetAllUsers", auth.AdminAuth(users.GetAllUsers(db), db))
	mux.HandleFunc("/CreateUser", auth.AdminAuth(users.CreateUser(db), db))
	mux.HandleFunc("/DeleteUser", auth.AdminAuth(users.DeleteUser(db), db))
	mux.HandleFunc("/CreateActor", auth.AdminAuth(actor.CreateActor(db), db))
	mux.HandleFunc("/UpdateActor", auth.AdminAuth(actor.UpdateActor(db), db))
	mux.HandleFunc("/CreateFilm", auth.AdminAuth(film.GetFilms(db), db))
	mux.HandleFunc("/UpdateFilm", auth.AdminAuth(film.UpdateFilm(db), db))
	mux.HandleFunc("/GetActors", auth.UserAuth(actor.GetActors(db), db))
	mux.HandleFunc("/DeleteFilm", auth.AdminAuth(film.DeleteFilm(db), db))
	mux.HandleFunc("/DeleteActor", auth.AdminAuth(actor.DeleteActor(db), db))
	mux.HandleFunc("/GetFilms", auth.UserAuth(film.GetFilms(db), db))
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
