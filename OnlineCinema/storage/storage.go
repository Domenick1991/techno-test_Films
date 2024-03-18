package storage

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Storage struct {
	DB *dbx.DB
}

type Actor struct {
	Id          int    `json:"id,omitempty"` //ИД Актёра
	Actorname   string `json:"actorname"`    //Имя актёра
	Sex         string `json:"sex"`          //Пол
	Birthdate   string `json:"birthdate"`    //Дата рождения
	AddFilms    []int  `json:"addfilms"`     //Массив из ИД фильмов, в которые нужно добавить Актёра. Не обязательный параметр
	DeleteFilms []int  `json:"DeleteFilms"`  //Массив из ИД фильмов, из которых нужно исключить Актёра. Не обязательный параметр
}

type OutActor struct {
	Id        int    `json:"id,omitempty"` //ИД Актёра
	Actorname string `json:"actorname"`    //Имя актёра
	Sex       string `json:"sex"`          //Пол
	Birthdate string `json:"birthdate"`    //Дата рождения
	Films     []Film `json:"films"`        //Массив с информацией о фильмах
}

type GetFilmsParam struct {
	SortType  string `json:"sorttype"`  //тип сортировки, значения: asc или desc
	SortName  string `json:"sortname"`  //имя сортировки, значения: rating, filmname, releasedate
	ActorName string `json:"actorname"` //Фильтр по имени актёра
	FilmName  string `json:"filmname"`  //фильтр по названию фильма
}

type ActorDB struct {
	Id        int
	Actorname string
	Sex       string
	Birthdate time.Time
}

func (actor *ActorDB) TableName() string {
	return "actors"
}

func (actor *Actor) ConvertToDB() (ActorDB, []ErrorList) {
	var errlist []ErrorList

	actordb := ActorDB{}
	actordb.Id = actor.Id

	if actor.Sex != "" {
		if strings.ToLower(actor.Sex) == "мужской" || strings.ToLower(actor.Sex) == "женский" {
			actordb.Sex = actor.Sex
		} else {
			errlist = append(errlist, ErrorList{"пол указывается в формате 'мужской' или 'женский'"})
		}
	}
	if actor.Actorname != "" {
		if utf8.RuneCountInString(actor.Actorname) > 200 {
			errlist = append(errlist, ErrorList{"имя актёра не должно превышать 200 символов"})
		} else {
			actordb.Actorname = actor.Actorname
		}
	}

	if actor.Birthdate != "" {
		date, err := time.Parse("2006-01-02", actor.Birthdate)
		if err != nil {
			errlist = append(errlist, ErrorList{"неверный формат даты. Используйте формат '01-12-2000' или '2000-12-01'"})
		} else {
			actordb.Birthdate = date
		}
	}

	if len(errlist) > 0 {
		return actordb, errlist
	} else {
		return actordb, nil
	}
}
func (actor *Actor) GetNewActorFilmInfo() []ActorAndFilms {
	var result []ActorAndFilms
	for _, filmId := range actor.AddFilms {
		result = append(result, ActorAndFilms{actor.Id, filmId})
	}
	return result
}
func (actor *Actor) GetDeleteActorFilmInfo() []ActorAndFilms {
	var result []ActorAndFilms
	for _, filmId := range actor.DeleteFilms {
		result = append(result, ActorAndFilms{actor.Id, filmId})
	}
	return result
}

// GetUpdatesData Функция возвращает Мар содержащую только измененные поля, которые необходимо записать в БД. Если поле не было передано для обновления, то и записываться в базу оно не будет
func (actor *ActorDB) GetUpdatesData() map[string]interface{} {
	var data = make(map[string]interface{})
	if actor.Actorname != "" {
		data["actorname"] = actor.Actorname
	}
	if actor.Sex != "" {
		data["sex"] = actor.Sex
	}
	if !actor.Birthdate.IsZero() {
		data["birthdate"] = actor.Birthdate
	}
	return data
}

type Film struct {
	Id           int    `json:"id"`           //ИД Фильма
	Filmname     string `json:"filmname"`     //Имя фильма
	Description  string `json:"description"`  //описание
	Releasedate  string `json:"releasedate"`  //дата выхода
	Rating       string `json:"rating"`       //рейтинг
	AddActors    []int  `json:"addActors"`    //Массив ИД Актёров, которых нужно добавить в фильм
	DeleteActors []int  `json:"deleteActors"` //Массив ИД Актёров, которых нужно удалить из фильма
}

type OutFilm struct {
	Id          int     `json:"id"`          //ИД Фильма
	Filmname    string  `json:"filmname"`    //Имя фильма
	Description string  `json:"description"` //описание
	Releasedate string  `json:"releasedate"` //дата выхода
	Rating      string  `json:"rating"`      //рейтинг
	Actors      []Actor `json:"actors"`      //Массив с информацией об Актёра в фильме
}

func (film *Film) GetActorFilmInfo() []ActorAndFilms {
	var result []ActorAndFilms
	for _, actorID := range film.AddActors {
		result = append(result, ActorAndFilms{actorID, film.Id})
	}
	return result
}

type FilmDB struct {
	Id          int
	Filmname    string
	Description string
	Releasedate time.Time
	Rating      int
}

type ActorAndFilms struct {
	Actor int
	Film  int
}

func (a *ActorAndFilms) TableName() string {
	return "filmsactors"
}

func (film *Film) ConvertToDB(isCreate bool) (FilmDB, []ErrorList) {
	var errlist []ErrorList

	filmDB := FilmDB{}
	filmDB.Id = film.Id

	if film.Filmname != "" || isCreate {
		if utf8.RuneCountInString(film.Filmname) > 1 && utf8.RuneCountInString(film.Filmname) < 150 {
			filmDB.Filmname = film.Filmname
		} else {
			errlist = append(errlist, ErrorList{"название фильма должно содержать от 1 до 150 символов"})
		}
	}

	if film.Description != "" {
		if utf8.RuneCountInString(film.Description) < 1000 {
			filmDB.Description = film.Description
		} else {
			errlist = append(errlist, ErrorList{"описание фильма не должно превышать 1000 символов"})
		}
	}

	if film.Rating != "" {
		result, err := strconv.Atoi(film.Rating)
		if err != nil || result < 0 || result > 10 {
			errlist = append(errlist, ErrorList{"рейтинг фильма должен быть от 0 до 10"})
			filmDB.Rating = -1
		}
		filmDB.Rating = result
	}
	if film.Releasedate != "" {
		date, err := time.Parse("2006-01-02", film.Releasedate)
		if err != nil {
			errlist = append(errlist, ErrorList{"неверный формат даты. Используйте формат '01-12-2000' или '2000-12-01'"})
		}
		filmDB.Releasedate = date
	}
	if len(errlist) > 0 {
		return filmDB, errlist
	} else {
		return filmDB, nil
	}
}
func (film *Film) GetNewActorFilmInfo() []ActorAndFilms {
	var result []ActorAndFilms
	for _, actorId := range film.AddActors {
		result = append(result, ActorAndFilms{actorId, film.Id})
	}
	return result
}
func (film *Film) GetDeleteActorFilmInfo() []ActorAndFilms {
	var result []ActorAndFilms
	for _, actorId := range film.DeleteActors {
		result = append(result, ActorAndFilms{actorId, film.Id})
	}
	return result
}

// GetUpdatesData Функция возвращает Мар содержащую только измененные поля, которые необходимо записать в БД. Если поле не было передано для обновления, то и записываться в базу оно не будет
func (film *FilmDB) GetUpdatesData() map[string]interface{} {
	var data = make(map[string]interface{})
	if film.Filmname != "" {
		data["filmname"] = film.Filmname
	}
	if film.Description != "" {
		data["description"] = film.Description
	}
	if film.Rating != -1 {
		data["rating"] = film.Rating
	}
	if !film.Releasedate.IsZero() {
		data["releasedate"] = film.Releasedate
	}
	return data
}

func (a *FilmDB) TableName() string {
	return "films"
}

type ErrorList struct {
	Error string
}

func HttpResponse(w http.ResponseWriter, status int, text string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	result, _ := json.MarshalIndent(text, "", "\t")
	w.Write([]byte(result))
}

func HttpResponseObject(w http.ResponseWriter, status int, text []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(text)
}

// New возвразает соединение с БД
func New(storagePath string) (*Storage, error) {

	db, err := dbx.MustOpen("postgres", storagePath)

	if err != nil {
		return nil, err
	}
	return &Storage{DB: db}, nil
}

// Init иницилизирует БД
func (storage *Storage) Init() error {
	//region Создаем таблицу Пользователей, индекс и пользователя администратора
	queryText := `CREATE TABLE IF NOT EXISTS users (
								id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
								userName varchar(20) NOT NULL,
								password varchar(20) NOT NULL,
								isAdmin boolean	 NOT NULL
								)`
	q := storage.DB.NewQuery(queryText)
	_, err := q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'Users' complete with error: %s", err.Error())
	}

	queryText = `CREATE UNIQUE INDEX IF NOT EXISTS UsersIndex on USERS (username, password)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index for table 'Users' complete with error: %s", err.Error())
	}

	//проверяем существует ли администратор, если нет - создаем.
	queryText = `Select count(*) as count from USERS where username ='admin'`
	q = storage.DB.NewQuery(queryText)
	resultRows, err := q.Rows()
	if err != nil {
		return fmt.Errorf("check Admin user complete with error: %s", err)
	}
	count := 0
	for resultRows.Next() {
		resultRows.Scan(&count)
	}

	if count == 0 {
		queryText = "INSERT INTO USERS(username, password, isAdmin) VALUES ( 'admin', " + "'" + EncodePassword("123") + "', true )"
		q = storage.DB.NewQuery(queryText)
		_, err = q.Execute()
		if err != nil {
			return fmt.Errorf("create Admin user complete with error: %s", err)
		}
	}

	//endregion

	//region TODO Создаем таблицу Актеров
	queryText = `CREATE TABLE IF NOT EXISTS actors (
								id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
								actorName varchar(200) NOT NULL,
								Sex varchar(7) NOT NULL,
								birthDate date NOT NULL
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'actors' complete with error: %s", err.Error())
	}

	queryText = `CREATE UNIQUE INDEX IF NOT EXISTS actorsindex on actors (actorName)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index 'actorsindex' for table 'actors' complete with error: %s", err.Error())
	}
	//endregion

	//region TODO Создаем таблицу Фильмов
	queryText = `CREATE TABLE IF NOT EXISTS films (
								id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
								filmName varchar(150) NOT NULL,
								description varchar(1000) NOT NULL,
								releaseDate date NOT NULL,
    							rating smallint
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'films' complete with error: %s", err.Error())
	}

	queryText = `CREATE INDEX IF NOT EXISTS filmsindex on films (filmName)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index 'filmsindex' for table 'films' complete with error: %s", err.Error())
	}

	queryText = `CREATE INDEX IF NOT EXISTS ratingindex on films (rating)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index 'rating 'for table 'films' complete with error: %s", err.Error())
	}
	//endregion

	//region TODO Создаем таблицу Фильмотека
	queryText = `CREATE TABLE IF NOT EXISTS filmsactors (
								film integer NOT NULL,
								actor integer NOT NULL,
                                       PRIMARY KEY (film,actor)
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'filmsactors' complete with error: %s", err.Error())
	}

	queryText = `CREATE INDEX IF NOT EXISTS filmindex on filmsactors (film)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index 'filmindex' for table 'filmsactors' complete with error: %s", err.Error())
	}

	queryText = `CREATE INDEX IF NOT EXISTS actorsindex on filmsactors (actor)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create Index 'actorsindex' for table 'filmsactors' complete with error: %s", err.Error())
	}
	//endregion

	return nil
}

func (storage *Storage) AddActorToFilms(datas []ActorAndFilms) error {

	for _, data := range datas {
		//Проверяем, что такие данные уже есть
		var oldactor ActorAndFilms
		err := storage.DB.Select().From(oldactor.TableName()).Where(dbx.HashExp{"film": data.Film, "actor": data.Actor}).One(&oldactor)
		if err != nil {
			err = storage.DB.Model(&data).Insert()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (storage *Storage) DeleteActorToFilms(datas []ActorAndFilms) error {

	for _, data := range datas {
		_, err := storage.DB.Delete("filmsactors", dbx.HashExp{"actor": data.Actor, "film": data.Film}).Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

// EncodePassword функция возвращает хэш для пароля
func EncodePassword(password string) string {
	// Для пэт проекта сделаем просто, добавим в конец к паролю @1 и сдвинем каждый символы на 1
	passwordNew := password + "@1"
	bs := []byte(passwordNew)
	for i := range bs {
		bs[i] = bs[i] + 1
	}
	return string(bs)
}

// DecodePassword Функция возвращает пароль по хэшу
func DecodePassword(password string) string {
	bs := []byte(password)
	for i := range bs {
		bs[i] = bs[i] - 1
	}
	pass := string(bs)
	return pass[0:strings.LastIndex(pass, "@1")]
}

// TODO Тело запроса
func RequestTolog(r *http.Request, logger *slog.Logger) {
	username, _, ok := r.BasicAuth()
	if !ok {
		username = ""
	}
	logger.Debug(
		"incoming request",
		"url", r.URL.Path,
		"method", r.Method,
		"user", username,
		"time", time.Now().Format("02-01-2006 15:04:05"),
	)
}
