package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string
	DatabaseUrl string
	JwtKey string
}

func MustLoad() Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == ""{
		panic("PORT is required")
	}

	env := os.Getenv("ENV")
	if env == ""{
		panic("ENV is required")
	}

	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == ""{
		panic("DATABASE_URL is required")
	}

	jwtkey := os.Getenv("JWT_KEY")

	if jwtkey == ""{
		panic("JWT_KEY is required")
	}

	return Config{
		Port: port,
		Env: env,
		DatabaseUrl: dbUrl,
		JwtKey: jwtkey,
	}

}