package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ethContractBalanceMetric  prometheus.Gauge
	usdcContractBalanceMetric prometheus.Gauge
)

func InitMetricsGauges(registry *prometheus.Registry) {
	setupEthContractBalanceGauge(registry)
	setupUsdcContractBalanceGauge(registry)
}

func setupEthContractBalanceGauge(registry *prometheus.Registry) {
	ethContractBalanceMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "flatfeestack",
		Subsystem: "eth_payout",
		Name:      "remaining_balance",
		Help:      "Remaining ETH available on payout contract",
	})

	registry.MustRegister(ethContractBalanceMetric)
}

func setupUsdcContractBalanceGauge(registry *prometheus.Registry) {
	usdcContractBalanceMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "flatfeestack",
		Subsystem: "usdc_payout",
		Name:      "remaining_balance",
		Help:      "Remaining USDC available on payout contract",
	})

	registry.MustRegister(usdcContractBalanceMetric)
}
