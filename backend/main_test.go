package main

import (
	"backend/clients"
	"backend/db"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/go-jose/go-jose/v3/json"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	opts = &Opts{}
	opts.Env = "test"
	opts.DBDriver = "sqlite3"
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())
	opts.DBPath = file.Name()

	err = dbLib.InitDb("sqlite3", file.Name(), "db/init.sql")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	os.Exit(code)
}

func setup() {
	err := dbLib.RunSQL("db/init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}

func teardown() {
	err := dbLib.RunSQL("db/delAll_test.sql")
	if err != nil {
		log.Fatalf("Could not run delAll_test.sql: %s", err)
	}
	clients.EmailNotifications = 0
	clients.EmailNoNotifications = 0
}

func setupAnalysisTestServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/analyze":
			var request db.AnalysisRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.Nil(t, err)

			err = json.NewEncoder(w).Encode(clients.AnalysisResponse2{RequestId: request.Id})
			require.Nil(t, err)
		default:
			http.NotFound(w, r)
		}
	}))

	clients.InitAnalyzer(server.URL, "test", "test")

	return server
}

func setupPayoutTestServer(t *testing.T) *httptest.Server {
	usdcData, err := os.ReadFile("./fixtures/payout_response_usdc.json")
	require.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/sign/usdc":
			w.WriteHeader(200)
			w.Write(usdcData)
		default:
			http.NotFound(w, r)
		}
	}))

	opts.Payout.Url = server.URL

	return server
}
