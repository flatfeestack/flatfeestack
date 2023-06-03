package metrics

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"math/big"
	"payout/contracts"
	"time"
)

var (
	ethContractAddress  common.Address
	ethConnection       *ethclient.Client
	usdcContractAddress common.Address
)

func InitMetricsCron(ethConnection2 *ethclient.Client, usdcContractAddress2 string, ethContractAddress2 string) {
	ethContractAddress = common.HexToAddress(ethContractAddress2)
	ethConnection = ethConnection2
	usdcContractAddress = common.HexToAddress(usdcContractAddress2)

	cron()
}

func cron() {
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every("5m").Do(checkEthBalance)
	if err != nil {
		log.Errorf("error when executing ETH balance check: %s", err)
	}

	_, err = s.Every("5m").Do(checkUsdcBalance)
	if err != nil {
		log.Errorf("error when executing USDC balance check: %s", err)
	}

	s.StartAsync()
}

func checkEthBalance() {
	balanceFloat64, err := retrieveBalanceFromContract(ethContractAddress)
	if err == nil {
		log.Printf("Successfully retrieved contract balance for ETH: %f", balanceFloat64)
		ethContractBalanceMetric.Set(balanceFloat64)
	} else {
		log.Errorf("Failed to retrieve ETH balance: %v", err)
		ethContractBalanceMetric.Set(0)
	}
}

func checkUsdcBalance() {
	balanceFloat64, err := retrieveBalanceFromContract(usdcContractAddress)
	if err == nil {
		log.Printf("Successfully retrieved contract balance for USDC: %f", balanceFloat64)
		usdcContractBalanceMetric.Set(balanceFloat64)
	} else {
		log.Errorf("Failed to retrieve USDC balance: %v", err)
		usdcContractBalanceMetric.Set(0)
	}
}

func retrieveBalanceFromContract(contractAddress common.Address) (float64, error) {
	payoutBase, err := contracts.NewPayoutBase(contractAddress, ethConnection)
	if err != nil {
		return 0, err
	}

	balanceInt, err := payoutBase.GetContractBalance(nil)
	if err != nil {
		return 0, err
	}

	balanceBigFloat := new(big.Float).SetInt(balanceInt)
	usdcDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
	balanceBigFloat.Quo(balanceBigFloat, new(big.Float).SetInt(usdcDecimals))
	balanceFloat64, _ := balanceBigFloat.Float64()

	return balanceFloat64, nil
}
