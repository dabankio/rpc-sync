package infra

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func TestEmbedSQL(t *testing.T) {
	for _, sql := range strings.Split(schemaSQL, ";") {
		if strings.HasPrefix(strings.TrimSpace(sql), "--") {
			continue
		}
		fmt.Println(sql)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{args: "Name", want: "name"},
		{args: "RealName", want: "real_name"},
		{args: "realName", want: "real_name"},
		{args: "ID", want: "id"},
		{args: "URL", want: "url"},
		{args: "HTTPHost", want: "http_host"},
		{args: "ServerHTTPHost", want: "server_http_host"},
		{args: "RealIP", want: "real_ip"},
		{args: "RealHTTPServerIP", want: "real_http_server_ip"},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			if gotRet := ToSnakeCase(tt.args); gotRet != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", gotRet, tt.want)
			}
		})
	}
}
