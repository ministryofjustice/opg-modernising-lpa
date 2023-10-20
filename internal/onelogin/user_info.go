package onelogin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
)

type UserInfo struct {
	Sub             string `json:"sub"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	Phone           string `json:"phone"`
	PhoneVerified   bool   `json:"phone_verified"`
	UpdatedAt       int    `json:"updated_at"`
	CoreIdentityJWT string `json:"https://vocab.account.gov.uk/v1/coreIdentityJWT"`
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
	req, err := http.NewRequestWithContext(ctx, "GET", c.openidConfiguration.UserinfoEndpoint, nil)
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
	publicKeyBytes, err := c.secretsClient.SecretBytes(ctx, secrets.GovUkOneLoginIdentityPublicKey)
	if err != nil {
		return identity.UserData{}, err
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return identity.UserData{}, err
	}

	if u.CoreIdentityJWT == "" {
		return identity.UserData{}, errors.New("UserInfo missing CoreIdentityJWT property")
	}

	token, err := jwt.ParseWithClaims(u.CoreIdentityJWT, &CoreIdentityClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, jwt.NewValidationError(fmt.Sprintf("signing method %v is invalid", token.Header["alg"]), jwt.ValidationErrorSignatureInvalid)
		}

		return publicKey, nil
	})
	if err != nil {
		return identity.UserData{}, err
	}
	if !token.Valid {
		return identity.UserData{}, errors.New("jwt not valid")
	}

	claims := token.Claims.(*CoreIdentityClaims)

	currentName := claims.Vc.CredentialSubject.CurrentNameParts()
	if len(currentName) == 0 || claims.IssuedAt == nil {
		return identity.UserData{OK: false}, nil
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
		return identity.UserData{OK: false}, nil
	}

	return identity.UserData{
		OK:          true,
		FirstNames:  strings.Join(givenName, " "),
		LastName:    strings.Join(familyName, " "),
		DateOfBirth: birthDates[0].Value,
		RetrievedAt: claims.IssuedAt.Time,
	}, nil
}
