package api

import (
	"backend/internal/db"
	"backend/pkg/config"
	"backend/pkg/util"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PayoutInfoDTO struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

type UserBalanceDto struct {
	UserId           uuid.UUID  `json:"userId"`
	Balance          *big.Int   `json:"balance"`
	PaymentCycleInId *uuid.UUID `json:"paymentCycleInId"`
	BalanceType      string     `json:"balanceType"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type UserBalances struct {
	PaymentCycle db.PaymentCycle     `json:"paymentCycle"`
	UserBalances []UserBalanceDto    `json:"userBalances"`
	Total        map[string]*big.Int `json:"total"`
	DaysLeft     int64               `json:"daysLeft"`
}

type TotalUserBalance struct {
	Currency string   `json:"currency"`
	Balance  *big.Int `json:"balance"`
}

const (
	UserBalancesError = "Oops something went wrong with retrieving user balances. Please try again."
	PaymentError      = "Oops something went wrong with the payment. Please try again."
	PayoutError       = "Oops something went wrong with the payout. Please try again."
)

type ResourceHandler struct {
	Config *config.Config
}

func NewResourceHandler(cfg *config.Config) *ResourceHandler {
	return &ResourceHandler{Config: cfg}
}

func FakePayment(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	slog.Info("fake payment")

	emailEsc := r.PathValue("email")
	email, err := url.QueryUnescape(emailEsc)

	if err != nil {
		slog.Error("Query unescape fake payment email",
			slog.String("email", emailEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}

	seatsEsc := r.PathValue("seats")
	seatsStr, err := url.QueryUnescape(seatsEsc)

	if err != nil {
		slog.Error("Query unescape fake payment seats",
			slog.String("seats", seatsStr),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}

	u, err := db.FindUserByEmail(email)
	if err != nil {
		slog.Error("Error while finding user by email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	seats, err := strconv.ParseInt(seatsStr, 10, 64)
	if err != nil {
		slog.Error("Error while parsing int",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	e := uuid.New()
	payInEvent := db.PayInEvent{
		Id:         uuid.New(),
		ExternalId: e,
		UserId:     u.Id,
		Balance:    big.NewInt(Plans[1].PriceBase),
		Currency:   "USD",
		Status:     db.PayInRequest,
		Seats:      seats,
		Freq:       365,
		CreatedAt:  time.Time{},
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		slog.Error("Error with inserting pay in event",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	err = db.PaymentSuccess(e, big.NewInt(1))
	if err != nil {
		slog.Error("Error with payment success",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	return
}

func StrategyDeductMax(userId uuid.UUID, balances map[string]*big.Int, subs map[string]*big.Int, futSub map[string]*big.Int) (string, int64, *big.Int, error) {
	var maxDailySpending *big.Int
	var maxFreq = int64(0)
	var maxCurrency string

	//
	for currency, newBalance := range balances {
		if subs[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, subs[currency])
		}
		if futSub[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, futSub[currency])
		}

		ds, _, _, _, err := db.FindLatestDailyPayment(userId, currency)
		if err != nil {
			return "N/A", 0, nil, err
		}
		freq := new(big.Int).Div(newBalance, ds).Int64()
		if freq > 0 {
			if freq > maxFreq {
				maxFreq = freq
				maxDailySpending = ds
				maxCurrency = currency
			}
		}
	}

	if maxDailySpending == nil {
		return "N/A", 0, nil, nil
	}
	return maxCurrency, maxFreq, maxDailySpending, nil
}

func CancelSub(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	err := db.UpdateSeatsFreq(user.Id, user.Seats, 0)
	if err != nil {
		slog.Error("Could not cancel subscription",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong while canceling the subscription. Please try again.")
		return
	}
}

func StatusSponsoredUsers(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	userStatus, err := db.FindSponsoredUserBalances(user.Id)
	if err != nil {
		slog.Error("Error while finding sponsored user balances",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	err = json.NewEncoder(w).Encode(userStatus)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func (h *ResourceHandler) RequestPayout(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	targetCurrencyEsc := r.PathValue("targetCurrency")
	targetCurrency, err := url.QueryUnescape(targetCurrencyEsc)
	if err != nil {
		slog.Error("Query unescape invite-by email",
			slog.String("email", targetCurrencyEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}

	ownContributionIds, err := db.FindOwnContributionIds(user.Id, targetCurrency)
	if err != nil {
		slog.Error("Error while trying to retrieve own contribution ids",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	}

	// notabene: For USDC, 10^6 units are one dollar
	// See explorer https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
	// FlatFeeStack already calculates in micro dollars
	totalEarnedAmount, err := db.SumTotalEarnedAmountForContributionIds(ownContributionIds)
	if err != nil {
		slog.Error("Unable to retrieve already earned amount in target currency",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	}

	signature, err := SignETH(h.Config.ETHPrivateKey, h.Config.ETHContractAddress, user.Id, totalEarnedAmount)
	util.WriteJson(w, signature)
}

func PaymentEvent(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	payInEvents, err := db.FindPayInUser(user.Id)
	if err != nil {
		slog.Error("Error while finding pay in user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	err = json.NewEncoder(w).Encode(payInEvents)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func UserBalance(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	mAdd, err := db.FindSumPaymentByCurrency(user.Id, db.PayInSuccess)
	if err != nil {
		slog.Error("Error while finding sum payments by currency",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	//either the user spent it on a repo that does not have any devs who can claim
	mFut, err := db.FindSumFutureSponsors(user.Id)
	if err != nil {
		slog.Error("Error while finding sum future sponsors",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}
	//or the user spent it on for a repo with a dev who can claim
	mSub, err := db.FindSumDailySponsors(user.Id)
	if err != nil {
		slog.Error("Error while finding sum daily sponsors",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	totalUserBalances := []TotalUserBalance{}
	for currency, _ := range mAdd {
		if mSub[currency] != nil {
			mAdd[currency] = new(big.Int).Sub(mAdd[currency], mSub[currency])
			mSub[currency] = big.NewInt(0)
		}
		if mFut[currency] != nil {
			mAdd[currency] = new(big.Int).Sub(mAdd[currency], mFut[currency])
			mFut[currency] = big.NewInt(0)
		}
		//
		totalUserBalances = append(totalUserBalances, TotalUserBalance{
			Currency: currency,
			Balance:  mAdd[currency],
		})
	}

	//now sanity check if all the deducted mSub/mFut are set to zero
	for currency, balance := range mSub {
		if balance.Cmp(big.NewInt(0)) != 0 {
			slog.Error("Something is off with",
				slog.String("currency", currency))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}
	for currency, balance := range mFut {
		if balance.Cmp(big.NewInt(0)) != 0 {
			slog.Error("Something is off with",
				slog.String("currency", currency))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}

	err = json.NewEncoder(w).Encode(totalUserBalances)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

}

func SignNeo(privateKeyHex string, userId uuid.UUID, amount *big.Int) (string, error) {
	privateKey, err := keys.NewPrivateKeyFromWIF(privateKeyHex)
	if err != nil {
		return "", err
	}

	ownerIdBytes, _ := userId.MarshalBinary()
	teaArray := amount.Bytes()
	for i := 0; i < len(teaArray)/2; i++ {
		opp := len(teaArray) - 1 - i
		teaArray[i], teaArray[opp] = teaArray[opp], teaArray[i]
	}
	message := append(ownerIdBytes, teaArray...)
	signature := privateKey.Sign(message)

	return hex.EncodeToString(signature), nil
}

func SignETH(privateKeyHex string, contractAddress string, userId uuid.UUID, amount *big.Int) (string, error) {
	var arguments abi.Arguments
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.AddressTy},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.StringTy},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.UintTy, Size: 256},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.StringTy},
	})
	arguments = append(arguments, abi.Argument{
		Type: abi.Type{T: abi.UintTy, Size: 256},
	})

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}

	encodedUserId := [32]byte(crypto.Keccak256([]byte(userId.String())))
	packed, err := arguments.Pack(contractAddress, "calculateWithdraw", encodedUserId, "#", amount)
	hashRaw := crypto.Keccak256(packed)

	// Add Ethereum Signed Message prefix to hash
	prefix := []byte("\x19Ethereum Signed Message:\n32")
	prefixedHash := crypto.Keccak256(append(prefix, hashRaw[:]...))

	signature, err := crypto.Sign(prefixedHash[:], privateKey)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}
