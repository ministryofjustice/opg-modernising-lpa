package place

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

type mockDoer struct {
	mock.Mock
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestLookupPostcode(t *testing.T) {
	multipleAddressJson, _ := os.ReadFile("testdata/postcode-multiple-addresses.json")
	noAddressJson, _ := os.ReadFile("testdata/postcode-no-addresses.json")

	ctx := context.Background()

	testCases := map[string]struct {
		postcode      Postcode
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
				},
				{
					Line1:      "87A",
					Line2:      "MELTON ROAD",
					Line3:      "KINGS HEATH",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ET",
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

	t.Run("returns an error on request initialisation errors", func(t *testing.T) {
		client := NewClient("not an url", "fake-api-key", http.DefaultClient)
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
		ad   addressDetails
		want Address
	}{
		"building number no building name": {
			ad: addressDetails{
				Address:           "1, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "1",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "1 MELTON ROAD", Line2: "", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"building name no building number": {
			ad: addressDetails{
				Address:           "1A, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "1A",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "1A", Line2: "MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"building name and building number": {
			ad: addressDetails{
				Address:           "MELTON HOUSE, 2 MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "2",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "2 MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"dependent locality building number": {
			ad: addressDetails{
				Address:           "3, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "3",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "3 MELTON ROAD", Line2: "KINGS HEATH", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"dependent locality building name": {
			ad: addressDetails{
				Address:           "MELTON HOUSE, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"dependent locality building name and building number": {
			ad: addressDetails{
				Address:           "MELTON HOUSE, 5 MELTON ROAD, KINGS HEATH BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "5",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			want: Address{Line1: "MELTON HOUSE", Line2: "5 MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		"building name and sub building name": {
			ad: addressDetails{
				Address:          "APARTMENT 34, CHARLES HOUSE, PARK ROW, NOTTINGHAM, NG1 6GR",
				SubBuildingName:  "APARTMENT 34",
				BuildingName:     "CHARLES HOUSE",
				ThoroughFareName: "PARK ROW",
				Town:             "NOTTINGHAM",
				Postcode:         "NG1 6GR",
			},
			want: Address{Line1: "APARTMENT 34, CHARLES HOUSE", Line2: "PARK ROW", TownOrCity: "NOTTINGHAM", Postcode: "NG1 6GR"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ad.transformToAddress())
		})
	}
}

func TestPostcodeIsUkFormat(t *testing.T) {
	testCases := map[string]struct {
		postcode           Postcode
		expectedIsUkFormat bool
	}{
		"valid space": {
			postcode:           "AA1A 1AA",
			expectedIsUkFormat: true,
		},
		"valid no space": {
			postcode:           "AA1A1AA",
			expectedIsUkFormat: true,
		},
		"valid excess whitespace": {
			postcode:           " A  A   1 A  1 A A    ",
			expectedIsUkFormat: true,
		},
		"valid mixed case": {
			postcode:           "Aa1A 1Aa",
			expectedIsUkFormat: true,
		},
		"valid shorter format": {
			postcode:           "AA1 1AA",
			expectedIsUkFormat: true,
		},
		"valid shortest format": {
			postcode:           "A1 1AA",
			expectedIsUkFormat: true,
		},
		"invalid too many first chars": {
			postcode:           "A1AAA 1AA",
			expectedIsUkFormat: false,
		},
		"invalid too many second chars": {
			postcode:           "A1A 1AAAA",
			expectedIsUkFormat: false,
		},
		"invalid all alpha": {
			postcode:           "AAA AAA",
			expectedIsUkFormat: false,
		},
		"invalid all numeric": {
			postcode:           "111 111",
			expectedIsUkFormat: false,
		},
		"invalid non alpha-numeric": {
			postcode:           "*&^ £@!",
			expectedIsUkFormat: false,
		},
		"invalid empty": {
			postcode:           "",
			expectedIsUkFormat: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsUkFormat, tc.postcode.IsUkFormat())
		})
	}

}
