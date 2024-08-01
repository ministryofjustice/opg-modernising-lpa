package onelogin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

var ErrMissingCoreIdentityJWT = errors.New("UserInfo missing CoreIdentityJWT property")

type UserInfo struct {
	Sub             string              `json:"sub"`
	Email           string              `json:"email"`
	EmailVerified   bool                `json:"email_verified"`
	Phone           string              `json:"phone"`
	PhoneVerified   bool                `json:"phone_verified"`
	UpdatedAt       int                 `json:"updated_at"`
	CoreIdentityJWT string              `json:"https://vocab.account.gov.uk/v1/coreIdentityJWT"`
	ReturnCodes     []ReturnCodeInfo    `json:"https://vocab.account.gov.uk/v1/returnCode,omitempty"`
	Addresses       []credentialAddress `json:"https://vocab.account.gov.uk/v1/address,omitempty"`
}

type ReturnCodeInfo struct {
	Code string `json:"code"`
}

type CoreIdentityClaims struct {
	jwt.RegisteredClaims

	Vot string     `json:"vot"`
	Vtm string     `json:"vtm"`
	Vc  Credential `json:"vc"`
}

type Credential struct {
	Type              []string          `json:"type"`
	CredentialSubject CredentialSubject `json:"credentialSubject"`
}

type CredentialSubject struct {
	Names []CredentialName `json:"name"`

	// BirthDate may list multiple values if there’s evidence an incorrect date of
	// birth was previously recorded for your user. The date of birth GOV.UK One
	// Login has highest confidence in will be the first item in the list.
	BirthDate []CredentialBirthDate `json:"birthDate"`
}

func (s CredentialSubject) CurrentNameParts() []NamePart {
	for _, name := range s.Names {
		if time.Time(name.ValidUntil).IsZero() {
			return name.NameParts
		}
	}

	return nil
}

type CredentialName struct {
	// ValidFrom shows when a name started to be used. If the zero value then the
	// user may have used that name from birth.
	ValidFrom Date `json:"validFrom"`

	// ValidUntil shows when the name ceased to be used. If the zero value then
	// this is the current name.
	ValidUntil Date `json:"validUntil"`

	// NameParts contains the components of the name in any order. The order of
	// names may depend on either your user’s preferences or the order they appear
	// on documents used to prove your user’s identity.
	NameParts []NamePart `json:"nameParts"`
}

type CredentialBirthDate struct {
	Value date.Date `json:"value"`
}

type credentialAddress struct {
	UPRN                           json.Number `json:"uprn"`
	SubBuildingName                string      `json:"subBuildingName"`
	BuildingName                   string      `json:"buildingName"`
	BuildingNumber                 string      `json:"buildingNumber"`
	DependentStreetName            string      `json:"dependentStreetName"`
	StreetName                     string      `json:"streetName"`
	DoubleDependentAddressLocality string      `json:"doubleDependentAddressLocality"`
	DependentAddressLocality       string      `json:"dependentAddressLocality"`
	AddressLocality                string      `json:"addressLocality"`
	PostalCode                     string      `json:"postalCode"`
	AddressCountry                 string      `json:"addressCountry"`
	ValidFrom                      string      `json:"validFrom"`
	ValidUntil                     string      `json:"validUntil"`
}

func (a credentialAddress) transformToAddress() place.Address {
	ad := place.AddressDetails{
		SubBuildingName:   a.SubBuildingName,
		BuildingName:      a.BuildingName,
		BuildingNumber:    a.BuildingNumber,
		ThoroughFareName:  a.StreetName,
		DependentLocality: a.DependentAddressLocality,
		Town:              a.AddressLocality,
		Postcode:          a.PostalCode,
	}

	return ad.TransformToAddress()
}

type NamePart struct {
	Value string `json:"value"`

	// Type is either 'GivenName' or 'FamilyName'
	Type string `json:"type"`
}

type Date time.Time

func (d *Date) UnmarshalText(text []byte) error {
	t, err := time.Parse("2006-01-02", string(text))
	*d = Date(t)
	return err
}

func (c *Client) UserInfo(ctx context.Context, idToken string) (UserInfo, error) {
	endpoint, err := c.openidConfiguration.UserinfoEndpoint()
	if err != nil {
		return UserInfo{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Add("Authorization", "Bearer "+idToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}

	defer res.Body.Close()

	var userinfoResponse UserInfo
	err = json.NewDecoder(res.Body).Decode(&userinfoResponse)

	return userinfoResponse, err
}

func (c *Client) ParseIdentityClaim(ctx context.Context, u UserInfo) (identity.UserData, error) {
	if len(u.ReturnCodes) > 0 {
		for _, c := range u.ReturnCodes {
			if c.Code == "X" {
				return identity.UserData{Status: identity.StatusInsufficientEvidence}, nil
			}
		}

		return identity.UserData{Status: identity.StatusFailed}, nil
	}

	publicKey, err := c.identityPublicKeyFunc(ctx)
	if err != nil {
		return identity.UserData{}, err
	}

	if u.CoreIdentityJWT == "" {
		return identity.UserData{}, ErrMissingCoreIdentityJWT
	}

	token, err := jwt.ParseWithClaims(u.CoreIdentityJWT, &CoreIdentityClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("signing method %v is invalid", token.Header["alg"])
		}

		return publicKey, nil
	}, jwt.WithIssuedAt())
	if err != nil {
		return identity.UserData{}, err
	}
	if !token.Valid {
		return identity.UserData{}, errors.New("jwt not valid")
	}

	claims := token.Claims.(*CoreIdentityClaims)

	currentName := claims.Vc.CredentialSubject.CurrentNameParts()
	if len(currentName) == 0 || claims.IssuedAt == nil {
		return identity.UserData{Status: identity.StatusFailed}, nil
	}

	var givenName, familyName []string
	for _, part := range currentName {
		if part.Type == "GivenName" {
			givenName = append(givenName, part.Value)
		} else {
			familyName = append(familyName, part.Value)
		}
	}

	birthDates := claims.Vc.CredentialSubject.BirthDate
	if len(birthDates) == 0 || !birthDates[0].Value.Valid() {
		return identity.UserData{Status: identity.StatusFailed}, nil
	}

	var currentAddress credentialAddress
	for _, a := range u.Addresses {
		if a.ValidUntil == "" {
			currentAddress = a
			break
		}
	}

	return identity.UserData{
		Status:         identity.StatusConfirmed,
		FirstNames:     strings.Join(givenName, " "),
		LastName:       strings.Join(familyName, " "),
		DateOfBirth:    birthDates[0].Value,
		RetrievedAt:    claims.IssuedAt.Time,
		CurrentAddress: currentAddress.transformToAddress(),
	}, nil
}
