package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

var (
	opts *Opts
)

type Opts struct {
	Port         int
	BackendToken string
	CallbackUrl  string
	GitBasePath  string
}

func NewOpts() *Opts {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT", 9083), "listening HTTP port")
	flag.StringVar(&o.BackendToken, "token", lookupEnv("BACKEND_TOKEN",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbmFseXNpcy1lbmdpbmVAZmxhdGZlZXN0YWNrLmlvIn0.HJInRFNeQNTZhdQghG1Ylbng23wKxFQscJTLAkf8hu8"),
		"Backend Token")
	flag.StringVar(&o.CallbackUrl, "callback", lookupEnv("WEBHOOK_CALLBACK_URL", "http://backend:9082/hooks/analysis-engine"), "Callback URL")
	flag.StringVar(&o.GitBasePath, "git-base-path", lookupEnv("GO_GIT_BASE_PATH", "/tmp"), "Git base storage path")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return o
}

func lookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			return v
		}
	}
	return ""
}

func lookupEnvInt(key string, defaultValues ...int) int {
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
			return v
		}
	}
	return 0
}

func main() {
	opts = NewOpts()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/webhook", analyzeRepository).Methods("POST")
	log.Println("Starting api on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
