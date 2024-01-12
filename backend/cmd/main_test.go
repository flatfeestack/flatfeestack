package main

import (
	"backend/pkg/config"
	dbLib "github.com/flatfeestack/go-lib/database"
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cfg = &config.Config{}
	cfg.Env = "test"
	cfg.DBDriver = "sqlite3"
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())
	cfg.DBPath = file.Name()

	err = dbLib.InitDb("sqlite3", file.Name(), "db/init.sql")
	if err != nil {
		slog.Error("DB error", slog.Any("error", err))
	}

	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		slog.Warn("Could not start resource", slog.Any("error", err))
	}

	if err != nil {
		slog.Warn("Could not start resource:", slog.Any("error", err))
	}

	os.Exit(code)
}

/*func setup() {
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

func SetupAnalysisTestServer(t *testing.T) *httptest.Server {
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
}*/
