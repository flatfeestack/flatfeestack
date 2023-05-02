package db

import (
	"database/sql"
	"fmt"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time
}

type UserDetail struct {
	User
	InvitedId          *uuid.UUID
	StripeId           *string `json:"-"`
	Image              *string `json:"image"`
	PaymentMethod      *string `json:"paymentMethod"`
	StripeClientSecret *string `json:"clientSecret"`
	Last4              *string `json:"last4"`
	Seats              int     `json:"seats"`
	Freq               int     `json:"freq"`
	Claims             *jwt.Claims
	Role               *string `json:"role,omitempty"`
}

func FindAllEmails() ([]string, error) {
	emails := []string{}
	s := `SELECT email from users`
	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func FindUserByEmail(email string) (*UserDetail, error) {
	var u UserDetail
	err := db.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
                               email, name, image, seats, freq, created_at
                        FROM users 
                        WHERE email=$1`, email).
		Scan(&u.Id, &u.StripeId, &u.InvitedId, &u.PaymentMethod, &u.Last4,
			&u.Email, &u.Name, &u.Image, &u.Seats, &u.Freq, &u.CreatedAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func FindUserById(uid uuid.UUID) (*UserDetail, error) {
	var u UserDetail
	err := db.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
                               stripe_client_secret, email, name, image, seats, freq, created_at                          
                        FROM users 
                        WHERE id=$1`, uid).
		Scan(&u.Id, &u.StripeId, &u.InvitedId, &u.PaymentMethod, &u.Last4,
			&u.StripeClientSecret, &u.Email, &u.Name, &u.Image, &u.Seats, &u.Freq, &u.CreatedAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func InsertUser(user *UserDetail) error {
	stmt, err := db.Prepare("INSERT INTO users (id, email, stripe_id, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(user.Id, user.Email, user.StripeId, user.CreatedAt)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func UpdateStripe(user *UserDetail) error {
	stmt, err := db.Prepare(`UPDATE users 
                                   SET stripe_id=$1, stripe_payment_method=$2, stripe_last4=$3
                                   WHERE id=$4`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(user.StripeId, user.PaymentMethod, user.Last4, user.Id)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateUserName(uid uuid.UUID, name string) error {
	stmt, err := db.Prepare("UPDATE users SET name=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(name, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateUserImage(uid uuid.UUID, data string) error {
	stmt, err := db.Prepare("UPDATE users SET image=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(data, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateSeatsFreq(userId uuid.UUID, seats int, freq int) error {
	stmt, err := db.Prepare("UPDATE users SET seats=$1, freq=$2 WHERE id=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(seats, freq, userId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindSponsoredUserBalances(invitedUserId uuid.UUID) ([]User, error) {
	s := `SELECT id, name, email
          FROM users 
          WHERE invited_id = $1`
	rows, err := db.Query(s, invitedUserId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userStatus []User
	for rows.Next() {
		var userState User
		err = rows.Scan(&userState.Id, &userState.Name, &userState.Email)
		if err != nil {
			return nil, err
		}
		userStatus = append(userStatus, userState)
	}
	return userStatus, nil
}

func UpdateUserInviteId(uid uuid.UUID, inviteId uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE users SET invited_id=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(inviteId, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateClientSecret(uid uuid.UUID, stripeClientSecret string) error {
	stmt, err := db.Prepare("UPDATE users SET stripe_client_secret=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(stripeClientSecret, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func ClearUserName(uid uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE users SET name=NULL WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)
	var res sql.Result
	res, err = stmt.Exec(uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}
