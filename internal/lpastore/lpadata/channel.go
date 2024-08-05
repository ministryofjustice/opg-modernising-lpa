package lpadata

//go:generate enumerator -type Channel -linecomment -empty -trimprefix
type Channel uint8

const (
	ChannelPaper  Channel = iota + 1 // paper
	ChannelOnline                    // online
)
