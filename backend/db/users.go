package db

import (
	"database/sql"
	"fmt"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id                uuid.UUID `json:"id" sql:",type:uuid"`
	InvitedId         *uuid.UUID
	StripeId          *string    `json:"-"`
	PaymentCycleInId  *uuid.UUID `json:"paymentCycleInId,omitempty"`
	PaymentCycleOutId uuid.UUID  `json:"paymentCycleOutId"`
	Email             string     `json:"email"`
	Name              *string    `json:"name"`
	Image             *string    `json:"image"`
	PaymentMethod     *string    `json:"paymentMethod"`
	Last4             *string    `json:"last4"`
	CreatedAt         time.Time
	Claims            *jwt.Claims
	Role              *string `json:"role,omitempty"`
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

func FindUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, payment_cycle_in_id,
                                payment_cycle_out_id, stripe_last4, email, name, image 
                         FROM users WHERE email=$1`, email).
		Scan(&u.Id, &u.StripeId, &u.InvitedId,
			&u.PaymentMethod, &u.PaymentCycleInId, &u.PaymentCycleOutId, &u.Last4, &u.Email, &u.Name,
			&u.Image)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func FindUserById(uid uuid.UUID) (*User, error) {
	var u User
	err := db.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, payment_cycle_in_id,
                                payment_cycle_out_id, stripe_last4, email, name, image 
                         FROM users WHERE id=$1`, uid).
		Scan(&u.Id, &u.StripeId, &u.InvitedId,
			&u.PaymentMethod, &u.PaymentCycleInId, &u.PaymentCycleOutId, &u.Last4, &u.Email, &u.Name,
			&u.Image)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func InsertUser(user *User) error {
	stmt, err := db.Prepare("INSERT INTO payment_cycle_out (id, created_at) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(user.PaymentCycleOutId, user.CreatedAt)
	if err != nil {
		return err
	}
	err = handleErrMustInsertOne(res)
	if err != nil {
		return err
	}

	stmt, err = db.Prepare("INSERT INTO users (id, email, stripe_id, payment_cycle_out_id, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	res, err = stmt.Exec(user.Id, user.Email, user.StripeId, user.PaymentCycleOutId, user.CreatedAt)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func UpdateStripe(user *User) error {
	stmt, err := db.Prepare(`UPDATE users SET 
                                           stripe_id=$1,  
                                           stripe_payment_method=$2, 
                                           stripe_last4=$3
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

func UpdatePaymentCycleInId(userId uuid.UUID, paymentCycleInId uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE users SET payment_cycle_in_id=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(paymentCycleInId, userId)
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

func FindSponsoredUserBalances(userId uuid.UUID) ([]UserStatus, error) {
	s := `SELECT u.id, u.name, u.email
          FROM users u
          INNER JOIN payment_cycle_in p ON p.id = u.payment_cycle_in_id
          WHERE u.invited_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userStatus []UserStatus
	for rows.Next() {
		var userState UserStatus
		err = rows.Scan(&userState.UserId, &userState.Name, &userState.Email, &userState.DaysLeft)
		if err != nil {
			return nil, err
		}
		userStatus = append(userStatus, userState)
	}
	return userStatus, nil
}

// for testing
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
