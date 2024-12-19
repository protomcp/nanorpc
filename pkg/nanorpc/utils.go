package nanorpc

import "strings"

func writeString(buf *strings.Builder, ss ...string) {
	for _, s := range ss {
		_, _ = buf.WriteString(s)
	}
}
