package test

import "os"

func GetAPIBaseUrl() string {
	return os.Getenv("APP_BASE_URL")
}
