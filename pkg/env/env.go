package env

import (
	"os"
	"strconv"
)

const (
	OTEL_URL string = "OTEL_EXPORTER_URL"
)

func GetString(env string, def string) string {
	if e := os.Getenv(env); e != "" {
		return e
	}
	return def
}

func GetInt(env string, def int) int {
	i, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return def
	}
	return i
}
