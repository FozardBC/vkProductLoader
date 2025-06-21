package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Log        string `env:"LOG_MODE" env-default:"debug"`
	ServerHost string `env:"SRV_HOST"`
	ServerPort string `env:"SRV_PORT" env-default:"8080"`
	VkToken    string `env:"VK_TOKEN"`
	VkGroupID  int    `env:"VK_GROUP_ID"`
	DbPath     string `env:"DB_PATH" env-default:"../../storage/db.sqlite3"`
}

func MustRead() *Config {

	if err := godotenv.Load(); err != nil { // DEBUG:
		log.Print("INFO: file .env is not exists. Loading env variables ")
	}

	cfg := Config{}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		help, _ := cleanenv.GetDescription(cfg, nil)
		log.Print(help)
		log.Fatal(err)
	}

	return &cfg
}
