package main

import (
	"crypto/sha256"
	"encoding/base32"
	"flag"
	"fmt"
	"forum/api"
	database "forum/db"
	"forum/globals"
	"forum/jwt"
	"forum/types"
	"forum/utils"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func NewOpts() *types.Opts {
	o := &types.Opts{}
	flag.StringVar(&o.Env, "env", utils.LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&o.Port, "port", utils.LookupEnvInt("PORT", 9086), "listening HTTP port")
	flag.StringVar(&o.HS256, "hs256", utils.LookupEnv("HS256", "test-seed"), "HS256 key")
	flag.StringVar(&o.DBPath, "db-path", utils.LookupEnv("DB_PATH",
		"postgresql://postgres:password@localhost:5432/flatfeestack?sslmode=disable"), "DB path")
	flag.StringVar(&o.DBDriver, "db-driver", utils.LookupEnv("DB_DRIVER",
		"postgres"), "DB driver")
	flag.StringVar(&o.DBScripts, "db-scripts", utils.LookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&o.Admins, "admins", utils.LookupEnv("ADMINS"), "Admins")

	//set defaults, be explicit
	if o.Env == "local" || o.Env == "dev" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if o.HS256 != "" {
		var err error
		globals.JwtKey, err = base32.StdEncoding.DecodeString(o.HS256)
		if err != nil {
			h := sha256.New()
			h.Write([]byte(o.HS256))
			globals.JwtKey = h.Sum(nil)
			log.Debugf("jwtKey: %v", globals.JwtKey)
		}
	} else {
		log.Fatalf("HS256 seed is required, non was provided")
	}

	globals.ADMINS = strings.Split(o.Admins, ";")

	return o
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not find env file [%v], using defaults", err)
	}
	globals.OPTS = NewOpts()

	globals.DB, err = database.InitDb()
	if err != nil {
		log.Fatal(err)
	}

	server := &api.Server{}
	router := mux.NewRouter()

	serverOptions := api.GorillaServerOptions{
		BaseURL:    "",
		BaseRouter: router,
		Middlewares: []api.MiddlewareFunc{
			jwt.AuthMiddleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			// Define your custom error handling logic
			fmt.Println("Handling error:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}

	handler := api.HandlerWithOptions(server, serverOptions)

	log.Println("Starting forum on port " + strconv.Itoa(globals.OPTS.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(globals.OPTS.Port), handler))

}
