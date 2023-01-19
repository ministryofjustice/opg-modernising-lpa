package place

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const postcodeEndpoint = "/search/places/v1/postcode?"

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl string
	apiKey  string
	doer    Doer
}

type addressDetails struct {
	Address           string `json:"ADDRESS"`
	SubBuildingName   string `json:"SUB_BUILDING_NAME"`
	BuildingName      string `json:"BUILDING_NAME"`
	BuildingNumber    string `json:"BUILDING_NUMBER"`
	ThoroughFareName  string `json:"THOROUGHFARE_NAME"`
	DependentLocality string `json:"DEPENDENT_LOCALITY"`
	Town              string `json:"POST_TOWN"`
	Postcode          string `json:"POSTCODE"`
}

type postcodeLookupResponse struct {
	Results []ResultSet `json:"results"`
}

type ResultSet struct {
	AddressDetails addressDetails `json:"DPA"`
}

func NewClient(baseUrl, apiKey string, httpClient Doer) *Client {
	return &Client{
		baseUrl: baseUrl,
		apiKey:  apiKey,
		doer:    httpClient,
	}
}

func (c *Client) LookupPostcode(ctx context.Context, postcode Postcode) ([]Address, error) {
	query := url.Values{
		"postcode": {strings.ReplaceAll(string(postcode), " ", "")},
		"key":      {c.apiKey},
	}

	reqUrl := c.baseUrl + postcodeEndpoint + query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)

	if err != nil {
		return []Address{}, err
	}

	req.Header.Add("accept", "application/json")

	resp, err := c.doer.Do(req)

	if err != nil {
		return []Address{}, err
	}

	defer resp.Body.Close()

	var postcodeLookupResponse postcodeLookupResponse

	if err := json.NewDecoder(resp.Body).Decode(&postcodeLookupResponse); err != nil {
		return []Address{}, err
	}

	var addresses []Address

	for _, resultSet := range postcodeLookupResponse.Results {
		addresses = append(addresses, resultSet.AddressDetails.transformToAddress())
	}

	return addresses, nil
}

type Postcode string

func (p Postcode) IsUkFormat() bool {
	isUkPostcode, _ := regexp.MatchString("(?i)^([A-Z]{1,2}\\d[A-Z\\d]? ?\\d[A-Z]{2})$", strings.ReplaceAll(p.String(), " ", ""))
	return isUkPostcode
}

func (p Postcode) String() string {
	return string(p)
}

type Address struct {
	Line1      string
	Line2      string
	Line3      string
	TownOrCity string
	Postcode   Postcode
}

func (a Address) Encode() string {
	x, _ := json.Marshal(a)
	return string(x)
}

func (a Address) String() string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.Line3 != "" {
		parts = append(parts, a.Line3)
	}
	if a.TownOrCity != "" {
		parts = append(parts, a.TownOrCity)
	}
	if a.Postcode != "" {
		parts = append(parts, a.Postcode.String())
	}

	return strings.Join(parts, ", ")
}

func (ad *addressDetails) transformToAddress() Address {
	a := Address{}

	if len(ad.BuildingName) > 0 {
		if len(ad.SubBuildingName) > 0 {
			a.Line1 = fmt.Sprintf("%s, %s", ad.SubBuildingName, ad.BuildingName)
		} else {
			a.Line1 = ad.BuildingName
		}

		if len(ad.BuildingNumber) > 0 {
			a.Line2 = fmt.Sprintf("%s %s", ad.BuildingNumber, ad.ThoroughFareName)
		} else {
			a.Line2 = ad.ThoroughFareName
		}

		a.Line3 = ad.DependentLocality
	} else {
		a.Line1 = fmt.Sprintf("%s %s", ad.BuildingNumber, ad.ThoroughFareName)
		a.Line2 = ad.DependentLocality
	}

	a.TownOrCity = ad.Town
	a.Postcode = Postcode(ad.Postcode)

	return a
}
