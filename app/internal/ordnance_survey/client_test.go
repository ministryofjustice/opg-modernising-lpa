package ordnance_survey

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

const multipleResultsJson = `
{
	"header": {
		"uri": "https://api.os.uk/search/places/v1/postcode?postcode=B147ET",
		"query": "postcode=B147ET",
		"offset": 0,
		"totalresults": 2,
		"format": "JSON",
		"dataset": "DPA",
		"lr": "EN,CY",
		"maxresults": 100,
		"epoch": "96",
		"output_srs": "EPSG:27700"
	},
	"results": [ {
		"DPA": {
			"UPRN": "100071390703",
			"UDPRN": "432175",
			"ADDRESS": "123, MELTON ROAD, BIRMINGHAM, B14 7ET",
			"BUILDING_NUMBER": "123",
			"THOROUGHFARE_NAME": "MELTON ROAD",
			"POST_TOWN": "BIRMINGHAM",
			"POSTCODE": "B14 7ET",
			"RPC": "1",
			"X_COORDINATE": 407783.0,
			"Y_COORDINATE": 281505.0,
			"STATUS": "APPROVED",
			"LOGICAL_STATUS_CODE": "1",
			"CLASSIFICATION_CODE": "RD04",
			"CLASSIFICATION_CODE_DESCRIPTION": "Terraced",
			"LOCAL_CUSTODIAN_CODE": 4605,
			"LOCAL_CUSTODIAN_CODE_DESCRIPTION": "BIRMINGHAM",
			"COUNTRY_CODE": "E",
			"COUNTRY_CODE_DESCRIPTION": "This record is within England",
			"POSTAL_ADDRESS_CODE": "D",
			"POSTAL_ADDRESS_CODE_DESCRIPTION": "A record which is linked to PAF",
			"BLPU_STATE_CODE": "2",
			"BLPU_STATE_CODE_DESCRIPTION": "In use",
			"TOPOGRAPHY_LAYER_TOID": "osgb1000020369531",
			"LAST_UPDATE_DATE": "10/02/2016",
			"ENTRY_DATE": "16/04/2001",
			"BLPU_STATE_DATE": "29/04/2013",
			"LANGUAGE": "EN",
			"MATCH": 1.0,
			"MATCH_DESCRIPTION": "EXACT",
			"DELIVERY_POINT_SUFFIX": "1Q"
		}
	}, {
		"DPA": {
			"UPRN": "100070449924",
			"UDPRN": "432202",
			"ADDRESS": "87A, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
			"BUILDING_NAME": "87A",
			"THOROUGHFARE_NAME": "MELTON ROAD",
			"DEPENDENT_LOCALITY": "KINGS HEATH",
			"POST_TOWN": "BIRMINGHAM",
			"POSTCODE": "B14 7ET",
			"RPC": "1",
			"X_COORDINATE": 407799.0,
			"Y_COORDINATE": 281591.0,
			"STATUS": "APPROVED",
			"LOGICAL_STATUS_CODE": "1",
			"CLASSIFICATION_CODE": "RD04",
			"CLASSIFICATION_CODE_DESCRIPTION": "Terraced",
			"LOCAL_CUSTODIAN_CODE": 4605,
			"LOCAL_CUSTODIAN_CODE_DESCRIPTION": "BIRMINGHAM",
			"COUNTRY_CODE": "E",
			"COUNTRY_CODE_DESCRIPTION": "This record is within England",
			"POSTAL_ADDRESS_CODE": "D",
			"POSTAL_ADDRESS_CODE_DESCRIPTION": "A record which is linked to PAF",
			"BLPU_STATE_CODE": "2",
			"BLPU_STATE_CODE_DESCRIPTION": "In use",
			"TOPOGRAPHY_LAYER_TOID": "osgb1000020369554",
			"LAST_UPDATE_DATE": "10/02/2016",
			"ENTRY_DATE": "16/04/2001",
			"BLPU_STATE_DATE": "29/04/2013",
			"LANGUAGE": "EN",
			"MATCH": 1.0,
			"MATCH_DESCRIPTION": "EXACT",
			"DELIVERY_POINT_SUFFIX": "3E"
		}
	}]
}
`

const noResultsJson = `
{
	"header": {
		"uri": "https://api.os.uk/search/places/v1/postcode?postcode=XXXXXX",
		"query": "postcode=XXXXXX",
		"offset": 0,
		"totalresults": 0,
		"format": "JSON",
		"dataset": "DPA",
		"lr": "EN,CY",
		"maxresults": 100,
		"epoch": "96",
		"output_srs": "EPSG:27700"
	}
}
`

type mockDoer struct {
	mock.Mock
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestFindAddress(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                   string
		postcode               string
		queryPostcode          string
		responseJson           string
		expectedResponseObject PostcodeLookupResponse
	}{
		{
			"Multiple results",
			"B14 7ET",
			"B147ET",
			multipleResultsJson,
			PostcodeLookupResponse{
				TotalResults: 2,
				Results: []AddressDetails{
					{
						Address:           "123, MELTON ROAD, BIRMINGHAM, B14 7ET",
						BuildingName:      "",
						BuildingNumber:    "123",
						ThoroughFareName:  "MELTON ROAD",
						DependentLocality: "",
						Town:              "BIRMINGHAM",
						Postcode:          "B14 7ET",
					},
					{
						Address:           "87A, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
						BuildingName:      "87A",
						BuildingNumber:    "",
						ThoroughFareName:  "MELTON ROAD",
						DependentLocality: "KINGS HEATH",
						Town:              "BIRMINGHAM",
						Postcode:          "B14 7ET",
					},
				},
			},
		},
		{
			"No results",
			"  X XX XX X ",
			"XXXXXX",
			noResultsJson,
			PostcodeLookupResponse{TotalResults: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, tc.queryPostcode, req.URL.Query().Get("postcode"), "Request was missing 'postcode' query with expected value")
				assert.Equal(t, "fake-api-key", req.URL.Query().Get("key"), "Request was missing 'key' query with expected value")

				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(tc.responseJson))
			}))

			defer server.Close()

			client := NewClient(server.URL, "fake-api-key", server.Client())
			results, err := client.LookupPostcode(ctx, tc.postcode)

			assert.Equal(t, tc.expectedResponseObject, results)
			assert.Nil(t, err)
		})
	}

	t.Run("returns an error on request initialisation errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		defer server.Close()

		client := NewClient("not an url", "fake-api-key", server.Client())
		_, err := client.LookupPostcode(ctx, "ABC")

		assert.ErrorContains(t, err, "unsupported protocol scheme")
	})

	t.Run("returns an error on making a request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		defer server.Close()

		mockDoer := mockDoer{}
		mockDoer.
			On("Do", mock.Anything).
			Return(&http.Response{}, errors.New("an error occurred"))

		client := NewClient(server.URL, "fake-api-key", &mockDoer)
		_, err := client.LookupPostcode(ctx, "ABC")

		assert.ErrorContains(t, err, "an error occurred")
	})

	t.Run("returns an error on json marshalling errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("not JSON"))
		}))

		defer server.Close()

		client := NewClient(server.URL, "fake-api-key", server.Client())
		_, err := client.LookupPostcode(ctx, "ABC")

		assert.ErrorContains(t, err, "invalid character")
	})
}

func TestPostcodeLookupResponse(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("simplifies JSON structure in PostcodeLookupResponse object", func(t *testing.T) {
			want := PostcodeLookupResponse{
				TotalResults: 2,
				Results: []AddressDetails{
					{
						Address:           "123, MELTON ROAD, BIRMINGHAM, B14 7ET",
						BuildingName:      "",
						BuildingNumber:    "123",
						ThoroughFareName:  "MELTON ROAD",
						DependentLocality: "",
						Town:              "BIRMINGHAM",
						Postcode:          "B14 7ET",
					},
					{
						Address:           "87A, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
						BuildingName:      "87A",
						BuildingNumber:    "",
						ThoroughFareName:  "MELTON ROAD",
						DependentLocality: "KINGS HEATH",
						Town:              "BIRMINGHAM",
						Postcode:          "B14 7ET",
					},
				},
			}

			plr := PostcodeLookupResponse{}
			plr.UnmarshalJSON([]byte(multipleResultsJson))

			assert.Equal(t, want, plr)
		})

		t.Run("returns an error on unsuccessful an unmarshal", func(t *testing.T) {
			plr := PostcodeLookupResponse{}
			err := plr.UnmarshalJSON([]byte("not json"))

			assert.ErrorContains(t, err, "invalid character")
		})
	})

}
