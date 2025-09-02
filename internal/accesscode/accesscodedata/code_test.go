package accesscodedata

import (
	"bytes"
	"log/slog"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	plain, hashed := Generate("Smith")

	s := plain.Plain()
	assert.Equal(t, HashedFromString(s[:4]+s[5:], "Smith"), hashed)
}

func TestPlainText(t *testing.T) {
	plain := PlainText("abcdefgh")

	assert.Equal(t, "<accesscode>", plain.String())
	assert.Equal(t, "<accesscode>", plain.GoString())

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	logger.Info("hey", slog.Any("code", plain))
	assert.NotContains(t, buf.String(), "abc")
	assert.Contains(t, buf.String(), "code=<accesscode>")

	assert.Equal(t, "abcd-efgh", plain.Plain())
}

func TestHashed(t *testing.T) {
	hash := Hashed([32]byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
	})

	assert.Equal(t, "6162636465666768616263646566676861626364656667686162636465666768", hash.String())
	assert.Equal(t, url.Values{
		"code": {"6162636465666768616263646566676861626364656667686162636465666768"},
	}, hash.Query())
}

func TestHashedFromString(t *testing.T) {
	assert.Equal(t, Hashed([32]byte{
		0x6e, 0x94, 0x58, 0x29, 0x44, 0xab, 0x45, 0x63, 0x79, 0xe6, 0x98, 0x9e, 0x7f, 0x4d, 0xcb, 0x9b,
		0x18, 0xb5, 0x6e, 0x49, 0x78, 0x9e, 0x48, 0x42, 0xae, 0x31, 0x10, 0x15, 0xb9, 0xc5, 0x29, 0x14,
	}), HashedFromString("hello", "Smith"))
}

func TestHashedFromQuery(t *testing.T) {
	assert.Equal(t, Hashed([32]byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
	}), HashedFromQuery(url.Values{"code": {"6162636465666768616263646566676861626364656667686162636465666768"}}))
}

func TestHashedFromQueryWhenNotLength(t *testing.T) {
	assert.Equal(t, Hashed([32]byte{}), HashedFromQuery(url.Values{"code": {"162636465666768616263646566676861626364656667686162636465666768"}}))
}
