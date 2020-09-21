package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const ExchangeRecordRefresInterval = time.Hour

type ExchangeRateRecord struct {
	LastUpdate time.Time `json:"date"`
	Rate       float64   `json:"rate"`
}

func (er *ExchangeRateRecord) IsStale() bool {
	return er == nil || time.Since(er.LastUpdate) > ExchangeRecordRefresInterval
}

type ExchangeRateCache map[string]*ExchangeRateRecord

func (erc ExchangeRateCache) Convert(er *ParseResult) {
	pairName := fmt.Sprintf("%s_%s", er.SrcCurrency, er.TgtCurrency)
	rate, ok := erc[pairName]
	if !ok {
		// get from the net
		rate = &ExchangeRateRecord{
			LastUpdate: time.Now(),
			Rate:       GetConversion(pairName),
		}
		erc[pairName] = rate
		// Save
		if !args.NoCache {

			data, err := json.Marshal(&erc)
			if err != nil {
				return
			}
			_ = ioutil.WriteFile(CurConvCacheFile, data, 0755)

		}
	}
	er.Tgt = er.Src * rate.Rate
}

func EmptyExchangRatesCache() ExchangeRateCache {
	return make(ExchangeRateCache)
}

func LoadExchangRatesCache() ExchangeRateCache {
	ratesData, err := ioutil.ReadFile(CurConvCacheFile)
	if err != nil {
		ratesData = []byte("{}")
	}

	exchangeData := make(map[string]*ExchangeRateRecord)
	_ = json.Unmarshal(ratesData, &exchangeData)
	for key, val := range exchangeData {
		if val.IsStale() {
			delete(exchangeData, key)
		}
	}

	return exchangeData
}

func GetConversion(pairName string) float64 {
	base, _ := url.Parse("https://free.currconv.com/api/v7/convert")

	// Query params
	params := url.Values{}
	params.Add("apiKey", args.ApiKey)
	params.Add("compact", "ultra")
	params.Add("q", pairName)

	base.RawQuery = params.Encode()
	resp, err := http.DefaultClient.Get(base.String())
	if err != nil {
		panic(UIError{`failed to get the exchange rate from API`})
	}

	resp_data := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&resp_data)
	if err != nil {
		panic(UIError{"failed to get the exchange rate"})
	}

	if resp.StatusCode != 200 {
		errorObj, ok1 := resp_data["error"]
		errorText, ok2 := errorObj.(string)
		if !ok1 || !ok2 {
			panic(UIError{"Can't get the rate"})
		} else {
			panic(UIError{fmt.Sprintf(`failed to get the rate: "%s"`, errorText)})
		}
	}

	rateObj, ok := resp_data[pairName]
	rate, ok2 := rateObj.(float64)
	if !ok || !ok2 {
		panic(UIError{"failed to get the exchange rate"})
	}

	return rate
}
