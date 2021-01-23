package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}

	o := &Opts{}
	flag.IntVar(&o.Port, "port", lookupEnvInt("PORT", 9083), "listening HTTP port")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	return o
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
