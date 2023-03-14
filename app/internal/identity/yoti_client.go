package identity

import (
	"time"

	"github.com/getyoti/yoti-go-sdk/v3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type YotiClient struct {
	yoti       *yoti.Client
	scenarioID string
}

func NewYotiClient(scenarioID, clientID string, privateKeyBytes []byte) (*YotiClient, error) {
	if clientID == "" {
		return &YotiClient{scenarioID: scenarioID}, nil
	}

	client, err := yoti.NewClient(clientID, privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &YotiClient{yoti: client, scenarioID: scenarioID}, nil
}

func (c *YotiClient) ScenarioID() string {
	return c.scenarioID
}

func (c *YotiClient) SdkID() string {
	return c.yoti.SdkID
}

func (c *YotiClient) IsTest() bool {
	return c.yoti == nil
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

	details, err := c.yoti.GetActivityDetails(token)
	if err != nil {
		return UserData{}, err
	}

	dateOfBirth, err := details.UserProfile.DateOfBirth()
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
