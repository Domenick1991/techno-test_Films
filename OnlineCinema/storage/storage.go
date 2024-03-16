package storage

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
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
	Id          int    `json:"id,omitempty"`
	Actorname   string `json:"actorname"`
	Sex         string `json:"sex"`
	Birthdate   string `json:"birthdate"`
	AddFilms    []int  `json:"addfilms"`
	DeleteFilms []int  `json:"DeleteFilms"`
}

type OutActor struct {
	Id        int    `json:"id,omitempty"`
	Actorname string `json:"actorname"`
	Sex       string `json:"sex"`
	Birthdate string `json:"birthdate"`
	Films     []Film `json:"films"`
}

type GetFilmsParam struct {
	SortType  string `json:"sorttype"`
	SortName  string `json:"sortname"`
	ActorName string `json:"actorname"`
	FilmName  string `json:"filmname"`
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

func (actor *Actor) convertToDB() (ActorDB, []ErrorList) {
	var errlist []ErrorList

	actordb := ActorDB{}
	actordb.Id = actor.Id

	if actor.Sex != "" {
		if strings.ToLower(actor.Sex) != "мужской" || strings.ToLower(actor.Sex) != "женский" {
			errlist = append(errlist, ErrorList{"пол указывается в формате 'мужской' или 'женский'"})
		} else {
			actordb.Sex = actor.Sex
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
	Id           int    `json:"id"`
	Filmname     string `json:"filmname"`
	Description  string `json:"description"`
	Releasedate  string `json:"releasedate"`
	Rating       string `json:"rating"`
	AddActors    []int  `json:"addActors"`
	DeleteActors []int  `json:"deleteActors"`
}

type OutFilm struct {
	Id          int     `json:"id"`
	Filmname    string  `json:"filmname"`
	Description string  `json:"description"`
	Releasedate string  `json:"releasedate"`
	Rating      string  `json:"rating"`
	Actors      []Actor `json:"actors"`
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

func (film *Film) convertToDB(isCreate bool) (FilmDB, []ErrorList) {
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
	w.WriteHeader(status)
	result, _ := json.MarshalIndent(text, "", "\t")
	w.Write([]byte(result))
}

func HttpResponseObject(w http.ResponseWriter, status int, text []byte) {
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

// GetActors Функция добавляет информацию о Фильме
func (storage *Storage) GetActors(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		var actors []OutActor
		err := storage.DB.Select().From("actors").All(&actors)
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
			return
		}
		//Заполняем фильмы

		for i, actor := range actors {
			querytext := "Select id, filmname, description, releasedate, rating from films as f left join filmsactors as fa on f.id = fa.film where fa.actor = " + strconv.Itoa(actor.Id)
			var films []Film
			err = storage.DB.NewQuery(querytext).All(&films)
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
				return
			}
			for _, film := range films {
				actors[i].Films = append(actors[i].Films, film)
			}

		}

		result, _ := json.MarshalIndent(actors, "", "\t")
		HttpResponseObject(w, http.StatusInternalServerError, result)
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// GetFilms Функция добавляет информацию о Фильме
func (storage *Storage) GetFilms(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var getFilmsParam GetFilmsParam
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&getFilmsParam)
		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		if getFilmsParam.SortName != "" && strings.ToLower(getFilmsParam.SortName) != "rating" && strings.ToLower(getFilmsParam.SortName) != "filmname" && strings.ToLower(getFilmsParam.SortName) != "releasedate" {
			HttpResponse(w, http.StatusBadRequest, "Неверное название сортировки. Доступные значения  'rating','filmname' или 'releasedate'")
			return
		}
		if getFilmsParam.SortName == "" {
			getFilmsParam.SortName = "rating"
		}

		if getFilmsParam.SortType != "" && strings.ToLower(getFilmsParam.SortType) != "asc" && strings.ToLower(getFilmsParam.SortType) != "desc" {
			HttpResponse(w, http.StatusBadRequest, "Неверный тип сортировки. Доступные значения  'asc' или 'desc'")
			return
		}
		if getFilmsParam.SortType == "" {
			getFilmsParam.SortType = "desc"
		}
		getFilmsParam.SortType = strings.ToUpper(getFilmsParam.SortType)
		getFilmsParam.FilmName = strings.ToLower(getFilmsParam.FilmName)
		getFilmsParam.ActorName = strings.ToLower(getFilmsParam.ActorName)

		var films []OutFilm
		orderBy := getFilmsParam.SortName + " " + getFilmsParam.SortType
		whereText := "(select count(*) from filmsactors as fa left join actors as a on fa.actor = a.id where fa.film = f.id and lower(a.actorname) like '%" + getFilmsParam.ActorName + "%') > 0"
		err = storage.DB.Select().From("films as f").Where(dbx.Like("lower(filmname)", getFilmsParam.FilmName)).Where(dbx.NewExp(whereText)).OrderBy(orderBy).All(&films)
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных о фильмах")
			return
		}

		//Заполняем актёров
		for i, film := range films {
			querytext := "Select id, actorname, sex, birthdate from actors as a left join filmsactors as fa on a.id = fa.actor where fa.film = " + strconv.Itoa(film.Id)
			var actors []Actor
			err = storage.DB.NewQuery(querytext).All(&actors)
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
				return
			}
			for _, actor := range actors {
				films[i].Actors = append(films[i].Actors, actor)
			}
		}

		result, _ := json.MarshalIndent(films, "", "\t")
		HttpResponseObject(w, http.StatusOK, result)
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// CreateFilm Функция добавляет информацию о Фильме
func (storage *Storage) CreateFilm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var film Film

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&film)
		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		filmDB, errlist := film.convertToDB(true)
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}

		//Проверяем существует ли фильм по имени и дате релиза
		var oldfilm FilmDB
		err = storage.DB.Select().From("films").Where(dbx.HashExp{"filmname": filmDB.Filmname, "releasedate": filmDB.Releasedate}).One(&oldfilm)
		if err != nil { //если фильм не существует - создаем
			err = storage.DB.Model(&filmDB).Insert()
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить фильм или фильм уже существует")
			} else {

				//Если передавалась информация об актёрах - добавляем
				film.Id = filmDB.Id
				if film.AddActors != nil {
					err = storage.AddActorToFilms(film.GetActorFilmInfo())
					if err != nil {
						HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
						return
					}
				}
				HttpResponse(w, http.StatusOK, fmt.Sprint("Фильм успешно добавлен, id: ", filmDB.Id))
			}
		} else {
			HttpResponse(w, http.StatusInternalServerError, fmt.Sprint("Фильм уже внесен в базу с id: ", oldfilm.Id))
		}
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// UpdateFilm Функция обновляет информацию о Фильме
func (storage *Storage) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var film Film

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&film)
		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		if film.Id == 0 {
			HttpResponse(w, http.StatusBadRequest, "Не указан id фильма")
			return
		}

		filmDB, errlist := film.convertToDB(false)
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}
		params := filmDB.GetUpdatesData()
		if len(params) > 0 {
			_, err = storage.DB.Update(filmDB.TableName(), params, dbx.HashExp{"id": filmDB.Id}).Execute()
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
			}
		}

		//Если передавалась информация об исключении актёра из фильмов - обновляем
		filmDB.Id = film.Id
		if film.DeleteActors != nil {
			err = storage.DeleteActorToFilms(film.GetDeleteActorFilmInfo())
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
				return
			}
		}

		//Если передавалась информация о новых фильмах - обновляем
		if film.AddActors != nil {
			err = storage.AddActorToFilms(film.GetNewActorFilmInfo())
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации об актёрах "+err.Error())
				return
			}
		}
		HttpResponse(w, http.StatusOK, "Информация о фильме изменена")
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// DeleteFilm Функция удаляет информацию о Фильме
func (storage *Storage) DeleteFilm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		var film Film

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&film)
		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		if film.Id == 0 {
			HttpResponse(w, http.StatusBadRequest, "Не указан id фильма")
			return
		}

		filmDB, errlist := film.convertToDB(false)
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}

		err = storage.DB.Model(&filmDB).Delete()
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
		}

		//удаляем информацию из актёров
		_, err = storage.DB.Delete("filmsactors", dbx.HashExp{"film": filmDB.Id}).Execute()
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
			return
		}
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод DELETE")
	}
}

// CreateActor Функция добавляет данные об актёре
func (storage *Storage) CreateActor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var actor Actor

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&actor)
		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}

		actorDB, errlist := actor.convertToDB()
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}

		err = storage.DB.Model(&actorDB).Insert()
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить актёра или актёр уже существует")
		} else {

			//Если передавалась информация о фильмах - добавляем
			actor.Id = actorDB.Id
			if actor.AddFilms != nil {
				err = storage.AddActorToFilms(actor.GetNewActorFilmInfo())
				if err != nil {
					HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
					return
				}
			}
			result, _ := json.MarshalIndent(actorDB, "", "\t")
			HttpResponseObject(w, http.StatusOK, result)
		}

	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// UpdateActor Функция обновляет данные об актере
func (storage *Storage) UpdateActor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var actor Actor

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&actor)

		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		if actor.Id == 0 {
			HttpResponse(w, http.StatusBadRequest, "Не указан id актёра")
			return
		}

		actorDB, errlist := actor.convertToDB()
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}

		params := actorDB.GetUpdatesData()
		if len(params) > 0 {
			_, err = storage.DB.Update(actorDB.TableName(), params, dbx.HashExp{"id": actorDB.Id}).Execute()
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по актёру или актёр не существует")
			}
		}

		//Если передавалась информация об исключении актёра из фильмов - обновляем
		actor.Id = actorDB.Id
		if actor.DeleteFilms != nil {
			err = storage.DeleteActorToFilms(actor.GetDeleteActorFilmInfo())
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации о фильмах "+err.Error())
				return
			}
		}

		//Если передавалась информация о новых фильмах - обновляем
		if actor.AddFilms != nil {
			err = storage.AddActorToFilms(actor.GetNewActorFilmInfo())
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
				return
			}
		}
		HttpResponse(w, http.StatusOK, "Информация об актёре изменена")
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
	}
}

// DeleteActor Функция удаляет данные об актере
func (storage *Storage) DeleteActor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		var actor Actor

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&actor)

		if err != nil {
			HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
			return
		}
		if actor.Id == 0 {
			HttpResponse(w, http.StatusBadRequest, "Не указан id актёра")
			return
		}

		actorDB, errlist := actor.convertToDB()
		if len(errlist) > 0 {
			result, _ := json.MarshalIndent(errlist, "", "\t")
			HttpResponseObject(w, http.StatusInternalServerError, result)
			return
		}

		err = storage.DB.Model(&actorDB).Delete()
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
		}

		//удаляем информацию из актёров
		_, err = storage.DB.Delete("filmsactors", dbx.HashExp{"actor": actorDB.Id}).Execute()
		if err != nil {
			HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
			return
		}
	} else {
		HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод DELETE")
	}
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
