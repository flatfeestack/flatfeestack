package api

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"net/http"
)

func GetMyUser(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	utils.WriteJson(w, user)
}

func DeleteMethod(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	user.PaymentMethod = nil
	user.Last4 = nil
	err := db.UpdateStripe(user)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could update user: %v", err)
		return
	}
}

func UpdateMethod(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	a := params["method"]

	user.PaymentMethod = &a
	pm, err := paymentmethod.Get(
		*user.PaymentMethod,
		nil,
	)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could update method: %v", err)
		return
	}

	user.Last4 = &pm.Card.Last4
	err = db.UpdateStripe(user)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could update user: %v", err)
		return
	}

	utils.WriteJson(w, user)
}

func UpdateName(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	a := params["name"]
	err := db.UpdateUserName(user.Id, a)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
}

func ClearName(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	err := db.ClearUserName(user.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
}

func UpdateImage(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	var img ImageRequest
	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = db.UpdateUserImage(user.Id, img.Image)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
}

func Users(w http.ResponseWriter, _ *http.Request, _ string) {
	u, err := db.FindAllEmails()
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could fetch users: %v", err)
		return
	}
	utils.WriteJson(w, u)
}

func FakeUser(w http.ResponseWriter, r *http.Request, _ string) {
	log.Printf("fake user")
	m := mux.Vars(r)
	n := m["email"]

	uid := uuid.New()

	u := db.User{
		Email:     n,
		Id:        uid,
		CreatedAt: utils.TimeNow(),
	}
	ud := db.UserDetail{
		User: u,
	}

	err := db.InsertUser(&ud)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	id := uuid.New()
	err = db.InsertGitEmail(id, uid, n, nil, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}
}

func UserSummary2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		utils.WriteErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := db.FindUserById(uu)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user2 := db.User{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
	user2D := db.UserDetail{
		User:  user2,
		Image: user.Image,
	}
	utils.WriteJson(w, user2D)
}
