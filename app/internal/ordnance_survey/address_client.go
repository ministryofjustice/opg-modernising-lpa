package ordnance_survey

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const postcodeEndpoint = "/search/places/v1/postcode?"

type AddressClient struct {
	BaseUrl    string
	ApiKey     string
	HttpClient *http.Client
}

type AddressDetails struct {
	Address           string `json:"ADDRESS"`
	BuildingName      string `json:"BUILDING_NAME,omitempty"`
	BuildingNumber    string `json:"BUILDING_NUMBER,omitempty"`
	ThoroughFareName  string `json:"THOROUGHFARE_NAME"`
	DependentLocality string `json:"DEPENDENT_LOCALITY,omitempty"`
	Town              string `json:"POST_TOWN"`
	Postcode          string `json:"POSTCODE"`
}

type PostcodeLookupResponse struct {
	TotalResults int
	Results      []AddressDetails
}

type Address struct {
	Line1      string
	Line2      string
	TownOrCity string
	Postcode   string
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

func (plr *PostcodeLookupResponse) GetAddresses() []Address {
	var addresses []Address

	for _, addressDetail := range plr.Results {
		a := Address{}

		if len(addressDetail.BuildingNumber) > 0 {
			a.Line1 = fmt.Sprintf("%s %s", addressDetail.BuildingNumber, addressDetail.ThoroughFareName)
		} else {
			a.Line1 = fmt.Sprintf("%s %s", addressDetail.BuildingName, addressDetail.ThoroughFareName)
		}

		a.Line2 = addressDetail.DependentLocality
		a.TownOrCity = addressDetail.Town
		a.Postcode = addressDetail.Postcode

		addresses = append(addresses, a)
	}

	return addresses
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

func (ac *AddressClient) LookupPostcode(postcode string) (PostcodeLookupResponse, error) {
	query := url.Values{
		"postcode": {strings.ReplaceAll(postcode, " ", "")},
		"key":      {ac.ApiKey},
	}

	reqUrl := ac.BaseUrl + postcodeEndpoint + query.Encode()

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
