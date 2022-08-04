package clickhouse

import (
	"bytes"
	"fmt"
)

func escape(identifier string) string {
	return escapeBytes([]byte(identifier))
}

func escapeBytes(identifier []byte) string {
	var (
		escapeMap = map[byte]string{
			0:    "\\0",
			'\b': "\\b",
			'\f': "\\f",
			'\r': "\\r",
			'\n': "\\n",
			'\t': "\\t",
			'\\': "\\\\",
			'`':  "\\`",
		}
	)

	buf := new(bytes.Buffer)

	buf.WriteByte('`')

	for i := range identifier {
		b := identifier[i]

		escaped, ok := escapeMap[b]
		if ok {
			buf.WriteString(escaped)
		} else if b < 0x20 || b > 0x7e {
			buf.WriteString(fmt.Sprintf("\\x%x", b))
		} else {
			buf.WriteByte(b)
		}
	}

	buf.WriteByte('`')

	return buf.String()
}
