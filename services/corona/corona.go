package corona

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	RUB    string = "810"
	USD    string = "840"
	GEL    string = "981"
	ApiUrl string = "https://koronapay.com/transfers/online/api/transfers/tariffs"
)

type ResponseBody struct {
	SendingCurrency struct {
		ID   string `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"sendingCurrency"`
	SendingAmount                  int `json:"sendingAmount"`
	SendingAmountDiscount          int `json:"sendingAmountDiscount"`
	SendingAmountWithoutCommission int `json:"sendingAmountWithoutCommission"`
	SendingCommission              int `json:"sendingCommission"`
	SendingCommissionDiscount      int `json:"sendingCommissionDiscount"`
	SendingTransferCommission      int `json:"sendingTransferCommission"`
	PaidNotificationCommission     int `json:"paidNotificationCommission"`
	ReceivingCurrency              struct {
		ID   string `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"receivingCurrency"`
	ReceivingAmount      int                    `json:"receivingAmount"`
	ExchangeRate         float64                `json:"exchangeRate"`
	ExchangeRateType     string                 `json:"exchangeRateType"`
	ExchangeRateDiscount int                    `json:"exchangeRateDiscount"`
	Profit               int                    `json:"profit"`
	Properties           map[string]interface{} `json:"properties"`
}

func GetRates() map[string]string {
	result := map[string]string{}

	if rubUsd, err := getRate(RUB, USD); err == nil {
		result["USD"] = rubUsd
	}
	if rubGel, err := getRate(RUB, GEL); err == nil {
		result["GEL"] = rubGel
	}

	return result
}

func getRate(inCurrencyCode string, outCurrencyCode string) (string, error) {
	params := url.Values{}
	params.Add("sendingCurrencyId", inCurrencyCode)
	params.Add("receivingCurrencyId", outCurrencyCode)

	params.Add("receivingCountryId", "GEO")
	params.Add("paymentMethod", "debitCard")
	params.Add("receivingAmount", "10000")
	params.Add("receivingMethod", "cash")
	params.Add("sendingCountryId", "RUS")

	req, err := http.NewRequest("GET", ApiUrl, nil)

	if err != nil {
		log.Printf("failed to create request %v", err)
		return "", fmt.Errorf("failed to create request: %s", err)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Add("accept", "application/vnd.cft-data.v2.99+json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("authority", "koronapay.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.Header.Add("x-application", "Qpay-Web/3.0")
	req.Header.Add("referer", "https://koronapay.com/transfers/online/")
	req.Header.Add("cookie", "qpay-web/3.0_locale=en; qpay-web/3.0_csrf-token-v2=4cd2201d99a761a50fa4a82d2227242d; _gid=GA1.2.2095589933.1680792866; _gcl_au=1.1.1136328774.1680792866; _dc_gtm_UA-100141486-1=1; _dc_gtm_UA-100141486-2=1; _dc_gtm_UA-100141486-25=1; _dc_gtm_UA-100141486-26=1; _ga_H68H5PL1N6=GS1.1.1680792865.1.0.1680792865.60.0.0; _ym_uid=1680792866715419461; _ym_d=1680792866; _ym_isad=2; _ym_visorc=b; tmr_lvid=dd98f4bf3336d2a96ae945cac4f22a7a; tmr_lvidTS=1680792866645; tmr_detect=0%7C1680792868964; _ga=GA1.2.785185665.1680792866; ROUTEID=2b6828adbefbb2eb|ZC7dM; _gali=changeable-field-input-amount")
	req.Header.Add("accept-language", "en")
	req.Header.Add("ssr-fetch-site", "same-origin")

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
		log.Printf("failed to fetch exchange rate %v", resp.Status)
		return "", fmt.Errorf("failed to fetch exchange rate: %s", resp.Status)
	}

	var jsonResp []ResponseBody

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	return fmt.Sprintf("%.5f", jsonResp[0].ExchangeRate), nil
}
