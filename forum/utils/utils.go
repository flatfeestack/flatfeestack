package utils

import (
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func LookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			err := os.Setenv(key, v)
			if err != nil {
				log.Error("Could not set env variable", key, v, err)
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
			log.Printf("LookupEnvInt[%s]: %v", key, err)
			return 0
		}
		return v
	}
	for _, v := range defaultValues {
		if v != 0 {
			err := os.Setenv(key, strconv.Itoa(v))
			if err != nil {
				log.Error("Could not set env variable", key, v, err)
				return 0
			}
			return v
		}
	}
	return 0
}

func WriteErrorf(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	_, err := w.Write([]byte(`{"error":"` + msg + `"}`))
	if err != nil {
		log.Error("Could not write error", err)
	}
}

type KeyedMutex struct {
	mutexes sync.Map // Zero value is empty and ready for use
}

func (m *KeyedMutex) Lock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}

func EmptyOptions() *middleware.Options {
	// To deactivate authentication in the validator, set the AuthenticationFunc to NoopAuthenticationFunc
	// Validation is done
	options := &openapi3filter.Options{
		AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
	}

	options2 := &middleware.Options{
		Options: *options,
	}
	return options2
}
