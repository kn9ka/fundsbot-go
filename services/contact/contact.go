package contact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	USD     string = "USD"
	GEL     string = "GEL"
	ApiUrl  string = "https://online.contact-sys.com/api/contact/v2"
	SiteUrl string = "https://online.contact-sys.com"
	Name    string = "Contact "
)

type FeesResponseBody struct {
	Rate string `json:"rate"`
}

type AuthTokenResponseBody struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type CreateFormResponseBody struct {
	Id string `json:"id"`
}

type FetchClient struct {
	Client  *http.Client
	Token   string
	Cookies struct {
		Refresh string
		Access  string
	}
}

var instance *FetchClient

func GetRates() map[string]string {
	result := map[string]string{}

	if rubUsd, err := getRate(USD); err == nil {
		result[USD] = rubUsd
	}
	if rubGel, err := getRate(GEL); err == nil {
		result[GEL] = rubGel
	}

	return result
}

func getFetchClient() *FetchClient {
	if instance == nil {
		instance = &FetchClient{
			Client: &http.Client{},
		}
	}
	if instance.Token == "" {
		instance.RefreshAccessToken()
	}
	return instance
}

func (fc *FetchClient) RefreshAccessToken() {
	data := struct {
		TokenType string `json:"tokenType"`
		GrantType string `json:"grantType"`
		Ticket    string `json:"ticket"`
	}{
		TokenType: "SplitTokenV2",
		GrantType: "anonymous",
		Ticket:    "D5267BED-18CC-4661-B03A-65934CAE1CA4",
	}
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("failed to marshal JSON: %s\n", err)
	}

	req, err := http.NewRequest("POST", ApiUrl+"/auth/token", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("failed to send request: %s", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Unable to close request body: %v", err)
		}
	}(resp.Body)

	cookies := resp.Header.Values("Set-Cookie")
	for _, cookieString := range cookies {
		cookieParts := strings.SplitN(cookieString, "=", 2)
		if len(cookieParts) != 2 {
			fmt.Printf("wrong format for cookie: %s", cookieString)
		}

		cookieName := strings.TrimSpace(cookieParts[0])
		cookieValue := strings.TrimSpace(cookieParts[1])
		cookieParams := strings.SplitN(cookieValue, ";", 2)

		if cookieName == "tokenTailRefresh2" {
			fc.Cookies.Refresh = strings.TrimSpace(cookieParams[0])
		}
		if cookieName == "tokenTailAccess2" {
			fc.Cookies.Access = strings.TrimSpace(cookieParams[0])
		}
	}

	var jsonResp AuthTokenResponseBody

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
	}

	token := jsonResp.AccessToken

	if err != nil {
		log.Printf("failed to decode cookies: %v", err)
	}

	fc.Token = token
}

func (fc *FetchClient) SetHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	if fc.Token != "" {
		req.Header.Set("Authorization", "SplitTokenV2 "+fc.Token)
	}
	if fc.Cookies.Refresh != "" {
		req.AddCookie(&http.Cookie{Name: "tokenTailRefresh2", Value: fc.Cookies.Refresh})
	}
	if fc.Cookies.Access != "" {
		req.AddCookie(&http.Cookie{Name: "tokenTailAccess2", Value: fc.Cookies.Access})
	}
}

func (fc *FetchClient) DoRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	fc.SetHeaders(req)

	resp, err := fc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	cookies := resp.Header.Values("Set-Cookie")
	for _, cookieString := range cookies {
		cookieParts := strings.SplitN(cookieString, "=", 2)
		if len(cookieParts) != 2 {
			return nil, fmt.Errorf("wrong format for cookie: %s", cookieString)
		}

		cookieName := strings.TrimSpace(cookieParts[0])
		cookieValue := strings.TrimSpace(cookieParts[1])
		cookieParams := strings.SplitN(cookieValue, ";", 2)

		if cookieName == "tokenTailRefresh2" {
			fc.Cookies.Refresh = strings.TrimSpace(cookieParams[0])
		}
		if cookieName == "tokenTailAccess2" {
			fc.Cookies.Access = strings.TrimSpace(cookieParams[0])
		}
	}

	return resp, nil
}

func createExchangeForm() (string, error) {
	url := fmt.Sprintf("%s/trns/bank", ApiUrl)
	client := getFetchClient()

	data := struct {
		BankCode string `json:"bankCode"`
	}{
		BankCode: "CFRN",
	}

	// Преобразуем данные в формат JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %s", err)
	}

	resp, err := client.DoRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("failed to send request: %v", err)
		return "", fmt.Errorf("failed to send request: %s", err)
	}

	var jsonResp CreateFormResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	return jsonResp.Id, nil
}

func updateForm(formId string, outCurrency string) (bool, error) {
	url := fmt.Sprintf("%s/trns/%s/fields", ApiUrl, formId)
	client := getFetchClient()

	data := struct {
		Amount   string `json:"trnAmount"`
		Currency string `json:"trnCurrency"`
	}{
		Amount:   "1000",
		Currency: outCurrency,
	}

	// Преобразуем данные в формат JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal JSON: %s", err)
	}

	resp, err := client.DoRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("failed to send request: %v", err)
		return false, fmt.Errorf("failed to send request: %s", err)
	}

	return resp.StatusCode == 200, nil
}

func getRate(outCurrency string) (string, error) {
	client := getFetchClient()
	client.RefreshAccessToken()

	formId, err := createExchangeForm()
	_, err = updateForm(formId, outCurrency)
	if err != nil {
		return "", fmt.Errorf("error while updating form %v", err)
	}

	url := fmt.Sprintf("%s/trns/%s/fees", ApiUrl, formId)

	resp, err := client.DoRequest("POST", url, nil)
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

	var jsonResp FeesResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	return jsonResp.Rate, nil

}
