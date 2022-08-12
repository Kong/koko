package kong

import (
	"fmt"
	"strings"
	"text/template"
)

var adminServerConf = []byte(`
server {
  listen 8001;
  location / {
    default_type application/json;
    content_by_lua_block {
      Kong.admin_content()
    }
    header_filter_by_lua_block {
      Kong.admin_header_filter()
    }
  }
}
`)

var scriptTemplate = `#!/bin/bash
set -e 
set -x
cleanup () {
  echo "interrupt received, exiting now"
  docker rm -f {{ $.ContainerName }} 
  exit 0
}
DIR=$(dirname "$0")
trap cleanup INT
docker run \
  --rm \
  --name {{ $.ContainerName }} \
{{- range $k, $v := $.EnvVars }}
  -e {{ $k }}={{ sh $v }} \
{{- end -}}
{{- range $host, $container := $.Internal.Volumes }}
  -v "$DIR/{{ $host }}:{{ $container }}" \
{{- end -}}
{{- range $host, $container := .Internal.Ports }}
  -p {{ $host }}:{{ $container }} \
{{- end -}}
{{- range $hostname, $ip := .Internal.Hosts }}
  --add-host {{ $hostname }}:{{ $ip }} \
{{- end -}}
{{- if .Network }}
  --network {{ .Network }} \
{{- end }}
  {{ .Image }} &
wait
`

var t *template.Template

func init() {
	t = template.Must(template.New("run").Funcs(template.FuncMap{
		"sh": shellEscaper,
	}).Parse(scriptTemplate))
}

// shellEscaper handles escaping shell input, mostly used for handling of environment variables.
func shellEscaper(args ...any) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected a single string as input for shell escape function")
	}

	s, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("expected string input for shell escape function, but got %T", args[0])
	}

	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'", nil
}
