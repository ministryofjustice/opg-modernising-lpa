package task

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProgressNotification struct {
	Heading string
	Body    string
}

//go:generate enumerator -type NotificationSubject -empty -trimprefix
type NotificationSubject uint8

const (
	NotificationSuccessfulVouch NotificationSubject = iota
)

func (n NotificationSubjects) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	ns := make([]string, len(n))
	for i, v := range n {
		ns[i] = strconv.Itoa(int(v))
	}

	return &types.AttributeValueMemberNS{
		Value: ns,
	}, nil
}

type NotificationSubjects []NotificationSubject

func (n NotificationSubjects) HasSeen(subject NotificationSubject) bool {
	for _, s := range n {
		if s == subject {
			return true
		}
	}

	return false
}
