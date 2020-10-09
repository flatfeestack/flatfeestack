package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	// open DB connectino
	db := createConnection()

	// Repos

	userRepo := NewUserRepo(db)
	repoRepo := NewRepoRepo(db)

	h := NewBaseHandler(userRepo, repoRepo)


	// Routes
	protectedRouter := mux.NewRouter()
	protectedRouter.HandleFunc("/api/users/{id}", h.GetUserByID).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/api/users", h.CreateUser).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/api/repos", h.CreateRepo).Methods("POST", "OPTIONS")
	protectedRouter.Use(AuthMiddleware)

	publicRouter := mux.NewRouter()
	publicRouter.HandleFunc("/api/repos/{id}", h.GetRepoByID).Methods("GET", "OPTIONS")


	// http.Handle("/", fs)
	fmt.Println("Starting server on the port 8080...")

	log.Fatal(http.ListenAndServe(":8080", protectedRouter))
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