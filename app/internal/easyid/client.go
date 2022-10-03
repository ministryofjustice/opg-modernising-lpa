package easyid

import (
	"github.com/getyoti/yoti-go-sdk/v3"
	"github.com/getyoti/yoti-go-sdk/v3/profile"
	"github.com/getyoti/yoti-go-sdk/v3/profile/sandbox"
)

const sandboxBaseURL = "https://api.yoti.com/sandbox/v1"

type UserData struct {
	FullName string
}

type Client struct {
	yoti      *yoti.Client
	isSandbox bool
	details   profile.ActivityDetails
}

func New(clientID string, privateKeyBytes []byte) (*Client, error) {
	if clientID == "" {
		return &Client{}, nil
	}

	client, err := yoti.NewClient(clientID, privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &Client{yoti: client}, nil
}

func (c *Client) SetupSandbox() error {
	sandboxClient := &sandbox.Client{ClientSdkID: c.yoti.SdkID, Key: c.yoti.Key, BaseURL: sandboxBaseURL}

	tokenRequest := (&sandbox.TokenRequest{}).
		WithRememberMeID("remember_me_id_12345").
		WithGivenNames("some given names", nil).
		WithFamilyName("some family name", nil).
		WithFullName("some full name", nil).
		WithGender("some gender", nil).
		WithPhoneNumber("some phone number", nil).
		WithNationality("some nationality", nil).
		WithStructuredPostalAddress(
			map[string]interface{}{
				"building_number": "1",
				"address_line1":   "some street name",
			}, nil).
		WithEmailAddress("some@email", nil).
		WithDocumentDetails("PASSPORT USA 1234abc", nil)

	sandboxToken, err := sandboxClient.SetupSharingProfile(tokenRequest)
	if err != nil {
		return err
	}

	c.yoti.OverrideAPIURL(sandboxBaseURL)

	details, err := c.yoti.GetActivityDetails(sandboxToken)
	c.isSandbox = true
	c.details = details

	return err
}

func (c *Client) SdkID() string {
	return c.yoti.SdkID
}

func (c *Client) IsTest() bool {
	return c.yoti == nil || c.isSandbox
}

func (c *Client) User(token string) (UserData, error) {
	if c.yoti == nil {
		return UserData{FullName: "Test Person"}, nil
	}

	if c.isSandbox {
		return UserData{FullName: c.details.UserProfile.FullName().Value()}, nil
	}

	details, err := c.yoti.GetActivityDetails(token)
	if err != nil {
		return UserData{}, err
	}

	return UserData{FullName: details.UserProfile.FullName().Value()}, nil
}
