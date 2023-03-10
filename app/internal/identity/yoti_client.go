package identity

import (
	"time"

	"github.com/getyoti/yoti-go-sdk/v3"
	"github.com/getyoti/yoti-go-sdk/v3/profile"
	"github.com/getyoti/yoti-go-sdk/v3/profile/sandbox"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

const yotiSandboxBaseURL = "https://api.yoti.com/sandbox/v1"

type YotiClient struct {
	yoti      *yoti.Client
	isSandbox bool
	details   profile.ActivityDetails
}

func NewYotiClient(clientID string, privateKeyBytes []byte) (*YotiClient, error) {
	if clientID == "" {
		return &YotiClient{}, nil
	}

	client, err := yoti.NewClient(clientID, privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &YotiClient{yoti: client}, nil
}

func (c *YotiClient) SetupSandbox() error {
	sandboxClient := &sandbox.Client{ClientSdkID: c.yoti.SdkID, Key: c.yoti.Key, BaseURL: yotiSandboxBaseURL}

	tokenRequest := (&sandbox.TokenRequest{}).
		WithFullName("Test Person", nil)

	sandboxToken, err := sandboxClient.SetupSharingProfile(tokenRequest)
	if err != nil {
		return err
	}

	c.yoti.OverrideAPIURL(yotiSandboxBaseURL)

	details, err := c.yoti.GetActivityDetails(sandboxToken)
	c.isSandbox = true
	c.details = details

	return err
}

func (c *YotiClient) SdkID() string {
	return c.yoti.SdkID
}

func (c *YotiClient) IsTest() bool {
	return c.yoti == nil || c.isSandbox
}

func (c *YotiClient) User(token string) (UserData, error) {
	if c.yoti == nil {
		return UserData{
			OK:          true,
			Provider:    EasyID,
			FirstNames:  "Test",
			LastName:    "Person",
			DateOfBirth: date.New("2000", "1", "2"),
			RetrievedAt: time.Now(),
		}, nil
	}

	if c.isSandbox {
		dateOfBirth, err := c.details.UserProfile.DateOfBirth()
		if err != nil {
			return UserData{}, err
		}

		return UserData{
			OK:          true,
			Provider:    EasyID,
			FirstNames:  c.details.UserProfile.GivenNames().Value(),
			LastName:    c.details.UserProfile.FamilyName().Value(),
			DateOfBirth: date.FromTime(*dateOfBirth.Value()),
			RetrievedAt: time.Now(),
		}, nil
	}

	details, err := c.yoti.GetActivityDetails(token)
	if err != nil {
		return UserData{}, err
	}

	dateOfBirth, err := c.details.UserProfile.DateOfBirth()
	if err != nil {
		return UserData{}, err
	}

	return UserData{
		OK:          true,
		Provider:    EasyID,
		FirstNames:  details.UserProfile.GivenNames().Value(),
		LastName:    details.UserProfile.FamilyName().Value(),
		DateOfBirth: date.FromTime(*dateOfBirth.Value()),
		RetrievedAt: time.Now(),
	}, nil
}
