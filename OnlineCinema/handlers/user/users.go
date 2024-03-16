package users

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"net/http"
	storages "techno-test_Films/OnlineCinema/storage"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Isadmin  bool   `json:"userIsAdmin"`
}

func (u *User) TableName() string {
	return "users"
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

// GetAllUsers Функция возвращает всех пользователей БД
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

// CreateUser функция добавляет нового пользователя в БД
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
					HttpResponse(w, http.StatusInternalServerError, "Пользователь успешно добавлен")
				}
			} else {
				HttpResponse(w, http.StatusInternalServerError, "Пользователь уже существует")
			}
		} else {
			HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// DeleteUser функция удаляет пользователя из БД
func DeleteUser(storage *storages.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO сделать проверку что хотя бы один администратор должен оставаться.
		if r.Method == http.MethodDelete {
			type deleteUser struct {
				Id int `json:"id"`
			}
			var user deleteUser
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
