// Code generated by "enumerator -type Channel -linecomment -empty -trimprefix"; DO NOT EDIT.

package lpadata

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ChannelPaper-1]
	_ = x[ChannelOnline-2]
}

const _Channel_name = "paperonline"

var _Channel_index = [...]uint8{0, 5, 11}

func (i Channel) String() string {
	if i == 0 {
		return ""
	}
	i -= 1
	if i >= Channel(len(_Channel_index)-1) {
		return "Channel(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Channel_name[_Channel_index[i]:_Channel_index[i+1]]
}

func (i Channel) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Channel) UnmarshalText(text []byte) error {
	val, err := ParseChannel(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Channel) IsPaper() bool {
	return i == ChannelPaper
}

func (i Channel) IsOnline() bool {
	return i == ChannelOnline
}

func ParseChannel(s string) (Channel, error) {
	switch s {
	case "":
		return Channel(0), nil
	case "paper":
		return ChannelPaper, nil
	case "online":
		return ChannelOnline, nil
	default:
		return Channel(0), fmt.Errorf("invalid Channel '%s'", s)
	}
}

type ChannelOptions struct {
	Paper  Channel
	Online Channel
}

var ChannelValues = ChannelOptions{
	Paper:  ChannelPaper,
	Online: ChannelOnline,
}

func (i Channel) Empty() bool {
	return i == Channel(0)
}
