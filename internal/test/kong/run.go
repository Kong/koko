package kong

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/kong/koko/internal/test/certs"
)

var (
	newLineChar  = []byte{'\n'}
	stdoutPrefix = []byte("dp-stdout: ")
	stderrPrefix = []byte("dp-stderr: ")
)

type AuthMode string

var (
	AuthModeMTLSPKI    AuthMode = "pki"
	AuthModeMTLSShared AuthMode = "shared"
)

type Computed struct {
	CPHostname string
}

type DockerInput struct {
	Version    string
	EnvVars    map[string]string
	AuthMode   AuthMode
	ClientCert []byte
	ClientKey  []byte
	CACert     []byte
	CPAddr     string
	Computed   Computed
}

var defaultEnvVars = map[string]string{
	"KONG_VITALS":                 "off",
	"KONG_NGINX_WORKER_PROCESSES": "1",
	"KONG_CLUSTER_CONTROL_PLANE":  "localhost:3100",
	"KONG_ANONYMOUS_REPORTS":      "off",
	"KONG_NGINX_HTTP_INCLUDE":     "/conf/admin.conf",
}

var scripTemplate = `#!/bin/bash
set -x
cleanup () {
  echo "interrupt received, exiting now"
  docker rm -f koko-dp
}
DIR=$(dirname "$0")
trap cleanup SIGINT
docker rm -f koko-dp
docker run \
  --rm \
  --name koko-dp \
  -e "KONG_DATABASE=off" \
  -e "KONG_ROLE=data_plane" \
  -e "KONG_CLUSTER_CERT=/certs/cluster.crt" \
  -e "KONG_CLUSTER_CERT_KEY=/certs/cluster.key" \
  -e "KONG_LUA_SSL_TRUSTED_CERTIFICATE=/certs/cluster-ca.crt" \
  -e "KONG_CLUSTER_CA_CERT=/certs/cluster-ca.crt" \
{{- range $k, $v := $.EnvVars }}
  -e "{{ $k }}={{ $v }}" \
{{- end -}}
  -v "$DIR/cluster.crt:/certs/cluster.crt" \
  -v "$DIR/cluster-ca.crt:/certs/cluster-ca.crt" \
  -v "$DIR/cluster.key:/certs/cluster.key" \
  -v "$DIR/admin.conf:/conf/admin.conf" \
{{- if .Computed.CPHostname }}
  --add-host "{{- .Computed.CPHostname -}}:host-gateway" \
{{- end -}}
  --network host kong:{{ .Version }}
`

var t *template.Template

func init() {
	t = template.Must(template.New("run").Parse(scripTemplate))
}

func RunDP(ctx context.Context, input DockerInput) error {
	dockerInput := addDefaults(input)
	dir, err := os.MkdirTemp("", "koko-dp-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = setup(dir, dockerInput)
	if err != nil {
		return err
	}

	cmd := &exec.Cmd{
		Path: dir + "/run.sh",
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	const goroutineCount = 3
	wg.Add(goroutineCount)
	go func() {
		defer wg.Done()
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			_, _ = os.Stdout.Write(stdoutPrefix)
			_, _ = os.Stdout.Write(sc.Bytes())
			_, _ = os.Stdout.Write(newLineChar)
		}
	}()
	go func() {
		defer wg.Done()
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			_, _ = os.Stdout.Write(stderrPrefix)
			_, _ = os.Stdout.Write(sc.Bytes())
			_, _ = os.Stdout.Write(newLineChar)
		}
	}()
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = cmd.Process.Signal(os.Interrupt)
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}

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

func setup(dir string, input DockerInput) (err error) {
	err = os.WriteFile(dir+"/cluster.key", input.ClientKey, os.ModePerm)
	if err != nil {
		return
	}
	err = os.WriteFile(dir+"/cluster.crt", input.ClientCert, os.ModePerm)
	if err != nil {
		return
	}
	err = os.WriteFile(dir+"/cluster-ca.crt", input.CACert, os.ModePerm)
	if err != nil {
		return
	}
	err = os.WriteFile(dir+"/admin.conf", adminServerConf, os.ModePerm)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "run", &input)
	if err != nil {
		return
	}

	err = os.WriteFile(dir+"/run.sh", buf.Bytes(), os.ModePerm)
	if err != nil {
		return
	}
	return
}

func addDefaults(input DockerInput) DockerInput {
	res := input

	if res.Version == "" {
		panic("no version set")
	}
	if res.EnvVars == nil {
		res.EnvVars = map[string]string{}
	}
	for k, v := range defaultEnvVars {
		if _, ok := res.EnvVars[k]; !ok {
			res.EnvVars[k] = v
		}
	}
	if res.CPAddr != "" {
		res.EnvVars["KONG_CLUSTER_CONTROL_PLANE"] = res.CPAddr
		i := strings.Index(res.CPAddr, ":")
		if i == -1 {
			panic("incorrect")
		}
		res.Computed.CPHostname = res.CPAddr[:i]
	}
	return res
}

func GetKongConfForShared() DockerInput {
	kongVersion := os.Getenv("KOKO_TEST_KONG_DP_VERSION")
	if kongVersion == "" {
		panic("no KOKO_TEST_KONG_DP_VERSION set")
	}
	res := DockerInput{
		EnvVars: map[string]string{
			"KONG_CLUSTER_MTLS": "shared",
		},
		Version: kongVersion,
	}
	if testing.Verbose() {
		k := "KONG_LOG_LEVEL"
		v := "debug"
		if _, ok := res.EnvVars[k]; !ok {
			res.EnvVars[k] = v
		}
	}
	res.ClientKey = certs.DefaultSharedKey
	res.ClientCert = certs.DefaultSharedCert
	res.CACert = certs.DefaultSharedCert
	return res
}

func GetKongConf() DockerInput {
	res := GetKongConfForShared()
	res.EnvVars["KONG_CLUSTER_SERVER_NAME"] = "cp.example.com"
	res.EnvVars["KONG_CLUSTER_MTLS"] = "pki"
	res.CPAddr = "localhost:3100"
	res.CACert = certs.CPCACert
	res.ClientKey = certs.DPTree1Key
	res.ClientCert = certs.DPTree1Cert
	return res
}
