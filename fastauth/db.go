package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type dbRes struct {
	sms          *string
	password     []byte
	meta         *string
	smsVerified  *int
	totpVerified *int
	refreshToken *string
	emailToken   *string
	inviteEmail  *string
	totp         *string
	errorCount   *int
}

type dbInvite struct {
	Email     string    `json:"email"`
	Pending   bool      `json:"pending"`
	CreatedAt time.Time `json:"createdAt"`
}

func dbSelect(email string) (*dbRes, error) {
	var res dbRes
	var pw string
	err := db.
		QueryRow(`SELECT 
					       sms, password, meta, refreshToken, emailToken, 
                           inviteEmail, totp, smsVerified, totpVerified, errorCount 
					     FROM auth WHERE email = $1`, email).
		Scan(&res.sms, &pw, &res.meta, &res.refreshToken, &res.emailToken,
			&res.inviteEmail, &res.totp, &res.smsVerified, &res.totpVerified, &res.errorCount)
	if err != nil {
		return nil, err
	}
	res.password, err = base32.StdEncoding.DecodeString(pw)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func dbInvitations(email string) ([]dbInvite, error) {
	var res []dbInvite
	query := "SELECT email, emailToken, created_at FROM auth WHERE inviteEmail=$1"
	rows, err := db.Query(query, email)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer rows.Close()
		for rows.Next() {
			var inv dbInvite
			var token *string
			err = rows.Scan(&inv.Email, &token, &inv.CreatedAt)
			inv.Pending = token != nil
			if err != nil {
				return nil, err
			}
			res = append(res, inv)
		}
		return res, nil
	default:
		return nil, err
	}
}

func delInvite(inviteEmail string, email string) error {
	stmt, err := db.Prepare("DELETE from auth WHERE email = $1 AND inviteEmail = $2 AND emailToken IS NOT NULL")
	if err != nil {
		return fmt.Errorf("prepare DELETE auth pending status for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, inviteEmail)
	err = handleErr(res, err, "DELETE auth errorCount", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "DEL_INVITE")
}

func insertUser(email string, pwRaw []byte, meta *string, emailToken string, refreshToken string, inviteEmail *string) error {
	pw := base32.StdEncoding.EncodeToString(pwRaw)
	stmt, err := db.Prepare("INSERT INTO auth (email, password, meta, emailToken, refreshToken, inviteEmail) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO auth for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, pw, meta, emailToken, refreshToken, inviteEmail)
	err = handleErr(res, err, "INSERT INTO auth", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "INSERT")
}

func updateRefreshToken(email string, oldRefreshToken string, newRefreshToken string) error {
	stmt, err := db.Prepare("UPDATE auth SET refreshToken = $1 WHERE refreshToken = $2 and email=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE refreshTokenfor statement failed: %v", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(newRefreshToken, oldRefreshToken, email)
	err = handleErr(res, err, "UPDATE refreshToken", "n/a")
	if err != nil {
		return err
	}
	return insertAudit(email, "UPDATE_REFRESH")
}

func resetPasswordInvite(email string, emailToken string, newPw []byte) error {
	pw := base32.StdEncoding.EncodeToString(newPw)
	stmt, err := db.Prepare("UPDATE auth SET password = $1, emailToken = NULL WHERE email = $2 AND emailToken = $3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth password for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(pw, email, emailToken)
	err = handleErr(res, err, "UPDATE auth password invite", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "RESET_INVITE")
}

func resetPassword(email string, forgetEmailToken string, newPw []byte) error {
	pw := base32.StdEncoding.EncodeToString(newPw)
	stmt, err := db.Prepare("UPDATE auth SET password = $1, totp = NULL, sms = NULL, forgetEmailToken = NULL WHERE email = $2 AND forgetEmailToken = $3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth password for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(pw, email, forgetEmailToken)
	err = handleErr(res, err, "UPDATE auth password", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "RESET")
}

func updateEmailForgotToken(email string, token string) error {
	//TODO: don't accept too old forget tokens
	stmt, err := db.Prepare("UPDATE auth SET forgetEmailToken = $1 WHERE email = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth forgetEmailToken for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(token, email)
	err = handleErr(res, err, "UPDATE auth forgetEmailToken", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "FORGOT")
}

func updateTOTP(email string, totp string) error {
	stmt, err := db.Prepare("UPDATE auth SET totp = $1 WHERE email = $2 and totp IS NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth totp for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(totp, email)
	err = handleErr(res, err, "UPDATE auth totp", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "FORGOT")
}

func updateSMS(email string, totp string, sms string) error {
	stmt, err := db.Prepare("UPDATE auth SET totp = $1, sms = $2 WHERE email = $3 AND smsVerified IS NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth totp for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(totp, sms, email)
	err = handleErr(res, err, "UPDATE auth totp", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "SMS")
}

func updateEmailToken(email string, token string) error {
	stmt, err := db.Prepare("UPDATE auth SET emailToken = NULL WHERE email = $1 AND emailToken = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, token)
	err = handleErr(res, err, "UPDATE auth", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "TOKEN")
}

func updateSMSVerified(email string) error {
	stmt, err := db.Prepare("UPDATE auth SET smsVerified = 1 WHERE email = $1 AND sms IS NOT NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email)
	err = handleErr(res, err, "UPDATE auth SMS timestamp", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "SMS_VERIFIED")
}

func updateTOTPVerified(email string) error {
	stmt, err := db.Prepare("UPDATE auth SET totpVerified = 1 WHERE email = $1 AND totp IS NOT NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email)
	err = handleErr(res, err, "UPDATE auth totp timestamp", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "TOTP_VERIFIED")
}

func incErrorCount(email string) error {
	stmt, err := db.Prepare("UPDATE auth set errorCount = errorCount + 1 WHERE email = $1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth status for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email)
	err = handleErr(res, err, "UPDATE auth errorCount", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "INC_COUNTER")
}

func resetCount(email string) error {
	stmt, err := db.Prepare("UPDATE auth set errorCount = 0 WHERE email = $1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth status for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email)
	err = handleErr(res, err, "UPDATE auth status", email)
	if err != nil {
		return err
	}
	return insertAudit(email, "RESET_COUNTER")
}

func insertAudit(email string, action string) error {
	stmt, err := db.Prepare("INSERT INTO audit (email, action, created_at) VALUES ($1, $2, CURRENT_TIMESTAMP)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO audit for %v statement failed: %v", email, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, action)
	return handleErr(res, err, "INSERT INTO audit", email)
}

func handleErr(res sql.Result, err error, info string, email string) error {
	if err != nil {
		return fmt.Errorf("%v query %v failed: %v", info, email, err)
	}
	nr, err := res.RowsAffected()
	if nr == 0 || err != nil {
		return fmt.Errorf("%v %v rows %v, affected or err: %v", info, nr, email, err)
	}
	return nil
}

///////// Setup

func addInitialUserWithMeta(username string, password string, meta *string) error {
	res, err := dbSelect(username)
	if res == nil || err != nil {
		dk, err := newPw(password, 0)
		if err != nil {
			return err
		}
		err = insertUser(username, dk, meta, "emailToken", "refreshToken", nil)
		if err != nil {
			return err
		}
		err = updateEmailToken(username, "emailToken")
		if err != nil {
			return err
		}
	}
	return nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open(opts.DBDriver, opts.DBPath)
	if err != nil {
		return nil, err
	}

	//this will create or alter tables
	//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	for _, v := range strings.Split(opts.DBScripts, ":") {
		if _, err := os.Stat(v); err == nil {
			file, err := ioutil.ReadFile(v)
			if err != nil {
				return nil, err
			}
			requests := strings.Split(string(file), ";")
			for _, request := range requests {
				request = strings.Replace(request, "\n", "", -1)
				request = strings.Replace(request, "\t", "", -1)
				if !strings.HasPrefix(request, "#") {
					_, err = db.Exec(request)
					if err != nil {
						return nil, fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		}
	}

	return db, nil
}

func setupDB() {
	if opts.Users != "" {
		//add user for development
		users := strings.Split(opts.Users, ";")
		for _, user := range users {
			userPwMeta := strings.Split(user, ":")
			if len(userPwMeta) == 2 {
				err := addInitialUserWithMeta(userPwMeta[0], userPwMeta[1], nil)
				if err == nil {
					log.Printf("insterted user %v", userPwMeta[0])
				} else {
					log.Printf("could not insert %v: %v", userPwMeta[0], err)
				}
			} else if len(userPwMeta) == 3 {
				meta := userPwMeta[2]
				err := addInitialUserWithMeta(userPwMeta[0], userPwMeta[1], &meta)
				if err == nil {
					log.Printf("insterted user %v", userPwMeta[0])
				} else {
					log.Printf("could not insert %v: %v", userPwMeta[0], err)
				}
			} else {
				log.Printf("username and password need to be seperated by ':'")
			}
		}
	}
}
