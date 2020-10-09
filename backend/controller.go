package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

type BaseHandler struct {
	userRepo UserRepository
	repoRepo RepoRepository
}

// NewBaseHandler returns a new BaseHandler
func NewBaseHandler(userRepo UserRepository, repoRepo RepoRepository) *BaseHandler {
	return &BaseHandler{
		userRepo: userRepo,
		repoRepo: repoRepo,
	}
}

// HttpResponse format
type HttpResponse struct {
	Success bool `json:"success"`
	Message string `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}


func NewInternalError(message string) HttpResponse{
	return HttpResponse{Success: false, Message: message}
}


/*
 *	==== USER ====
 */

func (h *BaseHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var user User
	jsonErr := json.NewDecoder(r.Body).Decode(&user)
	if jsonErr != nil {
		res := NewInternalError("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// generate uuuid
	uid, uuidErr := uuid.NewRandom()
	if uuidErr != nil{
		res := NewInternalError("Unable to create uuid.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	user.ID = uid.String()
	// call insert user function and pass the user
	dbErr := h.userRepo.Save(&user)

	if dbErr != nil {
		res := NewInternalError("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data: user,
		Message: "User created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

// GetUser will return a single user by its id
func(h *BaseHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	user, err := h.userRepo.FindByID(id)
	if err != nil {
		w.WriteHeader(404)
		res := NewInternalError("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// send the HttpResponse
	_  = json.NewEncoder(w).Encode(user)
}

/*
 *	==== REPO ====
 */

func(h *BaseHandler) CreateRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var repo Repo
	jsonErr := json.NewDecoder(r.Body).Decode(&repo)
	if jsonErr != nil {
		res := NewInternalError("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// generate uuuid
	uid, uuidErr := uuid.NewRandom()
	if uuidErr != nil{
		res := NewInternalError("Unable to create uuid.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	repo.ID = uid.String()
	// call insert user function and pass the user
	dbErr := h.repoRepo.Save(&repo)

	if dbErr != nil {
		res := NewInternalError("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data: repo,
		Message: "Repo created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

func(h *BaseHandler) GetRepoByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	repo, err := h.repoRepo.FindByID(id)
	if err != nil {
		w.WriteHeader(404)
		res := NewInternalError("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// send the HttpResponse
	_  = json.NewEncoder(w).Encode(repo)
}



