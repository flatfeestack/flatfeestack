package config

type Config struct {
	Port                      int
	HS256                     string
	JwtKey                    []byte
	Env                       string
	StripeAPISecretKey        string
	StripeAPIPublicKey        string
	StripeWebhookSecretKey    string
	DBPath                    string
	DBDriver                  string
	DBScripts                 string
	Admins                    string
	EmailLinkPrefix           string
	EmailFrom                 string
	EmailFromName             string
	EmailUrl                  string
	EmailToken                string
	EmailMarketing            string
	NowpaymentsToken          string
	NowpaymentsIpnKey         string
	NowpaymentsApiUrl         string
	NowpaymentsIpnCallbackUrl string
	BackendUsername           string
	BackendPassword           string
	NEOPrivateKey             string
	ETHPrivateKey             string
	ETHContractAddress        string
	AnalyzerUrl               string
	AnalyzerUsername          string
	AnalyzerPassword          string
}
