package alphaVantage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	ApiUrl   = "https://www.alphavantage.co/query"
	Function = "CURRENCY_EXCHANGE_RATE"
)

type ResponseBody struct {
	RealtimeCurrencyExchangeRate struct {
		FromCurrencyCode string `json:"1. From_Currency Code"`
		FromCurrencyName string `json:"2. From_Currency Name"`
		ToCurrencyCode   string `json:"3. To_Currency Code"`
		ToCurrencyName   string `json:"4. To_Currency Name"`
		ExchangeRate     string `json:"5. Exchange Rate"`
		LastRefreshed    string `json:"6. Last Refreshed"`
		TimeZone         string `json:"7. Time Zone"`
		BidPrice         string `json:"8. Bid Price"`
		AskPrice         string `json:"9. Ask Price"`
	} `json:"Realtime Currency Exchange Rate"`
}

func GetRates() map[string]string {
	rates := map[string]string{}

	if rubUsd, err := getRate("USD", "RUB"); err == nil {
		rates["USD"] = rubUsd
	}
	if rubEur, err := getRate("EUR", "RUB"); err == nil {
		rates["EUR"] = rubEur
	}
	if eurGel, err := getRate("EUR", "GEL"); err == nil {
		rates["GEL"] = fmt.Sprintf("%.5f", parseFloat(rates["EUR"])/parseFloat(eurGel))
	}

	return rates
}

func getRate(inCurrencyCode string, outCurrencyCode string) (string, error) {
	params := url.Values{}
	params.Add("function", Function)
	params.Add("from_currency", inCurrencyCode)
	params.Add("to_currency", outCurrencyCode)
	params.Add("apikey", os.Getenv("ALPHA_VANTAGE_API_KEY"))

	req, err := http.NewRequest("GET", ApiUrl, nil)

	if err != nil {
		log.Printf("Unable to create request: %v", err)
		return "", err
	}

	req.URL.RawQuery = params.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Unable to send request: %v", err)
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Unable to close request body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to fetch exchange rate: %v", err)
		return "", fmt.Errorf("failed to fetch exchange rate: %s", resp.Status)
	}

	var jsonResp ResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	return fmt.Sprintf("%s", jsonResp.RealtimeCurrencyExchangeRate.ExchangeRate), nil

}

func parseFloat(str string) float64 {
	amount, err := strconv.ParseFloat(strings.Replace(str, ",", ".", -1), 64)
	if err != nil {
		log.Printf("failed to parse float: %v", err)
		return 0
	}
	return amount
}
