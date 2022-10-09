package ordnance_survey

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AddressClient struct {
	BaseUrl    string
	ApiKey     string
	HttpClient *http.Client
}

type AddressDetails struct {
	Address          string `json:"ADDRESS"`
	BuildingName     string `json:"BUILDING_NAME,omitempty"`
	BuildingNumber   string `json:"BUILDING_NUMBER,omitempty"`
	ThoroughFareName string `json:"THOROUGHFARE_NAME"`
	Town             string `json:"POST_TOWN"`
	Postcode         string `json:"POSTCODE"`
}

type PostcodeLookupResponse struct {
	Results []AddressDetails `json:"results"`
}

// Implemented to flatten the struct returned (see test for nested results structure)
func (plr *PostcodeLookupResponse) UnmarshalJSON(b []byte) error {
	var originalPlr postcodeLookupResponse
	if err := json.Unmarshal(b, &originalPlr); err != nil {
		return err
	}

	var addressDetails []AddressDetails

	for _, result := range originalPlr.Results {
		addressDetails = append(addressDetails, result.AddressDetails)
	}

	plr.Results = addressDetails

	return nil
}

type postcodeLookupResponse struct {
	Results []ResultSet `json:"results"`
}

type ResultSet struct {
	AddressDetails AddressDetails `json:"DPA"`
}

func NewClient(baseUrl, apiKey string, httpClient *http.Client) AddressClient {
	return AddressClient{
		BaseUrl:    baseUrl,
		ApiKey:     apiKey,
		HttpClient: httpClient,
	}
}

func (ac *AddressClient) FindAddress(postcode string) PostcodeLookupResponse {
	query := url.Values{
		"postcode": {strings.ReplaceAll(postcode, " ", "")},
		"key":      {ac.ApiKey},
	}

	reqUrl := ac.BaseUrl + "/search/places/v1/postcode?" + query.Encode()

	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("accept", "application/json")

	resp, _ := ac.HttpClient.Do(req)

	defer resp.Body.Close()

	var postcodeLookupResponse PostcodeLookupResponse

	if err := json.NewDecoder(resp.Body).Decode(&postcodeLookupResponse); err != nil {
		fmt.Println(err.Error())
		return PostcodeLookupResponse{}
	}

	return postcodeLookupResponse
}
