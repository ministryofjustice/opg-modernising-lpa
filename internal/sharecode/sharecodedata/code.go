package sharecodedata

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

func Generate() (PlainText, Hashed) {
	plain := random.Friendly(8)

	return PlainText(plain), HashedFromString(plain)
}

type PlainText string

func (PlainText) String() string {
	return "<sharecode>"
}

func (PlainText) GoString() string {
	return "<sharecode>"
}

func (PlainText) LogValue() slog.Value {
	return slog.StringValue("<sharecode>")
}

func (p PlainText) Plain() string {
	s := string(p)

	return s[:4] + "-" + s[4:]
}

type Hashed [32]byte

func (h Hashed) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hashed) Query() url.Values {
	return url.Values{"code": {h.String()}}
}

func HashedFromString(plain string) Hashed {
	var s []byte
	for _, b := range []byte(plain) {
		if b != ' ' && b != '\t' && b != '-' {
			s = append(s, b)
		}
	}

	hash := sha256.Sum256(s)

	return Hashed(hash)
}

func HashedFromQuery(q url.Values) Hashed {
	b, _ := hex.DecodeString(q.Get("code"))
	if len(b) != 32 {
		return Hashed([32]byte{})
	}

	return Hashed(b)
}
