package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeBytes(t *testing.T) {
	testdata := []struct {
		in  []byte
		out string
	}{
		{
			in:  []byte("O`sullivan"),
			out: "`O\\`sullivan`",
		},
		{
			in:  []byte("simple"),
			out: "`simple`",
		},
		{
			in:  []byte("random \x00 null byte"),
			out: "`random \\0 null byte`",
		},
		{
			in:  []byte{0xA3},
			out: "`\\xa3`",
		},
		{
			in: []byte("😀"),
			// GRINNING FACE is 0xF0 0x9F 0x98 0x80 in UTF 8
			out: "`\\xf0\\x9f\\x98\\x80`",
		},
	}

	for _, test := range testdata {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, test.out, escapeBytes(test.in))
		})
	}
}
