package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

var (
	opts   *Opts
	debug  bool
	jwtKey []byte
)

type Opts struct {
	Port         int
	Env          string
	HS256        string
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

	flag.StringVar(&o.Env, "env", lookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT", 9083), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", lookupEnv("HS256"), "HS256 key")
	flag.StringVar(&o.BackendToken, "token", lookupEnv("BACKEND_TOKEN"), "Backend Token")
	flag.StringVar(&o.CallbackUrl, "callback", lookupEnv("BACKEND_CALLBACK_URL"), "Callback URL")
	flag.StringVar(&o.GitBasePath, "git-base", lookupEnv("GIT_BASE", "/tmp"), "Git base storage path")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if o.Env == "local" || o.Env == "dev" {
		debug = true
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if o.HS256 != "" {
		h := sha256.New()
		h.Write([]byte(o.HS256))
		jwtKey = h.Sum(nil)
		log.Debugf("jwtKey: %v", jwtKey)
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
	}

	return o
}

func lookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			os.Setenv(key, v)
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
			os.Setenv(key, strconv.Itoa(v))
			return v
		}
	}
	return 0
}

func main() {
	//the .env should be loaded before showing the banner, as the banner shows also the ENVs
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}
	//this will set the default ENVs
	opts = NewOpts()

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		log.Printf("could not display banner...")
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/analyze", jwtAuth(jwtAuthServer(analyze))).Methods("POST")
	log.Println("Starting analysis on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
