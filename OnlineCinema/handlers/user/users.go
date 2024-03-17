package users

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"net/http"
	storages "techno-test_Films/OnlineCinema/storage"
)

// User model info
// @Description User информация о пользователе
type User struct {
	Id       int    `json:"id"`          // идентификатор пользователя
	Username string `json:"username"`    // имя пользователя
	Password string `json:"password"`    // пароль пользователя
	Isadmin  bool   `json:"userIsAdmin"` // признак того, что пользователь является администратором
}

func (u *User) TableName() string {
	return "users"
}

type DeleteUserStruct struct {
	Id int `json:"id"`
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

// GetUser Функция возвращает пользователя по логину и паролю
func GetUser(userName, password string, storage *storages.Storage) (User, error) {
	var user User
	hashPass := storages.EncodePassword(password)
	err := storage.DB.Select().From("users").Where(dbx.HashExp{"username": userName, "password": hashPass}).One(&user)
	if err != nil {
		return user, fmt.Errorf("select script 'getUsers' complete with error: %s", err.Error())
	}
	return user, nil
}

// @Summary получить пользователей
// @Tags user
// @Description Возвращает всех пользователей приложения
// @id GetAllUsers
// @Accept json
// @Procedure json
// @router /GetAllUsers [get]
// @Success 200 {string} string "ok"
// @Security BasicAuth
func GetAllUsers(storage *storages.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var users []User
			//TODO переделать
			q := storage.DB.Select("id", "username", "isadmin").From("users")
			q.All(&users)

			result, _ := json.MarshalIndent(users, "", "\t")
			if users != nil {
				HttpResponseObject(w, http.StatusOK, result)
			} else {
				HttpResponse(w, http.StatusOK, "Нет пользователей")
			}
		} else {
			HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод GET")
		}
	}
}

// @Summary Создать пользователя
// @Tags user
// @Description Создает нового пользователя приложения
// @id CreateUser
// @Accept json
// @Procedure json
// @param input body User true "Информация о пользователе"
// @router /CreateUser [post]
// @Security BasicAuth
func CreateUser(storage *storages.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var user User

			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			err := decoder.Decode(&user)
			if err != nil {
				HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса")
				return
			}

			//хешируем пароль
			user.Password = storages.EncodePassword(user.Password)

			//Проверяем существует ли такой пользователь по имени
			var oldUser User
			err = storage.DB.Select().From(user.TableName()).Where(dbx.HashExp{"username": user.Username}).One(&oldUser)
			if err != nil { //ошибка в случае dbx возвращается если значения не нашлись
				err1 := storage.DB.Model(&user).Insert()
				if err1 != nil {
					HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить пользователя")
				} else {
					HttpResponse(w, http.StatusOK, "Пользователь успешно добавлен")
				}
			} else {
				HttpResponse(w, http.StatusInternalServerError, "Пользователь уже существует")
			}
		} else {
			HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Удалить пользователя
// @Tags user
// @Description Удаляет пользователя приложения
// @id DeleteUser
// @Accept json
// @Procedure json
// @param input body DeleteUserStruct true "Идентификатор пользователя"
// @router /DeleteUser [Delete]
// @Security BasicAuth
func DeleteUser(storage *storages.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO сделать проверку что хотя бы один администратор должен оставаться.
		if r.Method == http.MethodDelete {
			var user DeleteUserStruct
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			err := decoder.Decode(&user)
			if err != nil {
				HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса")
				return
			}

			_, err = storage.DB.Delete("users", dbx.HashExp{"id": user.Id}).Execute()
			if err != nil {
				HttpResponse(w, http.StatusInternalServerError, "Ошибка удаления пользователя")
			} else {
				HttpResponse(w, http.StatusOK, "Пользователь успешно удален")
			}
		} else {
			HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод DELETE")
		}
	}
}
