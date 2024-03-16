package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	DbStorage  string `yaml:"database_connection_string" end-required:"true"`
	HttpServer `yaml:"http_server"`
}

type HttpServer struct {
	Address        string        `yaml:"address" end-default:"localhost:8081"`
	TimeoutRequest time.Duration `yaml:"timeout_request" end-default:"4s"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" end-default:"60s"`
}

func MustLoad() *Config {
	//Путь до конфига берет из переменной окружения
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("get current path error: %s", err)
	}
	configPath := currentPath + `\OnlineCinema\config\config.yml`

	//Проверяем существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file is  does not exsits: %s", configPath)
	}

	//Создаем объект конфига
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %s , err %s", configPath, err)
	}

	return &cfg
}
