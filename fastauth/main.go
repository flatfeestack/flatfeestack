package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/MatusOllah/slogcolor"
	"github.com/dimiro1/banner"
	"github.com/fatih/color"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/ed25519"
	"hash/fnv"
	"log/slog"
	rnd "math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	opts         *Opts
	jwtKey       []byte
	privRSA      *rsa.PrivateKey
	privRSAKid   string
	privEdDSA    *ed25519.PrivateKey
	privEdDSAKid string
	tokenExp     time.Duration
	refreshExp   time.Duration
	codeExp      time.Duration
	secondsAdd   int
	admins       []string
	debug        bool
	logger       = slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:         slog.LevelDebug,
		TimeFormat:    "15:04:05.000",
		SrcFileMode:   slogcolor.ShortFile,
		SrcFileLength: 16,
		MsgPrefix:     color.HiWhiteString("|"),
		MsgColor:      color.New(color.FgHiWhite),
		MsgLength:     16,
	}))
)

type Credentials struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
	TOTP     string `json:"totp,omitempty"`
	//here comes oauth, leave empty on regular login
	//If you want to use oauth, you need to configure
	//client-id with a matching redirect-uri from the
	//command line
	ClientId                string `json:"client_id,omitempty"`
	ResponseType            string `json:"response_type,omitempty"`
	State                   string `json:"state,omitempty"`
	Scope                   string `json:"scope"`
	RedirectUri             string `json:"redirect_uri,omitempty"`
	CodeChallenge           string `json:"code_challenge,omitempty"`
	CodeCodeChallengeMethod string `json:"code_challenge_method,omitempty"`
	//Token stuff
	EmailToken    string `json:"emailToken,omitempty"`
	RedirectAs201 bool   `json:"redirectAs201,omitempty"`
}

type TokenClaims struct {
	Scope string `json:"scope,omitempty"`
	jwt.Claims
}

type RefreshClaims struct {
	ExpiresAt int64  `json:"exp,omitempty"`
	Subject   string `json:"role,omitempty"`
	Token     string `json:"token,omitempty"`
}
type CodeClaims struct {
	ExpiresAt               int64  `json:"exp,omitempty"`
	Subject                 string `json:"role,omitempty"`
	CodeChallenge           string `json:"code_challenge,omitempty"`
	CodeCodeChallengeMethod string `json:"code_challenge_method,omitempty"`
}

type ProvisioningUri struct {
	Uri string `json:"uri"`
}

type OAuth struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expires      string `json:"expires_in"`
}

// system user does not need refresh token
type OAuthSystem struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expires     string `json:"expires_in"`
}

type Opts struct {
	Env             string
	Dev             string
	Issuer          string
	Port            int
	DBPath          string
	DBDriver        string
	DBScripts       string
	EmailFrom       string
	EmailFromName   string
	EmailUrl        string
	EmailToken      string
	EmailLinkPrefix string
	Audience        string
	ExpireAccess    int
	ExpireRefresh   int
	ExpireCode      int
	HS256           string
	EdDSA           string
	RS256           string
	OAuthUser       string
	OAuthPass       string
	ResetRefresh    bool
	Users           string
	AdminEndpoints  bool
	UserEndpoints   bool
	OauthEndpoints  bool
	DetailedError   bool
	Redirects       string
	PasswordFlow    bool
	Scope           string
	LogPath         string
	Admins          string
}

func init() {
	color.NoColor = false
	slog.SetDefault(logger)
}

func hash(s string) int64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func NewOpts() *Opts {
	opts := &Opts{}
	flag.StringVar(&opts.Env, "env", LookupEnv("ENV"), "ENV variable")
	flag.StringVar(&opts.Dev, "dev", LookupEnv("DEV"), "Dev settings with initial secret")
	flag.StringVar(&opts.Issuer, "issuer", LookupEnv("ISSUER"), "name of issuer, default in dev is my-issuer")
	flag.IntVar(&opts.Port, "port", LookupEnvInt("PORT",
		8080), "listening HTTP port")
	flag.StringVar(&opts.DBPath, "db-path", LookupEnv("DB_PATH",
		"./fastauth.db"), "DB path")
	flag.StringVar(&opts.DBDriver, "db-driver", LookupEnv("DB_DRIVER",
		"sqlite3"), "DB driver")
	flag.StringVar(&opts.DBScripts, "db-scripts", LookupEnv("DB_SCRIPTS"), "DB scripts to run at startup")
	flag.StringVar(&opts.EmailFrom, "email-from", LookupEnv("EMAIL_FROM"), "Email from, default is info@flatfeestack.io")
	flag.StringVar(&opts.EmailFromName, "email-from-name", LookupEnv("EMAIL_FROM_NAME",
		"email@fastauth"), "Email from name, default is a empty string")
	flag.StringVar(&opts.EmailUrl, "email-url", LookupEnv("EMAIL_URL"), "Email service URL")
	flag.StringVar(&opts.EmailToken, "email-token", LookupEnv("EMAIL_TOKEN"), "Email service token")
	flag.StringVar(&opts.EmailLinkPrefix, "email-prefix", LookupEnv("EMAIL_PREFIX"), "Email link prefix")
	flag.StringVar(&opts.Audience, "audience", LookupEnv("AUDIENCE"), "Audience, default in dev is my-audience")
	flag.IntVar(&opts.ExpireAccess, "expire-access", LookupEnvInt("EXPIRE_ACCESS",
		30*60), "Access token expiration in seconds, default 30min")
	flag.IntVar(&opts.ExpireRefresh, "expire-refresh", LookupEnvInt("EXPIRE_REFRESH",
		180*24*60*60), "Refresh token expiration in seconds, default 6month")
	flag.IntVar(&opts.ExpireCode, "expire-code", LookupEnvInt("EXPIRE_CODE",
		60), "Authtoken flow expiration in seconds, default 1min")
	flag.StringVar(&opts.HS256, "hs256", LookupEnv("HS256"), "HS256 key")
	flag.StringVar(&opts.RS256, "rs256", LookupEnv("RS256"), "RS256 key")
	flag.StringVar(&opts.EdDSA, "eddsa", LookupEnv("EDDSA"), "EdDSA key")
	flag.StringVar(&opts.OAuthUser, "oauth-user", LookupEnv("OAUTH_USER"), "Basic auth username for the server meta data")
	flag.StringVar(&opts.OAuthPass, "oauth-pass", LookupEnv("OAUTH_PASS"), "Basic auth password for the server meta data")
	flag.BoolVar(&opts.ResetRefresh, "reset-refresh", LookupEnv("RESET_REFRESH") != "", "Reset refresh token when setting the token")
	flag.StringVar(&opts.Users, "users", LookupEnv("USERS"), "add these initial users. E.g, -users tom@test.ch:pw123;test@test.ch:123pw")
	flag.BoolVar(&opts.AdminEndpoints, "admin-endpoints", LookupEnv("ADMIN_ENDPOINTS") != "", "Enable admin-facing endpoints. In dev mode these are enabled by default")
	flag.BoolVar(&opts.UserEndpoints, "user-endpoints", LookupEnv("USER_ENDPOINTS") != "", "Enable user-facing endpoints. In dev mode these are enabled by default")
	flag.BoolVar(&opts.OauthEndpoints, "oauth-enpoints", LookupEnv("OAUTH_ENDPOINTS") != "", "Enable oauth-facing endpoints. In dev mode these are enabled by default")
	flag.BoolVar(&opts.DetailedError, "details", LookupEnv("DETAILS") != "", "Enable detailed errors")
	flag.StringVar(&opts.Redirects, "redir", LookupEnv("REDIR"), "add client redirects. E.g, -redir clientId1:http://blabla;clientId2:http://blublu")
	flag.BoolVar(&opts.PasswordFlow, "pwflow", LookupEnv("PWFLOW") != "", "enable password flow, default disabled")
	flag.StringVar(&opts.Scope, "scope", LookupEnv("SCOPE"), "scope, default in dev is my-scope")
	flag.StringVar(&opts.LogPath, "log", LookupEnv("LOG",
		os.TempDir()+"/ffs"), "Log directory, default is /tmp/ffs")
	flag.StringVar(&opts.Admins, "admins", LookupEnv("ADMINS"), "Admins")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//set defaults, be explicit
	if opts.Env == "local" || opts.Env == "dev" {
		debug = true
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	//set defaults
	if opts.Dev != "" {
		if opts.Scope == "" {
			opts.Scope = "my-scope"
		}
		if opts.Audience == "" {
			opts.Audience = "my-audience"
		}
		if opts.Issuer == "" {
			opts.Issuer = "my-issuer"
		}
		if opts.EmailUrl == "" {
			opts.EmailUrl = "http://localhost:8080/send/email/{email}/{token}"
		}

		if strings.ToLower(opts.RS256) == "true" {
			//work around this issue: https://github.com/golang/go/issues/38548
			//we want for testing to have the same key, I don't care for any database keys
			rsaPrivKey1, err := rsa.GenerateKey(rnd.New(rnd.NewSource(hash(opts.Dev))), 2048)
			if err != nil {
				slog.Error("cannot generate rsa key", slog.Any("error", err))
			}
			rsaPrivKey2, err := rsa.GenerateKey(rnd.New(rnd.NewSource(hash(opts.Dev))), 2048)
			if err != nil {
				slog.Error("cannot generate rsa key", slog.Any("error", err))
			}
			for rsaPrivKey1.Equal(rsaPrivKey2) {
				rsaPrivKey2, err = rsa.GenerateKey(rnd.New(rnd.NewSource(hash(opts.Dev))), 2048)
				if err != nil {
					slog.Error("cannot generate rsa key", slog.Any("error", err))
				}
			}
			rsaPrivKey := rsaPrivKey1
			if rsaPrivKey1.N.Cmp(rsaPrivKey2.N) > 0 {
				rsaPrivKey = rsaPrivKey2
			}

			if err != nil {
				slog.Error("cannot generate rsa key", slog.Any("error", err))
			}
			encPrivRSA, err := x509.MarshalPKCS8PrivateKey(rsaPrivKey)
			if err != nil {
				slog.Error("cannot generate rsa key", slog.Any("error", err))
			}
			opts.RS256 = base32.StdEncoding.EncodeToString(encPrivRSA)
		} else if strings.ToLower(opts.EdDSA) == "true" {
			_, edPrivKey, err := ed25519.GenerateKey(rnd.New(rnd.NewSource(hash(opts.Dev))))
			if err != nil {
				slog.Error("cannot generate eddsa key", slog.Any("error", err))
			}
			opts.EdDSA = base32.StdEncoding.EncodeToString(edPrivKey)
		} else if strings.ToLower(opts.HS256) != "true" && opts.HS256 != "" {
			slog.Warn("DEV mode enabled, ignoring key", slog.String("opts.HS256", opts.HS256))
		} else {
			h := sha256.New()
			h.Write([]byte(opts.Dev))
			jwtKey = h.Sum(nil)
			opts.HS256 = base32.StdEncoding.EncodeToString(jwtKey)
		}
		if opts.OAuthUser == "" {
			opts.OAuthUser = "clientId"
		}
		if opts.OAuthPass == "" {
			opts.OAuthPass = "secret"
		}
		opts.AdminEndpoints = true
		opts.OauthEndpoints = true
		opts.UserEndpoints = true
		opts.DetailedError = true
		opts.PasswordFlow = true

		if opts.Users == "" {
			opts.Users = "user:pass"
		}

		slog.Debug("DEV mode active, key is %v, hex(%v)", slog.String("opts.Dev", opts.Dev), slog.String("opts.HS256", opts.HS256))
		slog.Debug("DEV mode active, rsa is hex(%v)", slog.String("opts.RS256", opts.RS256))
		slog.Debug("DEV mode active, eddsa is hex(%v)", slog.String("opts.EdDSA", opts.EdDSA))
	}

	admins = strings.Split(opts.Admins, ";")
	if opts.OAuthUser != "" {
		admins = append(admins, opts.OAuthUser)
	}

	// Check that exactly one of HS256, RS256, or EdDSA is set.
	count := 0
	if opts.HS256 != "" {
		count += 1
	}

	if opts.RS256 != "" {
		count += 1
	}

	if opts.EdDSA != "" {
		count += 1
	}
	if count != 1 {
		fmt.Printf("Exactly one of hs256, rs256, or eddsa must be set. Choose one\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if opts.HS256 != "" {
		var err error
		jwtKey, err = base32.StdEncoding.DecodeString(opts.HS256)
		if err != nil {
			slog.Error("cannot hash", slog.Any("error", err))
		}
	}

	if opts.RS256 != "" {
		rsaDec, err := base32.StdEncoding.DecodeString(opts.RS256)
		if err != nil {
			slog.Error("cannot decode", slog.String("opts.RS256", opts.RS256))
		}
		i, err := x509.ParsePKCS8PrivateKey(rsaDec)
		if err != nil {
			slog.Error("cannot create private key", slog.String("rsaDec", opts.RS256))
		}
		privRSA = i.(*rsa.PrivateKey)
		k := jose.JSONWebKey{Key: privRSA.Public()}
		kid, err := k.Thumbprint(crypto.SHA256)
		if err != nil {
			slog.Error("cannot thumb rsa", slog.Any("error", err))
		}
		privRSAKid = hex.EncodeToString(kid)
		slog.Info("kid", slog.String("privRSAKid", privRSAKid))
	}

	if opts.EdDSA != "" {
		eddsa, err := base32.StdEncoding.DecodeString(opts.EdDSA)
		if err != nil {
			slog.Error("cannot decode", slog.String("opts.EdDSA", opts.EdDSA))
		}
		privEdDSA0 := ed25519.PrivateKey(eddsa)
		privEdDSA = &privEdDSA0
		k := jose.JSONWebKey{Key: privEdDSA.Public()}
		kid, err := k.Thumbprint(crypto.SHA256)
		if err != nil {
			slog.Error("cannot thumb eddsa", slog.Any("error", err))
		}
		privEdDSAKid = hex.EncodeToString(kid)
	}

	tokenExp = time.Second * time.Duration(opts.ExpireAccess)
	refreshExp = time.Second * time.Duration(opts.ExpireRefresh)
	codeExp = time.Second * time.Duration(opts.ExpireCode)

	return opts
}

func checkEmailPassword(email string, password string) (*dbRes, string, error) {
	result, err := findAuthByEmail(email)
	if err != nil {
		return nil, "not-found", fmt.Errorf("ERR-checkEmail-01, DB select, %v err %v", email, err)
	}

	if result.emailToken != nil {
		return nil, "blocked", fmt.Errorf("ERR-checkEmail-02, user %v no email verified: %v", email, err)
	}

	if result.errorCount > 2 {
		return nil, "blocked", fmt.Errorf("ERR-checkEmail-03, user %v no email verified: %v", email, err)
	}

	storedPw, calcPw, err := checkPw(password, result.password)
	if err != nil {
		return nil, "blocked", fmt.Errorf("ERR-checkEmail-04, key %v error: %v", email, err)
	}

	if !bytes.Equal(calcPw, storedPw) {
		err = incErrorCount(email)
		if err != nil {
			return nil, "blocked", fmt.Errorf("ERR-checkEmail-05, key %v error: %v", email, err)
		}
		return nil, "refused", fmt.Errorf("ERR-checkEmail-06, user %v password mismatch", email)
	}
	err = resetCount(email)
	if err != nil {
		return nil, "blocked", fmt.Errorf("ERR-checkEmail-05, key %v error: %v", email, err)
	}
	return result, "", nil
}

func middlewareLog(handlerFunc func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	loggedHandler := logRequestHandler(http.HandlerFunc(handlerFunc))
	return loggedHandler
}

func middlewareJwtLog(handlerFunc func(http.ResponseWriter, *http.Request, *TokenClaims)) func(w http.ResponseWriter, r *http.Request) {
	jh := jwtAuth(handlerFunc)
	return middlewareLog(jh)
}

func middlewareJwtAdminLog(handlerFunc func(http.ResponseWriter, *http.Request, string)) func(w http.ResponseWriter, r *http.Request) {
	jh := jwtAuthAdmin(handlerFunc, admins)
	return middlewareLog(jh)
}

func setupRouter() *http.ServeMux {
	router := http.NewServeMux()

	if opts.UserEndpoints {
		router.HandleFunc("POST /login", middlewareLog(login))
		router.HandleFunc("POST /refresh", middlewareLog(refresh))
		router.HandleFunc("POST /signup", middlewareLog(signup))
		router.HandleFunc("POST /reset/{email}", middlewareLog(resetEmail))
		router.HandleFunc("GET /confirm/signup/{email}/{token}", middlewareLog(confirmEmail))
		router.HandleFunc("POST /confirm/signup", middlewareLog(confirmEmailPost))
		router.HandleFunc("POST /confirm/invite", middlewareLog(confirmInvite))
		router.HandleFunc("POST /confirm/reset", middlewareLog(confirmReset))

		router.HandleFunc("POST /invite/{email}", middlewareJwtLog(invite))
		router.HandleFunc("POST /setup/totp", middlewareJwtLog(setupTOTP))
		router.HandleFunc("POST /confirm/totp/{token}", middlewareJwtLog(confirmTOTP))
	}
	//logout
	router.HandleFunc("GET /authen/logout", middlewareJwtLog(logout))

	//maintenance stuff
	router.HandleFunc("GET /readiness", middlewareLog(readiness))
	router.HandleFunc("GET /liveness", middlewareLog(liveness))

	//display for debug and testing
	if debug {
		//admin endpoints
		router.HandleFunc("GET /admin/time", middlewareJwtAdminLog(serverTime))
		router.HandleFunc("POST /admin/timewarp/{hours}", middlewareJwtAdminLog(timeWarp))
	}

	if opts.OauthEndpoints {
		router.HandleFunc("POST /oauth/login", middlewareLog(login))
		router.HandleFunc("POST /oauth/token", middlewareLog(oauth))
		router.HandleFunc("POST /oauth/revoke", middlewareLog(revoke))
		router.HandleFunc("GET /oauth/authorize", middlewareLog(authorize))
		//convenience function
		if opts.Env == "dev" || opts.Env == "local" {
			router.HandleFunc("GET /", middlewareLog(authorize))
		}
		router.HandleFunc("GET /.well-known/jwks.json", middlewareLog(jwkFunc))
	}

	if opts.AdminEndpoints {
		router.HandleFunc("POST /users/usernames/{email}/cancellation", middlewareJwtAdminLog(deleteUser))
		router.HandleFunc("PATCH /users/usernames/{email}/attributes", middlewareJwtAdminLog(updateUser))
		router.HandleFunc("POST /admin/login-as/{email}", middlewareJwtAdminLog(asUser))
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("[404] no route matched for", "url", r.URL, "method", r.Method)
		w.WriteHeader(http.StatusNotFound)
	})

	return router
}

func main() {
	//the .env should be loaded before showing the banner, as the banner shows also the ENVs
	err := godotenv.Load()
	if err != nil {
		slog.Info("Could not find .env file, using defaults")
	}

	opts = NewOpts()

	f, err := os.Open("banner.txt")
	if err == nil {
		banner.Init(os.Stdout, true, false, f)
	} else {
		slog.Info("could not display banner...")
	}

	err = InitDb(opts.DBDriver, opts.DBPath, opts.DBScripts)
	if err != nil {
		slog.Error("DB not initialized",
			slog.Any("error", err))
		os.Exit(1)
	}
	defer CloseDb()

	addInitialUsers()

	router := setupRouter()

	s := &http.Server{
		Addr:         ":" + strconv.Itoa(opts.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		slog.Error("Listen failed",
			slog.Any("error", err))
	}

	if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
		slog.Error("Server stopped",
			slog.Any("error", err))
	}
}
