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
	Address        string        `yaml:"address" end-default:"localhost:8080"`
	TimeoutRequest time.Duration `yaml:"timeout_request" end-default:"4s"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" end-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = `config.yml`
	}

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
