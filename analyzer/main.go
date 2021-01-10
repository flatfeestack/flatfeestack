package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	opts *Opts
)

type Opts struct {
	Port int
}
func NewOpts() *Opts {
	o := &Opts{}
	flag.IntVar(&o.Port, "port", LookupEnvInt("PORT"), "listening HTTP port")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return o
}

func setDefaultInt(actualValue int, defaultValue int) int {
	if actualValue == 0 {
		return defaultValue
	}
	return actualValue
}

func LookupEnvInt(key string) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("LookupEnvOrInt[%s]: %v", key, err)
			return 0
		}
		return v
	}
	return 0
}

func defaultOpts(opts *Opts) {
	opts.Port = setDefaultInt(opts.Port, 8080)
}

func main() {
	opts = NewOpts()
	defaultOpts(opts)
	err := setEnvs()
	if err != nil {
		fmt.Println(err)
	}
	GClientWrapper = &GithubClientWrapperClient{
		GitHubURL: "https://api.github.com/graphql",
	}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/contributions", getAllContributions).Methods("GET")
	router.HandleFunc("/weights", getContributionWeights).Methods("GET")
	router.HandleFunc("/webhook", analyzeRepository).Methods("POST")
	log.Println("Starting api on port " + strconv.Itoa(opts.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), router))
}
