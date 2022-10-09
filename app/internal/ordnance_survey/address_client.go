package ordnance_survey

import (
	"encoding/json"
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
	TotalResults int
	Results      []AddressDetails
}

// Implemented to flatten the struct returned (see test for nested results structure)
func (plr *PostcodeLookupResponse) UnmarshalJSON(b []byte) error {
	var originalPlr postcodeLookupResponse
	if err := json.Unmarshal(b, &originalPlr); err != nil {
		return err
	}

	plr.TotalResults = originalPlr.Header.TotalResults

	if plr.TotalResults > 0 {
		var addressDetails []AddressDetails

		for _, result := range originalPlr.Results {
			addressDetails = append(addressDetails, result.AddressDetails)
		}

		plr.Results = addressDetails
	}

	return nil
}

type PostcodeLookupResponseHeader struct {
	TotalResults int `json:"totalresults"`
}

type postcodeLookupResponse struct {
	Header  PostcodeLookupResponseHeader `json:"header"`
	Results []resultSet                  `json:"results"`
}

type resultSet struct {
	AddressDetails AddressDetails `json:"DPA"`
}

func NewClient(baseUrl, apiKey string, httpClient *http.Client) AddressClient {
	return AddressClient{
		BaseUrl:    baseUrl,
		ApiKey:     apiKey,
		HttpClient: httpClient,
	}
}

func (ac *AddressClient) FindAddress(postcode string) (PostcodeLookupResponse, error) {
	query := url.Values{
		"postcode": {strings.ReplaceAll(postcode, " ", "")},
		"key":      {ac.ApiKey},
	}

	reqUrl := ac.BaseUrl + "/search/places/v1/postcode?" + query.Encode()

	req, _ := http.NewRequest("GET", reqUrl, nil)
	req.Header.Add("accept", "application/json")

	resp, err := ac.HttpClient.Do(req)

	if err != nil {
		return PostcodeLookupResponse{}, err
	}

	defer resp.Body.Close()

	var postcodeLookupResponse PostcodeLookupResponse

	if err := json.NewDecoder(resp.Body).Decode(&postcodeLookupResponse); err != nil {
		return PostcodeLookupResponse{}, err
	}

	return postcodeLookupResponse, nil
}
