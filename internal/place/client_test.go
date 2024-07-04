package place

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLookupPostcode(t *testing.T) {
	multipleAddressJson, _ := os.ReadFile("testdata/postcode-multiple-addresses.json")
	noAddressJson, _ := os.ReadFile("testdata/postcode-no-addresses.json")

	ctx := context.Background()

	testCases := map[string]struct {
		postcode      string
		queryPostcode string
		responseJson  string
		want          []Address
	}{
		"multiple": {
			postcode:      "B14 7ET",
			queryPostcode: "B147ET",
			responseJson:  string(multipleAddressJson),
			want: []Address{
				{
					Line1:      "123 MELTON ROAD",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ET",
					Country:    "GB",
				},
				{
					Line1:      "87A",
					Line2:      "MELTON ROAD",
					Line3:      "KINGS HEATH",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ET",
					Country:    "GB",
				},
			},
		},
		"no results": {
			postcode:      "  X XX XX X ",
			queryPostcode: "XXXXXX",
			responseJson:  string(noAddressJson),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, tc.queryPostcode, req.URL.Query().Get("postcode"), "Request was missing 'postcode' query with expected value")
				assert.Equal(t, "fake-api-key", req.URL.Query().Get("key"), "Request was missing 'key' query with expected value")

				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(tc.responseJson))
			}))

			defer server.Close()

			client := NewClient(server.URL, "fake-api-key", server.Client())
			results, err := client.LookupPostcode(ctx, tc.postcode)

			assert.Equal(t, tc.want, results)
			assert.Nil(t, err)
		})
	}

	t.Run("returns BadRequestError on invalid postcode", func(t *testing.T) {
		invalidPostcodeJson, _ := os.ReadFile("testdata/invalid-postcode-error.json")

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(invalidPostcodeJson)
		}))

		defer server.Close()

		client := NewClient(server.URL, "fake-api-key", server.Client())
		results, err := client.LookupPostcode(ctx, "ABC123")

		assert.Equal(t, []Address{}, results)
		assert.Equal(t, BadRequestError{
			Statuscode: 400,
			Message:    "Requested postcode must contain a minimum of the sector plus 1 digit of the district e.g. SO1. Requested postcode was ABC123",
		}, err)
	})

	t.Run("returns an error on request initialisation errors", func(t *testing.T) {
		client := NewClient("not an url", "fake-api-key", http.DefaultClient)
		_, err := client.LookupPostcode(ctx, "ABC")

		assert.ErrorContains(t, err, "unsupported protocol scheme")
	})

	t.Run("returns an error on making a request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))

		defer server.Close()

		mockDoer := newMockDoer(t)
		mockDoer.EXPECT().
			Do(mock.Anything).
			Return(&http.Response{}, errors.New("an error occurred"))

		client := NewClient(server.URL, "fake-api-key", mockDoer)
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

func TestAddress(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		testCases := []struct {
			name    string
			address Address
			want    string
		}{
			{
				"All props set",
				Address{
					Line1:      "Line 1",
					Line2:      "Line 2",
					Line3:      "Line 3",
					TownOrCity: "Town",
					Postcode:   "Postcode",
				},
				"Line 1, Line 2, Line 3, Town, Postcode",
			},
			{
				"Some props set",
				Address{
					Line1:      "Line 1",
					Line2:      "",
					Line3:      "Line 3",
					TownOrCity: "Town",
					Postcode:   "",
				},
				"Line 1, Line 3, Town",
			},
			{
				"No props set",
				Address{},
				"",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.want, tc.address.String())
			})
		}
	})
}

func TestTransformAddressDetailsToAddress(t *testing.T) {
	testCases := map[string]struct {
		ad   AddressDetails
		want Address
	}{
		"building number no building name": {
			ad: AddressDetails{
				Address:           "1, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "1",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "1 MELTON ROAD", Line2: "", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name no building number": {
			ad: AddressDetails{
				Address:           "1A, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "1A",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "1A", Line2: "MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name and building number": {
			ad: AddressDetails{
				Address:           "MELTON HOUSE, 2 MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "2",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "2 MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building number": {
			ad: AddressDetails{
				Address:           "3, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "3",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "3 MELTON ROAD", Line2: "KINGS HEATH", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building name": {
			ad: AddressDetails{
				Address:           "MELTON HOUSE, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building name and building number": {
			ad: AddressDetails{
				Address:           "MELTON HOUSE, 5 MELTON ROAD, KINGS HEATH BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "5",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "5 MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name and sub building name": {
			ad: AddressDetails{
				Address:          "APARTMENT 34, CHARLES HOUSE, PARK ROW, NOTTINGHAM, NG1 6GR",
				SubBuildingName:  "APARTMENT 34",
				BuildingName:     "CHARLES HOUSE",
				ThoroughFareName: "PARK ROW",
				Town:             "NOTTINGHAM",
				Postcode:         "NG1 6GR",
			},
			want: Address{Line1: "APARTMENT 34, CHARLES HOUSE", Line2: "PARK ROW", TownOrCity: "NOTTINGHAM", Postcode: "NG1 6GR", Country: "GB"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ad.TransformToAddress())
		})
	}
}
