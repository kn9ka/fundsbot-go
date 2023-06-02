package unistream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	RUB     string = "RUB"
	USD     string = "USD"
	GEL     string = "GEL"
	EUR     string = "EUR"
	ApiUrl  string = "https://api6.unistream.com/api/v1/transfer/calculate"
	SiteUrl string = "https://unistream.ru/"
	Name    string = "Юнистрим"
)

type ResponseBody struct {
	Message string `json:"message"`
	Fees    []struct {
		Name                     string  `json:"name"`
		AcceptedAmount           float64 `json:"acceptedAmount"`
		AcceptedCurrency         string  `json:"acceptedCurrency"`
		WithdrawAmount           float64 `json:"withdrawAmount"`
		WithdrawCurrency         string  `json:"withdrawCurrency"`
		Rate                     float64 `json:"rate"`
		AcceptedTotalFee         float64 `json:"acceptedTotalFee"`
		AcceptedTotalFeeCurrency string  `json:"acceptedTotalFeeCurrency"`
	} `json:"fees"`
}

func GetRates() map[string]string {
	result := map[string]string{}

	if rubUsd, err := getRate(RUB, USD); err == nil {
		result["USD"] = rubUsd
	}
	if rubGel, err := getRate(RUB, GEL); err == nil {
		result["GEL"] = rubGel
	}
	if rubEur, err := getRate(RUB, EUR); err == nil {
		result["EUR"] = rubEur
	}

	return result
}

func getRate(inCurrencyCode string, outCurrencyCode string) (string, error) {
	form := url.Values{}
	form.Add("senderBankId", "361934")
	form.Add("acceptedCurrency", inCurrencyCode)
	form.Add("withdrawCurrency", outCurrencyCode)
	form.Add("amount", "1000")
	form.Add("countryCode", "GEO")

	req, err := http.NewRequest("POST", ApiUrl, bytes.NewBufferString(form.Encode()))

	if err != nil {
		log.Printf("failed to create request: %v", err)
		return "", fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("failed to send request: %v", err)
		return "", fmt.Errorf("failed to send request: %s", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Unable to close request body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to fetch exchange rate: %v", resp.Status)
		return "", fmt.Errorf("failed to fetch exchange rate: %s", resp.Status)
	}

	var jsonResp ResponseBody

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	rate := fmt.Sprintf(
		"%.5f",
		jsonResp.Fees[0].AcceptedAmount/jsonResp.Fees[0].WithdrawAmount,
	)
	return rate, nil
}
