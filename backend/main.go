package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/flatfeestack/api/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)
// @title Flatfeestack API
// @version 0.0.1
// @host localhost:8080
// @BasePath /
func main() {
	db := createConnection()

	userRepo := NewUserRepo(db)
	repoRepo := NewRepoRepo(db)

	h := NewBaseHandler(userRepo, repoRepo)


	// Routes
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/users/{id}", h.GetUserByID).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users", h.CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos", h.CreateRepo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/repos/{id}", h.GetRepoByID).Methods("GET", "OPTIONS")
	//apiRouter.Use(AuthMiddleware)

	// Swagger
	router.PathPrefix("/swagger").Handler( httpSwagger.WrapHandler)

	fmt.Println("Starting server on the port 8080...")

	log.Fatal(http.ListenAndServe(":8080", router))
}


func AuthMiddleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")

		if token != "" {
			// We found the token in our map
			log.Printf("Authenticated user %s\n", token)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}


// create connection with postgres db
func createConnection() *sql.DB {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Open the connection
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}

	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}