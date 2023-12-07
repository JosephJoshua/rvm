package env

import (
	"encoding/base64"
	"fmt"
	"os"
)

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

func GetFirebaseCredentialsJSON() ([]byte, error) {
	env := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if env == "" {
		return nil, fmt.Errorf("GetFirebaseCredentialsJSON(): FIREBASE_CREDENTIALS_JSON not set")
	}

	decoded, err := base64.StdEncoding.DecodeString(env)
	if err != nil {
		return nil, fmt.Errorf("GetFirebaseCredentialsJSON(): failed to decode base64: %w", err)
	}

	return decoded, nil
}
