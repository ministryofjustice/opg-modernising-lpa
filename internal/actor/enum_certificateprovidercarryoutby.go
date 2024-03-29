// Code generated by "enumerator -type CertificateProviderCarryOutBy -linecomment -empty"; DO NOT EDIT.
package actor

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Paper-1]
	_ = x[Online-2]
}

const _CertificateProviderCarryOutBy_name = "paperonline"

var _CertificateProviderCarryOutBy_index = [...]uint8{0, 5, 11}

func (i CertificateProviderCarryOutBy) String() string {
	i -= 1
	if i >= CertificateProviderCarryOutBy(len(_CertificateProviderCarryOutBy_index)-1) {
		return "CertificateProviderCarryOutBy(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _CertificateProviderCarryOutBy_name[_CertificateProviderCarryOutBy_index[i]:_CertificateProviderCarryOutBy_index[i+1]]
}

func (i CertificateProviderCarryOutBy) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *CertificateProviderCarryOutBy) UnmarshalText(text []byte) error {
	val, err := ParseCertificateProviderCarryOutBy(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i CertificateProviderCarryOutBy) IsPaper() bool {
	return i == Paper
}

func (i CertificateProviderCarryOutBy) IsOnline() bool {
	return i == Online
}

func ParseCertificateProviderCarryOutBy(s string) (CertificateProviderCarryOutBy, error) {
	switch s {
	case "paper":
		return Paper, nil
	case "online":
		return Online, nil
	default:
		return CertificateProviderCarryOutBy(0), fmt.Errorf("invalid CertificateProviderCarryOutBy '%s'", s)
	}
}

type CertificateProviderCarryOutByOptions struct {
	Paper  CertificateProviderCarryOutBy
	Online CertificateProviderCarryOutBy
}

var CertificateProviderCarryOutByValues = CertificateProviderCarryOutByOptions{
	Paper:  Paper,
	Online: Online,
}

func (i CertificateProviderCarryOutBy) Empty() bool {
	return i == CertificateProviderCarryOutBy(0)
}
