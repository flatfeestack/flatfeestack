package main

import (
	"log/slog"
	"os"
	"strconv"
)

func LookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			err := os.Setenv(key, v)
			if err != nil {
				slog.Debug("Could not set env variable", slog.String("key", key), slog.String("value", v), slog.Any("error", err))
				return ""
			}
			return v
		}
	}
	return ""
}

func LookupEnvInt(key string, defaultValues ...int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			slog.Debug("LookupEnvInt[%s]: %v", slog.String(key, key), slog.Any("error", err))
			return 0
		}
		return v
	}
	for _, v := range defaultValues {
		if v != 0 {
			err := os.Setenv(key, strconv.Itoa(v))
			if err != nil {
				slog.Debug("Could not set env variable", slog.String("key", key), slog.Int("value", v), slog.Any("error", err))
				return 0
			}
			return v
		}
	}
	return 0
}
