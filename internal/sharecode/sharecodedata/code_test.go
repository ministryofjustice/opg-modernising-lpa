package sharecodedata

import (
	"bytes"
	"log/slog"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	plain, hashed := Generate()

	s := plain.Plain()
	assert.Equal(t, HashedFromString(s[:4]+s[5:]), hashed)
}

func TestPlainText(t *testing.T) {
	plain := PlainText("abcdefgh")

	assert.Equal(t, "<sharecode>", plain.String())
	assert.Equal(t, "<sharecode>", plain.GoString())

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	logger.Info("hey", slog.Any("code", plain))
	assert.NotContains(t, buf.String(), "abc")
	assert.Contains(t, buf.String(), "code=<sharecode>")

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
		0x2c, 0xf2, 0x4d, 0xba, 0x5f, 0xb0, 0xa3, 0xe,
		0x26, 0xe8, 0x3b, 0x2a, 0xc5, 0xb9, 0xe2, 0x9e,
		0x1b, 0x16, 0x1e, 0x5c, 0x1f, 0xa7, 0x42, 0x5e,
		0x73, 0x4, 0x33, 0x62, 0x93, 0x8b, 0x98, 0x24,
	}), HashedFromString("hello"))
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
