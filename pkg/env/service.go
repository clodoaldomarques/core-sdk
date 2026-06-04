package env

import (
	"os"
	"strconv"
)

const (
	OTEL_EXPORTER_ENDPOINT string = "OTEL_EXPORTER_OTLP_ENDPOINT"
	OTEL_SERVICE_NAME      string = "OTEL_SERVICE_NAME"
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
