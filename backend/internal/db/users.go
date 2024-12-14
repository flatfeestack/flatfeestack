package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
)

type PublicUser struct {
	Id    uuid.UUID `json:"id"`
	Name  *string   `json:"name,omitempty"`
	Image *string   `json:"image"`
}

type User struct {
	Id        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time
}

type Foundation struct {
	Id                   uuid.UUID   `json:"id"`
	MultiplierDailyLimit int         `json:"multiplierDailyLimit"`
	RepoIds              []uuid.UUID `json:"repoIds"`
}

type UserDetail struct {
	User
	InvitedId            *uuid.UUID
	StripeId             *string `json:"-"`
	Image                *string `json:"image"`
	PaymentMethod        *string `json:"paymentMethod"`
	StripeClientSecret   *string `json:"clientSecret"`
	Last4                *string `json:"last4"`
	Seats                int     `json:"seats"`
	Freq                 int     `json:"freq"`
	Claims               *jwt.Claims
	Role                 string `json:"role,omitempty"`
	Multiplier           bool   `json:"multiplier"`
	MultiplierDailyLimit int    `json:"multiplierDailyLimit"`
}

func FindAllEmails() ([]string, error) {
	emails := []string{}
	s := `SELECT email from users`
	rows, err := DB.Query(s)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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
	err := DB.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
                               email, name, image, seats, freq, created_at, multiplier, multiplier_daily_limit
                        FROM users 
                        WHERE email=$1`, email).
		Scan(&u.Id, &u.StripeId, &u.InvitedId, &u.PaymentMethod, &u.Last4,
			&u.Email, &u.Name, &u.Image, &u.Seats, &u.Freq, &u.CreatedAt,
			&u.Multiplier, &u.MultiplierDailyLimit)
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
	err := DB.
		QueryRow(`SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
                               stripe_client_secret, email, name, image, seats, freq, created_at,
							   multiplier, multiplier_daily_limit                        
                        FROM users 
                        WHERE id=$1`, uid).
		Scan(&u.Id, &u.StripeId, &u.InvitedId, &u.PaymentMethod, &u.Last4,
			&u.StripeClientSecret, &u.Email, &u.Name, &u.Image, &u.Seats, &u.Freq, &u.CreatedAt,
			&u.Multiplier, &u.MultiplierDailyLimit)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func FindPublicUserById(uid uuid.UUID) (*PublicUser, error) {
	var u PublicUser
	err := DB.
		QueryRow(`SELECT id, name, image 
                        FROM users 
                        WHERE id=$1`, uid).
		Scan(&u.Id, &u.Name, &u.Image)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

/*func GetUserThatSponsoredTrustedRepo() ([]User, error) {
	s := `SELECT
				DISTINCT user_sponsor_id AS user_id
			FROM
				daily_contribution
			WHERE
				created_at >= CURRENT_DATE - INTERVAL '1 month'`
	rows, err := DB.Query(s)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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
}*/

func InsertUser(user *UserDetail) error {
	stmt, err := DB.Prepare("INSERT INTO users (id, email, name, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(user.Id, user.Email, user.Name, user.CreatedAt)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func InsertFoundation(user *UserDetail) error {
	stmt, err := DB.Prepare("INSERT INTO users (id, email, stripe_id, created_at, multiplier, multiplier_daily_limit) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(user.Id, user.Email, user.StripeId, user.CreatedAt, user.Multiplier, user.MultiplierDailyLimit)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func UpdateStripe(user *UserDetail) error {
	stmt, err := DB.Prepare(`UPDATE users 
                                   SET stripe_id=$1, stripe_payment_method=$2, stripe_last4=$3
                                   WHERE id=$4`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(user.StripeId, user.PaymentMethod, user.Last4, user.Id)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateUserName(uid uuid.UUID, name string) error {
	stmt, err := DB.Prepare("UPDATE users SET name=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(name, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateUserImage(uid uuid.UUID, data string) error {
	stmt, err := DB.Prepare("UPDATE users SET image=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(data, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func DeleteUserImage(uid uuid.UUID) error {
	stmt, err := DB.Prepare("UPDATE users SET image=NULL WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateSeatsFreq(userId uuid.UUID, seats int, freq int) error {
	stmt, err := DB.Prepare("UPDATE users SET seats=$1, freq=$2 WHERE id=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", userId, err)
	}
	defer CloseAndLog(stmt)

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
	rows, err := DB.Query(s, invitedUserId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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
	stmt, err := DB.Prepare("UPDATE users SET invited_id=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(inviteId, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateClientSecret(uid uuid.UUID, stripeClientSecret string) error {
	stmt, err := DB.Prepare("UPDATE users SET stripe_client_secret=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(stripeClientSecret, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func ClearUserName(uid uuid.UUID) error {
	stmt, err := DB.Prepare("UPDATE users SET name=NULL WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)
	var res sql.Result
	res, err = stmt.Exec(uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateMultiplier(uid uuid.UUID, isSet bool) error {
	stmt, err := DB.Prepare("UPDATE users SET multiplier=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(isSet, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func UpdateMultiplierDailyLimit(uid uuid.UUID, amount int64) error {
	stmt, err := DB.Prepare("UPDATE users SET multiplier_daily_limit=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(amount, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func CheckDailyLimitStillAdheredTo(foundation *Foundation, amount *big.Int, currency string, yesterdayStart time.Time) (bool, error) {
	if foundation == nil {
		return false, fmt.Errorf("foundation cannot be nil")
	}

	multiplierDailyLimit := big.NewInt(int64(foundation.MultiplierDailyLimit))

	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM daily_contribution
		WHERE
			user_sponsor_id = $1
			AND foundation_payment
			AND currency = $2
			AND day = $3`

	var dailySumInt int64

	err := DB.QueryRow(query, foundation.Id, currency, yesterdayStart).Scan(&dailySumInt)
	if err != nil {
		if err == sql.ErrNoRows {
			if amount.Cmp(multiplierDailyLimit) <= 0 {
				return true, nil
			}
			return false, nil
		}
		return false, fmt.Errorf("failed to calculate daily contributions: %v", err)
	}

	dailySum := big.NewInt(dailySumInt)

	if new(big.Int).Add(dailySum, amount).Cmp(multiplierDailyLimit) <= 0 {
		return true, nil
	}

	return false, nil
}

func CheckFondsAmountEnough(foundation *Foundation, amount *big.Int, currency string) (bool, error) {
	if foundation == nil {
		return false, fmt.Errorf("foundation cannot be nil")
	}

	totalBalance, err := FindSumPaymentFromFoundation(foundation.Id, PayInSuccess, currency)
	if err != nil {
		return false, err
	}

	totalSpending, err := FindSumDailySponsorsFromFoundationByCurrency(foundation.Id, currency)
	if err != nil {
		return false, err
	}

	if new(big.Int).Add(totalSpending, amount).Cmp(totalBalance) <= 0 {
		return true, nil
	}

	return false, nil
}
