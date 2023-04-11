package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type Wallet struct {
	Id       uuid.UUID `json:"id"`
	Currency string    `json:"currency"`
	Address  string    `json:"address"`
}

func FindActiveWalletsByUserId(uid uuid.UUID) ([]Wallet, error) {
	userWallets := []Wallet{}
	s := "SELECT id, currency, address FROM wallet_address WHERE user_id=$1 AND is_deleted = false"
	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var userWallet Wallet
		err = rows.Scan(&userWallet.Id, &userWallet.Currency, &userWallet.Address)
		if err != nil {
			return nil, err
		}
		userWallets = append(userWallets, userWallet)
	}
	return userWallets, nil
}

func FindAllWalletsByUserId(uid uuid.UUID) ([]Wallet, error) {
	userWallets := []Wallet{}
	s := "SELECT id, currency, address FROM wallet_address WHERE user_id=$1"
	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var userWallet Wallet
		err = rows.Scan(&userWallet.Id, &userWallet.Currency, &userWallet.Address)
		if err != nil {
			return nil, err
		}
		userWallets = append(userWallets, userWallet)
	}
	return userWallets, nil
}

func InsertWallet(uid uuid.UUID, currency string, address string, isDeleted bool) (*uuid.UUID, error) {
	stmt, err := db.Prepare("INSERT INTO wallet_address(user_id, currency, address, is_deleted) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return nil, fmt.Errorf("prepare INSERT INTO wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(uid, currency, address, isDeleted).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil
}

func UpdateWallet(uid uuid.UUID, isDeleted bool) error {
	stmt, err := db.Prepare("UPDATE wallet_address set is_deleted = $2 WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid, isDeleted)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}
