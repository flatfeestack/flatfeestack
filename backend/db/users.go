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

func (db *DB) FindAllEmails() ([]string, error) {
	rows, err := db.Query(`SELECT email FROM users`)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var emails []string
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

func (db *DB) FindUserByEmail(email string) (*UserDetail, error) {
	var u UserDetail
	err := db.QueryRow(`
		SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
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

func (db *DB) FindUserById(uid uuid.UUID) (*UserDetail, error) {
	var u UserDetail
	err := db.QueryRow(`
		SELECT id, stripe_id, invited_id, stripe_payment_method, stripe_last4, 
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

func (db *DB) FindPublicUserById(uid uuid.UUID) (*PublicUser, error) {
	var u PublicUser
	err := db.QueryRow(`
		SELECT id, name, image 
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

func (db *DB) InsertUser(user *UserDetail) error {
	_, err := db.Exec(
		`INSERT INTO users (id, email, name, created_at) VALUES ($1, $2, $3, $4)`,
		user.Id, user.Email, user.Name, user.CreatedAt)
	return err
}

func (db *DB) InsertFoundation(user *UserDetail) error {
	_, err := db.Exec(
		`INSERT INTO users (id, email, stripe_id, created_at, multiplier, multiplier_daily_limit) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.Id, user.Email, user.StripeId, user.CreatedAt, user.Multiplier, user.MultiplierDailyLimit)
	return err
}

func (db *DB) UpdateStripe(user *UserDetail) error {
	_, err := db.Exec(
		`UPDATE users 
		 SET stripe_id=$1, stripe_payment_method=$2, stripe_last4=$3
		 WHERE id=$4`,
		user.StripeId, user.PaymentMethod, user.Last4, user.Id)
	return err
}

func (db *DB) UpdateUserName(uid uuid.UUID, name string) error {
	_, err := db.Exec(
		`UPDATE users SET name=$1 WHERE id=$2`,
		name, uid)
	return err
}

func (db *DB) UpdateUserImage(uid uuid.UUID, data string) error {
	_, err := db.Exec(
		`UPDATE users SET image=$1 WHERE id=$2`,
		data, uid)
	return err
}

func (db *DB) DeleteUserImage(uid uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE users SET image=NULL WHERE id=$1`,
		uid)
	return err
}

func (db *DB) UpdateSeatsFreq(userId uuid.UUID, seats int, freq int) error {
	_, err := db.Exec(
		`UPDATE users SET seats=$1, freq=$2 WHERE id=$3`,
		seats, freq, userId)
	return err
}

func (db *DB) FindSponsoredUserBalances(invitedUserId uuid.UUID) ([]User, error) {
	rows, err := db.Query(`
		SELECT id, name, email
		FROM users 
		WHERE invited_id = $1`, invitedUserId)
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

func (db *DB) UpdateUserInviteId(uid uuid.UUID, inviteId uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE users SET invited_id=$1 WHERE id=$2`,
		inviteId, uid)
	return err
}

func (db *DB) UpdateClientSecret(uid uuid.UUID, stripeClientSecret string) error {
	_, err := db.Exec(
		`UPDATE users SET stripe_client_secret=$1 WHERE id=$2`,
		stripeClientSecret, uid)
	return err
}

func (db *DB) UpdateMultiplier(uid uuid.UUID, isSet bool) error {
	_, err := db.Exec(
		`UPDATE users SET multiplier=$1 WHERE id=$2`,
		isSet, uid)
	return err
}

func (db *DB) UpdateMultiplierDailyLimit(uid uuid.UUID, amount int64) error {
	_, err := db.Exec(
		`UPDATE users SET multiplier_daily_limit=$1 WHERE id=$2`,
		amount, uid)
	return err
}

func (db *DB) CheckDailyLimitStillAdheredTo(foundation *Foundation, amount *big.Int, currency string, yesterdayStart time.Time) (*big.Int, error) {
	if foundation == nil {
		return big.NewInt(0), fmt.Errorf("foundation cannot be nil")
	}

	multiplierDailyLimit := big.NewInt(int64(foundation.MultiplierDailyLimit))

	var balanceStr string
	err := db.QueryRow(`
		SELECT COALESCE(SUM(balance), 0)
		FROM daily_contribution
		WHERE user_sponsor_id = $1
		  AND foundation_payment = TRUE
		  AND currency = $2
		  AND day = $3`,
		foundation.Id, currency, yesterdayStart).Scan(&balanceStr)
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed to calculate daily contributions: %w", err)
	}

	dailySum, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return big.NewInt(0), fmt.Errorf("invalid balance: %v", balanceStr)
	}

	if new(big.Int).Add(dailySum, amount).Cmp(multiplierDailyLimit) <= 0 {
		return amount, nil
	}

	restrictedAmountToPay := new(big.Int).Sub(multiplierDailyLimit, dailySum)

	if restrictedAmountToPay.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(-1), nil
	}

	return restrictedAmountToPay, nil
}

func (db *DB) CheckFondsAmountEnough(foundation *Foundation, amount *big.Int, currency string) (*big.Int, error) {
	if foundation == nil {
		return big.NewInt(0), fmt.Errorf("foundation cannot be nil")
	}

	totalBalance, err := db.FindSumPaymentFromFoundation(foundation.Id, PayInSuccess, currency)
	if err != nil {
		return big.NewInt(0), err
	}

	totalSpending, err := db.FindSumDailySponsorsFromFoundationByCurrency(foundation.Id, currency)
	if err != nil {
		return big.NewInt(0), err
	}

	if new(big.Int).Add(totalSpending, amount).Cmp(totalBalance) <= 0 {
		return amount, nil
	}

	restrictedAmountToPay := new(big.Int).Sub(totalBalance, totalSpending)

	if restrictedAmountToPay.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(-1), nil
	}

	return restrictedAmountToPay, nil
}