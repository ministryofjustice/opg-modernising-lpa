package onelogin

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testNow   = time.Now()
	testNowFn = func() time.Time { return testNow }
)

func TestUserInfo(t *testing.T) {
	expectedUserInfo := UserInfo{
		Sub:             "urn:fdc:gov.uk:2022:56P4CMsGh_02YOlWpd8PAOI-2sVlB2nsNU7mcLZYhYw=",
		Email:           "email@example.com",
		EmailVerified:   true,
		Phone:           "01406946277",
		PhoneVerified:   true,
		CoreIdentityJWT: "a jwt",
		Addresses: []credentialAddress{
			{
				UPRN:                           json.Number("10022812929"),
				SubBuildingName:                "FLAT 5",
				BuildingName:                   "WEST LEA",
				BuildingNumber:                 "16",
				DependentStreetName:            "KINGS PARK",
				StreetName:                     "HIGH STREET",
				DoubleDependentAddressLocality: "EREWASH",
				DependentAddressLocality:       "LONG EATON",
				AddressLocality:                "GREAT MISSENDEN",
				PostalCode:                     "HP16 0AL",
				AddressCountry:                 "GB",
				ValidFrom:                      "2022-01-01",
			},
			{
				UPRN:                     json.Number("10002345923"),
				BuildingName:             "SAWLEY MARINA",
				StreetName:               "INGWORTH ROAD",
				DependentAddressLocality: "LONG EATON",
				AddressLocality:          "NOTTINGHAM",
				PostalCode:               "BH12 1JY",
				AddressCountry:           "GB",
				ValidUntil:               "2022-01-01",
			},
		},
	}

	body := `{  "sub": "urn:fdc:gov.uk:2022:56P4CMsGh_02YOlWpd8PAOI-2sVlB2nsNU7mcLZYhYw=",
	"email": "email@example.com",
	"email_verified": true,
	"phone": "01406946277",
	"phone_verified": true,
	"https://vocab.account.gov.uk/v1/coreIdentityJWT": "a jwt",
	"https://vocab.account.gov.uk/v1/address": [
		{
			"uprn": 10022812929,
			"subBuildingName": "FLAT 5",
			"buildingName": "WEST LEA",
			"buildingNumber": "16",
			"dependentStreetName": "KINGS PARK",
			"streetName": "HIGH STREET",
			"doubleDependentAddressLocality": "EREWASH",
			"dependentAddressLocality": "LONG EATON",
			"addressLocality": "GREAT MISSENDEN",
			"postalCode": "HP16 0AL",
			"addressCountry": "GB",
			"validFrom": "2022-01-01"
		},
		{
			"uprn": 10002345923,
			"buildingName": "SAWLEY MARINA",
			"streetName": "INGWORTH ROAD",
			"dependentAddressLocality": "LONG EATON",
			"addressLocality": "NOTTINGHAM",
			"postalCode": "BH12 1JY",
			"addressCountry": "GB",
			"validUntil": "2022-01-01"
		}
	]
}`

	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) &&
				assert.Equal(t, "http://user-info", r.URL.String()) &&
				assert.Equal(t, "Bearer hey", r.Header.Get("Authorization"))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil)

	c := &Client{
		httpClient: httpClient,
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				UserinfoEndpoint: "http://user-info",
			},
		},
	}

	userinfo, err := c.UserInfo(context.Background(), "hey")
	assert.Nil(t, err)
	assert.Equal(t, expectedUserInfo, userinfo)
}

func TestUserInfoWhenConfigurationError(t *testing.T) {
	c := &Client{
		openidConfiguration: &configurationClient{},
	}

	_, err := c.UserInfo(context.Background(), "hey")
	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestUserInfoWhenRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{}, expectedError)

	c := &Client{
		httpClient: httpClient,
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				UserinfoEndpoint: "http://user-info",
			},
		},
	}

	_, err := c.UserInfo(context.Background(), "hey")
	assert.Equal(t, expectedError, err)
}

func TestParseIdentityClaim(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	issuedAt := time.Now().Add(-time.Minute).Round(time.Second)

	c := &Client{
		now: testNowFn,
		didClient: &didClient{
			controllerID: "blah",
			assertionMethods: map[string]crypto.PublicKey{
				"blah#thing": &privateKey.PublicKey,
			},
		},
	}

	namePart := []map[string]any{
		{
			"validFrom": "2020-03-01",
			"nameParts": []map[string]string{
				{
					"value": "Alice",
					"type":  "GivenName",
				},
				{
					"value": "Jane",
					"type":  "GivenName",
				},
				{
					"value": "Laura",
					"type":  "GivenName",
				},
				{
					"value": "Doe",
					"type":  "FamilyName",
				},
			},
		},
		{
			"validUntil": "2020-03-01",
			"nameParts": []map[string]string{
				{
					"value": "Alice",
					"type":  "GivenName",
				},
				{
					"value": "Eod",
					"type":  "FamilyName",
				},
			},
		},
	}

	birthDatePart := []map[string]any{
		{
			"value": "1970-01-02",
		},
	}

	vc := map[string]any{
		"credentialSubject": map[string]any{
			"name":      namePart,
			"birthDate": birthDatePart,
		},
	}

	mustSign := func(token *jwt.Token, key any) string {
		token.Header[jwkset.HeaderKID] = "blah#thing"
		s, err := token.SignedString(key)

		assert.Nil(t, err)
		return s
	}

	missingKIDHeaderToken, _ := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"vc":  vc,
	}).SignedString(privateKey)

	kidNotAStringToken := func(token *jwt.Token, key any) string {
		token.Header[jwkset.HeaderKID] = 1
		s, err := token.SignedString(key)

		assert.Nil(t, err)
		return s
	}

	testcases := map[string]struct {
		token       string
		userData    identity.UserData
		error       error
		returnCodes []ReturnCodeInfo
	}{
		"with required claims": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), privateKey),
			userData: identity.UserData{
				Status:      identity.StatusConfirmed,
				FirstNames:  "Alice Jane Laura",
				LastName:    "Doe",
				DateOfBirth: date.New("1970", "01", "02"),
				CheckedAt:   issuedAt,
				CurrentAddress: place.Address{
					Line1:    "1 Fake Road",
					Postcode: "B14 7ED",
					Country:  "GB",
				},
			},
		},
		"with return code 'A'": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), privateKey),
			userData: identity.UserData{
				Status:      identity.StatusConfirmed,
				FirstNames:  "Alice Jane Laura",
				LastName:    "Doe",
				DateOfBirth: date.New("1970", "01", "02"),
				CheckedAt:   issuedAt,
				CurrentAddress: place.Address{
					Line1:    "1 Fake Road",
					Postcode: "B14 7ED",
					Country:  "GB",
				},
			},
			returnCodes: []ReturnCodeInfo{{Code: "A"}},
		},
		"with return code 'P'": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), privateKey),
			userData: identity.UserData{
				Status:      identity.StatusConfirmed,
				FirstNames:  "Alice Jane Laura",
				LastName:    "Doe",
				DateOfBirth: date.New("1970", "01", "02"),
				CheckedAt:   issuedAt,
				CurrentAddress: place.Address{
					Line1:    "1 Fake Road",
					Postcode: "B14 7ED",
					Country:  "GB",
				},
			},
			returnCodes: []ReturnCodeInfo{{Code: "P"}},
		},
		"missing": {
			error: ErrMissingCoreIdentityJWT,
		},
		"without name": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed, CheckedAt: testNow},
		},
		"without dob": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc": map[string]any{
					"credentialSubject": map[string]any{
						"name": namePart,
					},
				},
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed, CheckedAt: testNow},
		},
		"with invalid dob": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc": map[string]any{
					"credentialSubject": map[string]any{
						"name": namePart,
						"birthDate": []map[string]any{
							{
								"value": "1970-100-02",
							},
						},
					},
				},
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed, CheckedAt: testNow},
		},
		"without iat": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"vc": vc,
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed, CheckedAt: testNow},
		},
		"with unexpected signing method": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), []byte("a key")),
			error: jwt.ErrTokenUnverifiable,
		},
		"with malformed token": {
			token: "what token",
			error: jwt.ErrTokenMalformed,
		},
		"with invalid token": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": time.Now().Add(time.Minute).Unix(),
				"vc":  vc,
			}), privateKey),
			error: jwt.ErrTokenInvalidClaims,
		},
		"missing header kid": {
			token: missingKIDHeaderToken,
			error: jwt.ErrTokenUnverifiable,
		},
		"kid not a string": {
			token: kidNotAStringToken(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), privateKey),
			error: jwt.ErrTokenUnverifiable,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			userInfo := UserInfo{
				CoreIdentityJWT: tc.token,
				Addresses: []credentialAddress{{
					UPRN:           json.Number("456"),
					BuildingNumber: "2",
					StreetName:     "Fake Road",
					PostalCode:     "B14 7ED",
					AddressCountry: "GB",
					ValidFrom:      "2019-01-01",
					ValidUntil:     "2019-31-12",
				}, {
					UPRN:           json.Number("123"),
					BuildingNumber: "1",
					StreetName:     "Fake Road",
					PostalCode:     "B14 7ED",
					AddressCountry: "GB",
					ValidFrom:      "2020-01-01",
					ValidUntil:     "",
				}},
				ReturnCodes: tc.returnCodes,
			}

			userData, err := c.ParseIdentityClaim(userInfo)
			assert.ErrorIs(t, err, tc.error)
			assert.Equal(t, tc.userData, userData)
		})
	}
}

func TestParseIdentityClaimWhenDIDClientErrors(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	issuedAt := time.Now().Add(-time.Minute).Round(time.Second)

	c := &Client{
		didClient: &didClient{
			controllerID: "blah-not-matching",
		},
	}

	vc := map[string]any{
		"credentialSubject": map[string]any{},
	}

	mustSign := func(token *jwt.Token, key any) string {
		token.Header[jwkset.HeaderKID] = "blah#thing"
		s, err := token.SignedString(key)

		assert.Nil(t, err)
		return s
	}

	token := mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"vc":  vc,
	}), privateKey)

	userInfo := UserInfo{
		CoreIdentityJWT: token,
	}

	_, err := c.ParseIdentityClaim(userInfo)
	assert.ErrorContains(t, err, "could not find jwk for kid")
}

func TestParseIdentityClaimWithNonPassReturnCode(t *testing.T) {
	testcases := map[string]struct {
		returnCodes    []ReturnCodeInfo
		identityStatus identity.Status
		error          error
	}{
		"D": {
			returnCodes:    []ReturnCodeInfo{{Code: "D"}},
			identityStatus: identity.StatusFailed,
		},
		"N": {
			returnCodes:    []ReturnCodeInfo{{Code: "N"}},
			identityStatus: identity.StatusFailed,
		},
		"T": {
			returnCodes:    []ReturnCodeInfo{{Code: "T"}},
			identityStatus: identity.StatusFailed,
		},
		"V": {
			returnCodes:    []ReturnCodeInfo{{Code: "V"}},
			identityStatus: identity.StatusFailed,
		},
		"X": {
			returnCodes:    []ReturnCodeInfo{{Code: "X"}},
			identityStatus: identity.StatusInsufficientEvidence,
		},
		"Z": {
			returnCodes:    []ReturnCodeInfo{{Code: "Z"}},
			identityStatus: identity.StatusFailed,
		},
		"A + fail code": {
			returnCodes:    []ReturnCodeInfo{{Code: "A"}, {Code: "D"}},
			identityStatus: identity.StatusFailed,
		},
		"P + fail code": {
			returnCodes:    []ReturnCodeInfo{{Code: "P"}, {Code: "D"}},
			identityStatus: identity.StatusFailed,
		},
		"X + fail code": {
			returnCodes:    []ReturnCodeInfo{{Code: "X"}, {Code: "D"}},
			identityStatus: identity.StatusFailed,
		},
		"A + P": {
			returnCodes:    []ReturnCodeInfo{{Code: "A"}, {Code: "P"}},
			identityStatus: identity.StatusInsufficientEvidence,
		},
		"P + A": {
			returnCodes:    []ReturnCodeInfo{{Code: "P"}, {Code: "A"}},
			identityStatus: identity.StatusInsufficientEvidence,
		},
		"unexpected code": {
			returnCodes:    []ReturnCodeInfo{{Code: "NOT A CODE"}},
			identityStatus: identity.StatusUnknown,
			error:          ErrUnexpectedReturnCode,
		},
	}

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	c := &Client{
		didClient: &didClient{
			controllerID: "blah",
			assertionMethods: map[string]crypto.PublicKey{
				"blah#thing": &privateKey.PublicKey,
			},
		},
		now: testNowFn,
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			userInfo := UserInfo{
				ReturnCodes: tc.returnCodes,
			}

			userData, err := c.ParseIdentityClaim(userInfo)

			assert.Equal(t, tc.error, err)
			assert.Equal(t, tc.identityStatus, userData.Status)
		})
	}
}

func TestCredentialAddressTransformToAddress(t *testing.T) {
	testCases := map[string]struct {
		ca   credentialAddress
		want place.Address
	}{
		"building number no building name": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "",
				BuildingNumber:           "1",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "1 MELTON ROAD", Line2: "", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name no building number": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "1A",
				BuildingNumber:           "",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "1A", Line2: "MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name and building number": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "MELTON HOUSE",
				BuildingNumber:           "2",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "MELTON HOUSE", Line2: "2 MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building number": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "",
				BuildingNumber:           "3",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "KINGS HEATH",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "3 MELTON ROAD", Line2: "KINGS HEATH", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building name": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "MELTON HOUSE",
				BuildingNumber:           "",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "KINGS HEATH",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "MELTON HOUSE", Line2: "MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"dependent locality building name and building number": {
			ca: credentialAddress{
				UPRN:                     json.Number("123"),
				BuildingName:             "MELTON HOUSE",
				BuildingNumber:           "5",
				StreetName:               "MELTON ROAD",
				DependentAddressLocality: "KINGS HEATH",
				AddressLocality:          "BIRMINGHAM",
				PostalCode:               "B14 7ET",
			},
			want: place.Address{Line1: "MELTON HOUSE", Line2: "5 MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET", Country: "GB"},
		},
		"building name and sub building name": {
			ca: credentialAddress{
				UPRN:            json.Number("123"),
				SubBuildingName: "APARTMENT 34",
				BuildingName:    "CHARLES HOUSE",
				StreetName:      "PARK ROW",
				AddressLocality: "NOTTINGHAM",
				PostalCode:      "NG1 6GR",
			},
			want: place.Address{Line1: "APARTMENT 34, CHARLES HOUSE", Line2: "PARK ROW", TownOrCity: "NOTTINGHAM", Postcode: "NG1 6GR", Country: "GB"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ca.transformToAddress())
		})
	}
}
