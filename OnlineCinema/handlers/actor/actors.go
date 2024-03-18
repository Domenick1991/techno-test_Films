package actor

import (
	"encoding/json"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"log/slog"
	"net/http"
	"strconv"
	storages "techno-test_Films/storage"
)

// @Summary Показать актёров
// @Tags actors
// @Description Возвращает всех актёров фильмотеки
// @id GetActors
// @Accept json
// @Procedure json
// @router /GetActors [get]
// @Success 200 {object} storages.OutActor
// @Security BasicAuth
func GetActors(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodGet {

			var actors []storages.OutActor
			err := storage.DB.Select().From("actors").All(&actors)
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
				return
			}
			//Заполняем фильмы
			for i, actor := range actors {
				querytext := "Select id, filmname, description, releasedate, rating from films as f left join filmsactors as fa on f.id = fa.film where fa.actor = " + strconv.Itoa(actor.Id)
				var films []storages.Film
				err = storage.DB.NewQuery(querytext).All(&films)
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных об актёрах")
					return
				}
				for _, film := range films {
					actors[i].Films = append(actors[i].Films, film)
				}

			}

			result, _ := json.MarshalIndent(actors, "", "\t")
			storages.HttpResponseObject(w, http.StatusOK, result)
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Добавить актёра
// @Tags actors
// @Description Функция добавляет нового актёра
// @id CreateActor
// @Accept json
// @Procedure json
// @router /CreateActor [Post]
// @param input body storages.Actor true "Информация об актёре"
// @Success 200 {object} storages.Actor
// @Security BasicAuth
func CreateActor(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var actor storages.Actor

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&actor)
			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}

			actorDB, errlist := actor.ConvertToDB()
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			err = storage.DB.Model(&actorDB).Insert()
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить актёра или актёр уже существует")
			} else {

				//Если передавалась информация о фильмах - добавляем
				actor.Id = actorDB.Id
				if actor.AddFilms != nil {
					err = storage.AddActorToFilms(actor.GetNewActorFilmInfo())
					if err != nil {
						storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
						return
					}
				}
				result, _ := json.MarshalIndent(actorDB, "", "\t")
				storages.HttpResponseObject(w, http.StatusOK, result)
			}

		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Изменить актёра
// @Tags actors
// @Description обновляет данные об актере
// @id UpdateActor
// @Accept json
// @Procedure json
// @router /UpdateActor [Post]
// @param input body storages.Actor true "Информация об актёре"
// @Success 200 {object} storages.Actor
// @Security BasicAuth
func UpdateActor(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var actor storages.Actor

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&actor)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			if actor.Id == 0 {
				storages.HttpResponse(w, http.StatusBadRequest, "Не указан id актёра")
				return
			}

			actorDB, errlist := actor.ConvertToDB()
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			params := actorDB.GetUpdatesData()
			if len(params) > 0 {
				_, err = storage.DB.Update(actorDB.TableName(), params, dbx.HashExp{"id": actorDB.Id}).Execute()
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по актёру или актёр не существует")
				}
			}

			//Если передавалась информация об исключении актёра из фильмов - обновляем
			actor.Id = actorDB.Id
			if actor.DeleteFilms != nil {
				err = storage.DeleteActorToFilms(actor.GetDeleteActorFilmInfo())
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации о фильмах "+err.Error())
					return
				}
			}

			//Если передавалась информация о новых фильмах - обновляем
			if actor.AddFilms != nil {
				err = storage.AddActorToFilms(actor.GetNewActorFilmInfo())
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при добавлении информации о фильмах "+err.Error())
					return
				}
			}
			storages.HttpResponse(w, http.StatusOK, "Информация об актёре изменена")
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Удалить актёра
// @Tags actors
// @Description Удаляет актёра
// @id DeleteActor
// @Accept json
// @Procedure json
// @router /DeleteActor [Delete]
// @param input body storages.Actor true "Информация об актёре"
// @Success 200
// @Security BasicAuth
func DeleteActor(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodDelete {
			var actor storages.Actor

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&actor)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}
			if actor.Id == 0 {
				storages.HttpResponse(w, http.StatusBadRequest, "Не указан id актёра")
				return
			}

			actorDB, errlist := actor.ConvertToDB()
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			err = storage.DB.Model(&actorDB).Delete()
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось обновить информацию по фильму или фильм не существует")
			}

			//удаляем информацию из актёров
			_, err = storage.DB.Delete("filmsactors", dbx.HashExp{"actor": actorDB.Id}).Execute()
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при удалении информации об актёрах "+err.Error())
				return
			}
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод DELETE")
		}
	}
}
