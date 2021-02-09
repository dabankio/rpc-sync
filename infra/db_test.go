package infra

import (
	"fmt"
	"strings"
	"testing"
)

func TestEmbedSQL(t *testing.T) {
	for _, sql := range strings.Split(schemaSQL, ";") {
		if strings.HasPrefix(strings.TrimSpace(sql), "--") {
			continue
		}
		fmt.Println(sql)
	}
}
