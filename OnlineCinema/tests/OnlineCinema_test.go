package tests

import (
	"github.com/gavv/httpexpect/v2"
	_ "github.com/stretchr/testify/require"
	"net/url"
	storages "techno-test_Films/OnlineCinema/storage"
	"testing"
)

const (
	host = "localhost:8080"
)

func TestAcotor_CreateActor(t *testing.T) {
	testCases := []struct {
		name      string
		Actorname string
		Sex       string
		Birthdate string
		error     string
		status    int
	}{
		{
			name:      "Успешный тест",
			Actorname: "Иванов Иван 66",
			Sex:       "Мужской",
			Birthdate: "1991-07-02",
			status:    200,
		},
		{
			name:      "Актёр существует",
			Actorname: "Иванов Иван Иванович7",
			Sex:       "Мужской",
			Birthdate: "1991-07-02",
			error:     "Не удалось добавить актёра или актёр уже существует",
			status:    500,
		},
		/*{
			name:      "Неверный формат данных",
			Actorname: "0",
			Sex:       "Муж",
			Birthdate: "19913-07-02",
		},*/
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			obj := e.POST("/CreateActor").
				WithJSON(storages.Actor{
					Actorname: test.Actorname,
					Sex:       test.Sex,
					Birthdate: test.Birthdate,
				}).
				WithBasicAuth("admin", "123").
				Expect().
				Status(test.status).
				JSON().Object()

			if test.error == "" {
				obj.Keys().ContainsOnly("Id", "Actorname", "Sex", "Birthdate")
				//id := obj.Value("id")
			} else {
				obj.Keys().ContainsOnly(test.error)
			}

		})
	}
}
