package env

import "os"

type AppEnv string

const (
	AppEnvProduction  = "production"
	AppEnvDevelopment = "development"
)

func GetAppEnv() AppEnv {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return AppEnvProduction
	}

	return AppEnvDevelopment
}

func GetDBPath() string {
	env := os.Getenv("DATABASE_FILE_PATH")
	if env == "" {
		return "./data.db"
	}

	return env
}
