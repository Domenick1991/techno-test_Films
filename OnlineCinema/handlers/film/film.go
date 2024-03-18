package film

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	storages "techno-test_Films/storage"
)

// @Summary Показать фильмы
// @Tags films
// @Description Получает информацию о фильмах
// @id GetFilms
// @Accept json
// @Procedure json
// @router /GetFilms [Post]
// @param input body storages.GetFilmsParam true "условия выборки"
// @Success 200 {object} storages.OutFilm
// @Security BasicAuth
func GetFilms(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {

			var getFilmsParam storages.GetFilmsParam
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&getFilmsParam)
			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			if getFilmsParam.SortName != "" && strings.ToLower(getFilmsParam.SortName) != "rating" && strings.ToLower(getFilmsParam.SortName) != "filmname" && strings.ToLower(getFilmsParam.SortName) != "releasedate" {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверное название сортировки. Доступные значения  'rating','filmname' или 'releasedate'")
				return
			}
			if getFilmsParam.SortName == "" {
				getFilmsParam.SortName = "rating"
			}

			if getFilmsParam.SortType != "" && strings.ToLower(getFilmsParam.SortType) != "asc" && strings.ToLower(getFilmsParam.SortType) != "desc" {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный тип сортировки. Доступные значения  'asc' или 'desc'")
				return
			}
			if getFilmsParam.SortType == "" {
				getFilmsParam.SortType = "desc"
			}
			getFilmsParam.SortType = strings.ToUpper(getFilmsParam.SortType)
			getFilmsParam.FilmName = strings.ToLower(getFilmsParam.FilmName)
			getFilmsParam.ActorName = strings.ToLower(getFilmsParam.ActorName)

			var films []storages.OutFilm
			orderBy := getFilmsParam.SortName + " " + getFilmsParam.SortType
			whereText := "(select count(*) from filmsactors as fa left join actors as a on fa.actor = a.id where fa.film = f.id and lower(a.actorname) like '%" + getFilmsParam.ActorName + "%') > 0"
			err = storage.DB.Select().From("films as f").Where(dbx.Like("lower(filmname)", getFilmsParam.FilmName)).Where(dbx.NewExp(whereText)).OrderBy(orderBy).All(&films)
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных о фильмах")
				return
			}

			//Заполняем актёров
			for i, film := range films {
				querytext := "Select id, actorname, sex, birthdate from actors as a left join filmsactors as fa on a.id = fa.actor where fa.film = " + strconv.Itoa(film.Id)
				var actors []storages.Actor
				err = storage.DB.NewQuery(querytext).All(&actors)
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
					return
				}
				for _, actor := range actors {
					films[i].Actors = append(films[i].Actors, actor)
				}
			}

			result, _ := json.MarshalIndent(films, "", "\t")
			storages.HttpResponseObject(w, http.StatusOK, result)
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Добавить фильм
// @Tags films
// @Description добавляет информацию о Фильме
// @id CreateFilm
// @Accept json
// @Procedure json
// @router /CreateFilm [Post]
// @param input body storages.Film true "Информация о фильме"
// @Security BasicAuth
func CreateFilm(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var film storages.Film

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&film)
			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			filmDB, errlist := film.ConvertToDB(true)
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			//Проверяем существует ли фильм по имени и дате релиза
			var oldfilm storages.FilmDB
			err = storage.DB.Select().From("films").Where(dbx.HashExp{"filmname": filmDB.Filmname, "releasedate": filmDB.Releasedate}).One(&oldfilm)
			if err != nil { //если фильм не существует - создаем
				err = storage.DB.Model(&filmDB).Insert()
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить фильм или фильм уже существует")
				} else {

					//Если передавалась информация об актёрах - добавляем
					film.Id = filmDB.Id
					if film.AddActors != nil {
						err = storage.AddActorToFilms(film.GetActorFilmInfo())
						if err != nil {
							storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
							return
						}
					}
					storages.HttpResponse(w, http.StatusOK, fmt.Sprint("Фильм успешно добавлен, id: ", filmDB.Id))
				}
			} else {
				storages.HttpResponse(w, http.StatusInternalServerError, fmt.Sprint("Фильм уже внесен в базу с id: ", oldfilm.Id))
			}
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Обновить фильм
// @Tags films
// @Description обновляет информацию о Фильме
// @id UpdateFilm
// @Accept json
// @Procedure json
// @router /UpdateFilm [Post]
// @param input body storages.Film true "Информация о фильме"
// @Security BasicAuth
func UpdateFilm(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var film storages.Film

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&film)
			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			if film.Id == 0 {
				storages.HttpResponse(w, http.StatusBadRequest, "Не указан id фильма")
				return
			}

			filmDB, errlist := film.ConvertToDB(false)
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}
			params := filmDB.GetUpdatesData()
			if len(params) > 0 {
				_, err = storage.DB.Update(filmDB.TableName(), params, dbx.HashExp{"id": filmDB.Id}).Execute()
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
				}
			}

			//Если передавалась информация об исключении актёра из фильмов - обновляем
			filmDB.Id = film.Id
			if film.DeleteActors != nil {
				err = storage.DeleteActorToFilms(film.GetDeleteActorFilmInfo())
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
					return
				}
			}

			//Если передавалась информация о новых фильмах - обновляем
			if film.AddActors != nil {
				err = storage.AddActorToFilms(film.GetNewActorFilmInfo())
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации об актёрах "+err.Error())
					return
				}
			}
			storages.HttpResponse(w, http.StatusOK, "Информация о фильме изменена")
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary удалить фильм
// @Tags films
// @Description удаляет информацию о Фильме
// @id DeleteFilm
// @Accept json
// @Procedure json
// @router /DeleteFilm [Delete]
// @param input body storages.Film true "Информация о фильме"
// @Security BasicAuth
func DeleteFilm(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodDelete {
			var film storages.Film

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&film)
			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			if film.Id == 0 {
				storages.HttpResponse(w, http.StatusBadRequest, "Не указан id фильма")
				return
			}

			filmDB, errlist := film.ConvertToDB(false)
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			err = storage.DB.Model(&filmDB).Delete()
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
			}

			//удаляем информацию из актёров
			_, err = storage.DB.Delete("filmsactors", dbx.HashExp{"film": filmDB.Id}).Execute()
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
				return
			}
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод DELETE")
		}
	}
}
