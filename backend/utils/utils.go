package utils

import (
	//db2 "backend/db"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/template"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spaolacci/murmur3"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	debug bool
)

type Currencies struct {
	Name      string `json:"name"`
	Short     string `json:"short"`
	Smallest  string `json:"smallest"`
	FactorPow int64  `json:"factorPow"`
	IsCrypto  bool   `json:"isCrypto"`
}

var SupportedCurrencies = map[string]Currencies{
	"ETH": {Name: "Ethereum", Short: "ETH", Smallest: "wei", FactorPow: 18, IsCrypto: true},
	"GAS": {Name: "Neo Gas", Short: "GAS", Smallest: "mGAS", FactorPow: 8, IsCrypto: true},
	"USD": {Name: "US Dollar", Short: "USD", Smallest: "mUSD", FactorPow: 6, IsCrypto: false},
}

func StringPointer(s string) *string {
	return &s
}

func WriteErrorf(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	if debug {
		w.Write([]byte(`{"error":"` + msg + `"}`))
	}
}

func WriteJson(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err = json.NewEncoder(w).Encode(obj)
	if err != nil {
		WriteErrorf(w, http.StatusBadRequest, "Could encode json: %v", err)
	}
}

func WriteJsonStr(w http.ResponseWriter, obj string) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(obj))
	if err != nil {
		WriteErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
	}
}

func IntPow(n int64, m int64) int64 {
	if m == 0 {
		return 1
	}
	result := n
	for i := int64(2); i <= m; i++ {
		result *= n
	}
	return result
}

func IsUUIDZero(id uuid.UUID) bool {
	for x := 0; x < 16; x++ {
		if id[x] != 0 {
			return false
		}
	}
	return true
}

func GenRnd(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func IsValidUrl(s string) *string {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return nil
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}

	if u.Path == "" {
		return StringPointer(u.Host)
	}
	ts := strings.TrimPrefix(u.Path, "/")
	return StringPointer(ts)
}

func PrintMap(balanceMap map[string]*big.Int) string {
	s := ""
	for k, c := range SupportedCurrencies {
		v := balanceMap[k]
		if v != nil {
			vf := new(big.Float).SetInt(v)
			fi := new(big.Int).Exp(big.NewInt(10), big.NewInt(c.FactorPow), nil)
			ff := new(big.Float).SetInt(fi)
			rs := new(big.Float).Quo(vf, ff)
			s += fmt.Sprintf("%v", rs) + " " + k
		}
	}
	return s
}

func UsdBaseToCent(base int64) int64 {
	return base / 10_000
}

func UsdBaseToPrice(base int64) float64 {
	return float64(base) / 1_000_000
}

func UsdCentToBase(base int64) int64 {
	return base * 10_000
}

func GetFactorInt(currency string) (int64, error) {
	for supportedCurrency, cryptoCurrency := range SupportedCurrencies {
		if supportedCurrency == strings.ToUpper(currency) {
			return IntPow(10, cryptoCurrency.FactorPow), nil
		}
	}
	return 0, fmt.Errorf("currency not found, %v", currency)
}

func GetFactor(currency string) (*big.Int, error) {
	for supportedCurrency, cryptoCurrency := range SupportedCurrencies {
		if supportedCurrency == strings.ToUpper(currency) {
			return new(big.Int).Exp(big.NewInt(10), big.NewInt(cryptoCurrency.FactorPow), nil), nil
		}
	}
	return nil, fmt.Errorf("currency not found, %v", currency)
}

func ParseTemplate(filename string, other map[string]string) string {
	textMessage := ""
	tmplPlain, err := template.ParseFiles(filename)
	if err == nil {
		var buf bytes.Buffer
		err = tmplPlain.Execute(&buf, other)
		if err == nil {
			textMessage = buf.String()
		} else {
			log.Printf("cannot execute template file [%v], err: %v", filename, err)
		}
	} else {
		log.Printf("cannot prepare file template file [%v], err: %v", filename, err)
	}
	return textMessage
}

func myHash(s string) float64 {
	i := murmur3.Sum32([]byte(s))
	const maxUint32 = ^uint32(0)
	return float64(i) / float64(maxUint32)
}

func GetColor1(input string) string {
	a := strconv.Itoa(int(12 * (30 * myHash(input+"a"))))
	b := strconv.Itoa(int(35 + 10*(5*myHash(input+"b"))))
	c := strconv.Itoa(int(25 + 10*(5*myHash(input+"c"))))
	return "hsl(" + a + "," + b + "%," + c + "%)"
}
