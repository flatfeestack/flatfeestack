package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"log/slog"
	"net/http"
)

func GetUserById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId := params["id"]
	convertedUserId, err := uuid.Parse(userId)

	if err != nil {
		slog.Error("Invalid user ID",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	}

	user, err := db.FindPublicUserById(convertedUserId)

	if user == nil {
		slog.Error("User not found %s",
			slog.String("userId", userId))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		slog.Error("Could not fetch user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, user)
	}

	return
}

func GetMyUser(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	util.WriteJson(w, user)
}

func DeleteMethod(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	user.PaymentMethod = nil
	user.Last4 = nil
	err := db.UpdateStripe(user)
	if err != nil {
		slog.Error("Could not delete method:",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func UpdateMethod(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	if user.StripeId == nil {
		slog.Error("Stripe ID is missing",
			slog.String("email", user.Email))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	params := mux.Vars(r)
	a := params["method"]

	user.PaymentMethod = &a
	pm, err := paymentmethod.Get(
		*user.PaymentMethod,
		nil,
	)
	if err != nil {
		slog.Error("Could not update retrieve payment method",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	_, err = paymentmethod.Attach(*user.PaymentMethod, &stripe.PaymentMethodAttachParams{
		Customer: user.StripeId,
	})
	if err != nil {
		slog.Error("Could not attach payment method to Stripe user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	user.Last4 = &pm.Card.Last4
	err = db.UpdateStripe(user)
	if err != nil {
		slog.Error("Could not update stripe method",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	util.WriteJson(w, user)
}

func UpdateName(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	a := params["name"]
	err := db.UpdateUserName(user.Id, a)
	if err != nil {
		slog.Error("Could not save name",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Could not save username. Please try again.")
		return
	}
}

func ClearName(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	err := db.ClearUserName(user.Id)
	if err != nil {
		slog.Error("Could not clear username",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Could not clear username. Please try again.")
		return
	}
}

func UpdateImage(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	var img ImageRequest
	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		slog.Error("Could not decode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = db.UpdateUserImage(user.Id, img.Image)
	if err != nil {
		slog.Error("Could not update user image",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Could not update user image. Please try again")
		return
	}
}

func DeleteImage(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	err := db.DeleteUserImage(user.Id)
	if err != nil {
		slog.Error("Could not delete user image",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Could not delete user image. Please try again")
		return
	}
}

func Users(w http.ResponseWriter, _ *http.Request, u *db.UserDetail) {
	users, err := db.FindAllEmails()
	if err != nil {
		slog.Error("Could not fetch users",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Could not fetch users. Please try again.")
		return
	}
	util.WriteJson(w, users)
}

func FakeUser(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	slog.Info("fake user")
	m := mux.Vars(r)
	n := m["email"]

	uid := uuid.New()

	u := db.User{
		Email:     n,
		Id:        uid,
		CreatedAt: util.TimeNow(),
	}
	ud := db.UserDetail{
		User: u,
	}

	err := db.InsertUser(&ud)
	if err != nil {
		slog.Error("Could insert fake user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Could not insert user. Please try again.")
		return
	}

	id := uuid.New()
	err = db.InsertGitEmail(id, uid, n, nil, util.TimeNow())
	if err != nil {
		slog.Error("Could not insert git email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "There is a problem with your git email. Please try again.")
		return
	}
}

func UserSummary2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		slog.Error("Parameter hours not set",
			slog.Any("params", m))
		util.WriteErrorf(w, http.StatusBadRequest, "Parameter not set. Please try again.")
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		slog.Error("Could not parse UUID",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	user, err := db.FindUserById(uu)
	if err != nil {
		slog.Error("Could not find user by id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
	util.WriteJson(w, user2D)
}

func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	email := m["email"]
	if email == "" {
		util.WriteErrorf(w, http.StatusBadRequest, "Parameter email not set: %v", m)
		return
	}

	user, err := db.FindUserByEmail(email)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not find user: %v", err)
		return
	}

	util.WriteJson(w, user)
}
