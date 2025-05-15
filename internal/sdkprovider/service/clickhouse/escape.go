package clickhouse

import (
	"bytes"
	"fmt"
)

func escape(identifier string) string {
	return escapeBytes([]byte(identifier))
}

func escapeBytes(identifier []byte) string {
	escapeMap := map[byte]string{
		0:    "\\0",
		'\b': "\\b",
		'\f': "\\f",
		'\r': "\\r",
		'\n': "\\n",
		'\t': "\\t",
		'\\': "\\\\",
		'`':  "\\`",
	}
	buf := new(bytes.Buffer)
	buf.WriteByte('`')

	for i := range identifier {
		b := identifier[i]

		escaped, ok := escapeMap[b]
		switch {
		case ok:
			buf.WriteString(escaped)
		case b < 0x20 || b > 0x7e:
			buf.WriteString(fmt.Sprintf("\\x%02x", b))
		default:
			buf.WriteByte(b)
		}
	}

	buf.WriteByte('`')
	return buf.String()
}
