package utils

import (
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
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
