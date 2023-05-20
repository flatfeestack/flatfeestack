package main

import (
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type dbRes struct {
	email        string
	password     []byte
	refreshToken string
	emailToken   *string
	sms          *string
	smsVerified  *time.Time
	totp         *string
	totpVerified *time.Time
	errorCount   int
	flowType     string
	metaSystem   *string
}

func findAuthByEmail(email string) (*dbRes, error) {
	var res dbRes
	var pw string
	query := `SELECT sms, password, meta_system, refresh_token, 
       				 email_token, totp, sms_verified, totp_verified, error_count, flow_type
              FROM auth WHERE email = $1`
	err := dbLib.DB.QueryRow(query, email).Scan(
		&res.sms, &pw, &res.metaSystem, &res.refreshToken, &res.emailToken,
		&res.totp, &res.smsVerified, &res.totpVerified, &res.errorCount, &res.flowType)

	if err != nil {
		return nil, err
	}
	res.password, err = base32.StdEncoding.DecodeString(pw)
	if err != nil {
		return nil, err
	}
	res.email = email
	return &res, nil
}

func insertUser(email string, pwRaw []byte, emailToken string, refreshToken string, flowType string, now time.Time) error {
	var pw *string
	if pwRaw != nil {
		s1 := base32.StdEncoding.EncodeToString(pwRaw)
		pw = &s1
	}
	stmt, err := dbLib.DB.Prepare(`INSERT INTO auth as a (email, password, email_token, refresh_token, flow_type, created_at) 
								   VALUES ($1, $2, $3, $4, $5, $6)
								   ON CONFLICT (email) DO
								     UPDATE SET password=$2, email_token = $3
								     WHERE a.email_token IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO auth for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(email, pw, emailToken, refreshToken, flowType, now)
	return handleErr(res, err, "INSERT INTO auth", email)
}

func deleteDbUser(email string) error {
	stmt, err := dbLib.DB.Prepare(`DELETE FROM auth where email = $1`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO auth for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(email, email)
	return handleErr(res, err, "DELETE FROM auth", email)
}

func updateRefreshToken(oldRefreshToken string, newRefreshToken string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET refresh_token = $1 WHERE refresh_token = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE refreshTokenfor statement failed: %v", err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(newRefreshToken, oldRefreshToken)
	return handleErrEffect(res, err, "UPDATE refreshToken", "n/a", false)
}

func updateSystemMeta(email string, systemMeta string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET meta_system = $1 WHERE email=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE meta_system statement failed: %v", err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(systemMeta, email)
	return handleErr(res, err, "UPDATE meta_system", "n/a")
}

func updatePasswordInvite(email string, emailToken string, newPw []byte) error {
	pw := base32.StdEncoding.EncodeToString(newPw)
	stmt, err := dbLib.DB.Prepare(`UPDATE auth SET password = $1, email_token = NULL 
								   WHERE email = $2 AND email_token = $3 AND password IS NULL`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth password for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(pw, email, emailToken)
	return handleErr(res, err, "UPDATE auth password invite", email)
}

func updatePasswordForgot(email string, forgetEmailToken string, newPw []byte) error {
	pw := base32.StdEncoding.EncodeToString(newPw)
	stmt, err := dbLib.DB.Prepare(`UPDATE auth SET password = $1, totp = NULL, sms = NULL, forget_email_token = NULL 
								   WHERE email = $2 AND forget_email_token = $3`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth password for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(pw, email, forgetEmailToken)
	return handleErr(res, err, "UPDATE auth password", email)
}

func updateEmailForgotToken(email string, token string) error {
	//TODO: don't accept too old forget tokens
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET forget_email_token = $1 WHERE email = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth forgetEmailToken for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(token, email)
	return handleErr(res, err, "UPDATE auth forgetEmailToken", email)
}

func updateTOTP(email string, totp string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET totp = $1 WHERE email = $2 and totp IS NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth totp for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(totp, email)
	return handleErr(res, err, "UPDATE auth totp", email)
}

func updateSMS(email string, totp string, sms string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET totp = $1, sms = $2 WHERE email = $3 AND sms_verified IS NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth totp for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(totp, sms, email)
	return handleErr(res, err, "UPDATE auth totp", email)
}

func updateEmailToken(email string, token string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET email_token = NULL WHERE email = $1 AND email_token = $2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(email, token)
	return handleErr(res, err, "UPDATE auth", email)
}

func updateSMSVerified(email string, now time.Time) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET sms_verified = $1 WHERE email = $2 AND sms IS NOT NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(now, email)
	return handleErr(res, err, "UPDATE auth SMS timestamp", email)
}

func updateTOTPVerified(email string, now time.Time) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth SET totp_verified = $1 WHERE email = $2 AND totp IS NOT NULL")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(now, email)
	return handleErr(res, err, "UPDATE auth totp timestamp", email)
}

func incErrorCount(email string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth set error_count = error_count + 1 WHERE email = $1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth status for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(email)
	return handleErr(res, err, "UPDATE auth errorCount", email)
}

func resetCount(email string) error {
	stmt, err := dbLib.DB.Prepare("UPDATE auth set error_count = 0 WHERE email = $1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE auth status for %v statement failed: %v", email, err)
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(email)
	return handleErr(res, err, "UPDATE auth status", email)
}

func handleErr(res sql.Result, err error, info string, email string) error {
	return handleErrEffect(res, err, info, email, true)
}

func handleErrEffect(res sql.Result, err error, info string, email string, mustHaveEffect bool) error {
	if err != nil {
		return fmt.Errorf("%v query %v failed: %v", info, email, err)
	}
	nr, err := res.RowsAffected()
	if mustHaveEffect && nr == 0 {
		return fmt.Errorf("%v rows %v, affected: %v", info, nr, email)
	}
	if err != nil {
		return fmt.Errorf("%v %v err: %v", info, email, err)
	}
	return nil
}

// /////// Setup
// Meta data can be additional information that will be encoded in the JWT token
func addInitialUserWithMeta(username string, password string, metaSystem *string) error {
	res, err := findAuthByEmail(username)
	if res == nil || err != nil {
		dk, err := newPw(password, 0)
		if err != nil {
			return err
		}
		err = insertUser(username, dk, "emailToken", "refreshToken", "sys", timeNow())
		if err != nil {
			return err
		}
		err = updateEmailToken(username, "emailToken")
		if err != nil {
			return err
		}
		if metaSystem != nil {
			if !json.Valid([]byte(*metaSystem)) {
				return fmt.Errorf("not valid json: %v", *metaSystem)
			}
			err = updateSystemMeta(username, *metaSystem)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addInitialUsers() {
	if opts.Users != "" {
		//add user for development
		users := strings.Split(opts.Users, ";")
		for _, user := range users {
			userPwMeta := strings.SplitN(user, ":", 3)

			var metaSystem *string
			if len(userPwMeta) >= 3 {
				metaSystem = &userPwMeta[2]
			}

			if len(userPwMeta) >= 2 {
				err := addInitialUserWithMeta(userPwMeta[0], userPwMeta[1], metaSystem)
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
