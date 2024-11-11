package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func GetTrustValueById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	convertedTrustValueId, err := strconv.Atoi(id)

	if err != nil {
		slog.Error("Invalid user ID",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	}

	trustValue, err := db.FindTrustValueById(convertedTrustValueId)

	if trustValue == nil {
		slog.Error("Trust Value not found %s",
			slog.String("id", id))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		slog.Error("Could not fetch trust value",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, trustValue)
	}
}

// I don't think we need to delete entries for trust value
//func DeleteMethod(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
//	user.PaymentMethod = nil
//	user.Last4 = nil
//	err := db.UpdateStripe(user)
//	if err != nil {
//		slog.Error("Could not delete method:",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
//		return
//	}
//}

// func UpdateMethod(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
func UpdateTrustValue(w http.ResponseWriter, r *http.Request, trustValue *db.TrustValueMetrics) {
	if trustValue.RepoId == uuid.Nil {
		slog.Error("RepoId is missing",
			slog.String("Trust Value id", string(trustValue.Id)))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	err := db.UpdateTrustValue(*trustValue)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not update trust value: %v", err)
		return
	}
	trustValue, err = db.FindTrustValueById(trustValue.Id)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not find trust value: %v", err)
		return
	}

	util.WriteJson(w, trustValue)
}

//
//	//a := r.PathValue("method")
//
//	//user.PaymentMethod = &a
//	//pm, err := paymentmethod.Get(
//	//	*user.PaymentMethod,
//	//	nil,
//	//)
//	//if err != nil {
//	//	slog.Error("Could not update retrieve payment method",
//	//		slog.Any("error", err))
//	//	util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
//	//	return
//	//}
//
//	//_, err = paymentmethod.Attach(*user.PaymentMethod, &stripe.PaymentMethodAttachParams{
//	//	Customer: user.StripeId,
//	//})
//	//if err != nil {
//	//	slog.Error("Could not attach payment method to Stripe user",
//	//		slog.Any("error", err))
//	//	util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
//	//	return
//	//}
//
//	//user.Last4 = &pm.Card.Last4
//	//err = db.UpdateStripe(user)
//	//if err != nil {
//	//	slog.Error("Could not update stripe method",
//	//		slog.Any("error", err))
//	//	util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
//	//	return
//	//}
//
//	util.WriteJson(w, user)
//}

//func UpdateName(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
//	a := r.PathValue("name")
//	err := db.UpdateUserName(user.Id, a)
//	if err != nil {
//		slog.Error("Could not save name",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusInternalServerError, "Could not save username. Please try again.")
//		return
//	}
//}
//
//func Users(w http.ResponseWriter, _ *http.Request, u *db.UserDetail) {
//	users, err := db.FindAllEmails()
//	if err != nil {
//		slog.Error("Could not fetch users",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusBadRequest, "Could not fetch users. Please try again.")
//		return
//	}
//	util.WriteJson(w, users)
//}
//
//func FakeUser(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
//	slog.Info("fake user")
//	n := r.PathValue("email")
//
//	uid := uuid.New()
//
//	u := db.User{
//		Email:     n,
//		Id:        uid,
//		CreatedAt: util.TimeNow(),
//	}
//	ud := db.UserDetail{
//		User: u,
//	}
//
//	err := db.InsertUser(&ud)
//	if err != nil {
//		slog.Error("Could insert fake user",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusBadRequest, "Could not insert user. Please try again.")
//		return
//	}
//
//	id := uuid.New()
//	err = db.InsertGitEmail(id, uid, n, nil, util.TimeNow())
//	if err != nil {
//		slog.Error("Could not insert git email",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusBadRequest, "There is a problem with your git email. Please try again.")
//		return
//	}
//}
//
//func UserSummary2(w http.ResponseWriter, r *http.Request) {
//	u := r.PathValue("uuid")
//	if u == "" {
//		slog.Error("Parameter hours not set")
//		util.WriteErrorf(w, http.StatusBadRequest, "Parameter not set. Please try again.")
//		return
//	}
//
//	uu, err := uuid.Parse(u)
//	if err != nil {
//		slog.Error("Could not parse UUID",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
//		return
//	}
//
//	user, err := db.FindUserById(uu)
//	if err != nil {
//		slog.Error("Could not find user by id",
//			slog.Any("error", err))
//		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
//		return
//	}
//
//	user2 := db.User{
//		Id:    user.Id,
//		Name:  user.Name,
//		Email: user.Email,
//	}
//	user2D := db.UserDetail{
//		User:  user2,
//		Image: user.Image,
//	}
//	util.WriteJson(w, user2D)
//}
//
//func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
//	email := r.PathValue("email")
//	if email == "" {
//		util.WriteErrorf(w, http.StatusBadRequest, "Parameter email not set")
//		return
//	}
//
//	user, err := db.FindUserByEmail(email)
//	if err != nil {
//		util.WriteErrorf(w, http.StatusNoContent, "Could not find user: %v", err)
//		return
//	}
//
//	util.WriteJson(w, user)
//}
//
