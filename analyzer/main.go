package main

import (
	"flag"
	"fmt"
	"github.com/MatusOllah/slogcolor"
	"github.com/dimiro1/banner"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var (
	cfg    *Config
	logger = slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:         slog.LevelDebug,
		TimeFormat:    "15:04:05.000",
		SrcFileMode:   slogcolor.ShortFile,
		SrcFileLength: 16,
		MsgPrefix:     color.HiWhiteString("|"),
		MsgColor:      color.New(color.FgHiWhite),
		MsgLength:     16,
	}))
)

type Config struct {
	Port               int
	Env                string
	HS256              string
	BackendToken       string
	BackendCallbackUrl string
	GitBasePath        string
	AnalyzerUsername   string
	AnalyzerPassword   string
	BackendUsername    string
	BackendPassword    string
}

func init() {
	color.NoColor = false
	slog.SetDefault(logger)
}

func parseFlags() {
	cfg = &Config{}

	flag.StringVar(&cfg.Env, "env", LookupEnv("ENV"), "ENV variable")
	flag.IntVar(&cfg.Port, "port", LookupEnvInt("PORT", 9083), "listening HTTP port")
	flag.StringVar(&cfg.GitBasePath, "git-base", LookupEnv("GIT_BASE", "/tmp"), "Git base storage path")

	flag.StringVar(&cfg.AnalyzerUsername, "analyzer-username", LookupEnv("ANALYZER_USERNAME"), "Username for accessing API")
	flag.StringVar(&cfg.AnalyzerPassword, "analyzer-password", LookupEnv("ANALYZER_PASSWORD"), "Password for accessing API")

	flag.StringVar(&cfg.BackendCallbackUrl, "callback", LookupEnv("BACKEND_CALLBACK_URL"), "Callback URL")
	flag.StringVar(&cfg.BackendUsername, "backend-username", LookupEnv("BACKEND_USERNAME"), "Username for accessing backend API")
	flag.StringVar(&cfg.BackendPassword, "backend-password", LookupEnv("BACKEND_PASSWORD"), "Password for accessing backend API")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if cfg.Env == "local" || cfg.Env == "dev" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}
}

func main() {
	//the .env should be loaded before showing the banner, as the banner shows also the ENVs
	err := godotenv.Load()
	if err != nil {
		slog.Info("Could not find .env file, using defaults")
	}
	//this will set the default ENVs
	parseFlags()

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		slog.Info("could not display banner...")
	}
	credentials := Credentials{
		Username: cfg.AnalyzerUsername,
		Password: cfg.AnalyzerPassword,
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /analyze", BasicAuth(credentials, analyze))

	slog.Info("Starting FlatFeeStack Git Analyzer", "port", cfg.Port)
	err = http.ListenAndServe(":"+strconv.Itoa(cfg.Port), router)
	if err != nil {
		slog.Error("Server stopped", slog.Any("error", err))
	}
}
