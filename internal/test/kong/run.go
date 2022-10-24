package kong

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/kong/koko/internal/test/certs"
	"github.com/samber/lo"
)

// Various files that will be created on the DP.
const (
	fileAdminConf     = "/conf/admin.conf"
	fileClusterCACert = "/certs/cluster-ca.crt"
	fileClusterCert   = "/certs/cluster.crt"
	fileClusterKey    = "/certs/cluster.key"
)

// Secrets used to test secrets management.
const (
	secretCert = `-----BEGIN CERTIFICATE-----
MIIEczCCAlugAwIBAgIJAMw8/GAiHIFBMA0GCSqGSIb3DQEBCwUAMDYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRIwEAYDVQQDDAlsb2NhbGhvc3Qw
HhcNMjIxMDA0MTg1MjI5WhcNMjcxMDAzMTg1MjI5WjA2MQswCQYDVQQGEwJVUzET
MBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3LUv6RauFfFn4a2BNSTE5oQNhASBh2Lk
0Gd0tfPcmTzJbohFwyAGskYj0NBRxnRVZdLPeoZIQyYSaiPWyeDITnXyKk3Nh9Zk
xQ03YsbZCk9jIsp78/ECdnYCCS4dpYGswu8b37dxUta+6AhEEte73ezrAhc+ZIy5
2yjcix4P5+vfhBf0EzBT8D7z+wZjji3F/A969EqphFwPz3KudkTOR6d0bQEVbN3x
cg4lcj49RzwS4sPbq6ub52QrKcx8s+d9bqC/nhHLn1HM/eef+cxROedcIWZs5RvG
mk/H+K2lcKL33gIcXgzSeunobV+8xwwoYk4GZroXjavUkgelKKjQBQIDAQABo4GD
MIGAMFAGA1UdIwRJMEehOqQ4MDYxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApDYWxp
Zm9ybmlhMRIwEAYDVQQDDAlsb2NhbGhvc3SCCQCjgi452nKnUDAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIE8DAUBgNVHREEDTALgglsb2NhbGhvc3QwDQYJKoZIhvcNAQEL
BQADggIBAJiKfxuh2g0fYcR7s6iiiOMhT1ZZoXhyDhoHUUMlyN9Lm549PBQHOX6V
f/g+jqVawIrGLaQbTbPoq0XPTZohk4aYwSgU2Th3Ku8Q73FfO0MM1jjop3WCatNF
VZj/GBc9uVunOV5aMcMIs2dFU2fwH4ZjOwnv7rJRldoW1Qlk3uWeIktbmv4yZKvu
FWPuo3ks+7B+BniqKXuYkNcuhlE+iVr3kJ55qRgX1RxKo4CB3Tynkp7sikp4al7x
jlHSM9YAqvPFFMOhlU2U3SxE4CLasL99zP0ChINKp9XqzW/qo+F0/Jd4rZmddU2f
M9Cx62cc0L5IlsHLVJj3zwHZzc/ifpBUeebB98IjoQAfiRkbX0Oe/c2TxtR4o/RH
GWNeKCThdliZkXaLiOPswOV1BYfA00etorcY0CIy0aTaZgfvrYsJe4aT/hkF/JvF
tHJ/iD67m8RhLysRL/w50+quVMluUDqJps0HhKrB9wzNJWrddWhvplVeuOXEJfTM
i80W1JE4OApdISMEn56vRi+BMQMgIeYWznfyQnI4G3rUJQFMI5KzLxkvfYNTF3Ci
3Am0KaJ7X2oLOq4Qc6CM3qDkvvId61COlfJb2Fo3ETnoT66mxcb6bjtz1WTWOopm
UcmBKErRUKksINUxwuvP/VW007tXOjZH7wmiM2IW5LUZVkbhB1iE
-----END CERTIFICATE-----`

	secretKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEA3LUv6RauFfFn4a2BNSTE5oQNhASBh2Lk0Gd0tfPcmTzJbohF
wyAGskYj0NBRxnRVZdLPeoZIQyYSaiPWyeDITnXyKk3Nh9ZkxQ03YsbZCk9jIsp7
8/ECdnYCCS4dpYGswu8b37dxUta+6AhEEte73ezrAhc+ZIy52yjcix4P5+vfhBf0
EzBT8D7z+wZjji3F/A969EqphFwPz3KudkTOR6d0bQEVbN3xcg4lcj49RzwS4sPb
q6ub52QrKcx8s+d9bqC/nhHLn1HM/eef+cxROedcIWZs5RvGmk/H+K2lcKL33gIc
XgzSeunobV+8xwwoYk4GZroXjavUkgelKKjQBQIDAQABAoIBAQDcd/nmAwvfS4iT
vTgGmDZAdsTxjXa+gSFEtTO21mUUhc5JpcLaSdGmn74DRzWI4oiz8EPlhuIEgbF/
aVGT1AEDr3o6nAGloZqD5NHgz/XbALZs+IudgLEPGI6sEO74d3LWPvg/IAYJ1A5b
xnYJxIscAyA2tHVVB+ZYcJbuORd2eSKeZSjfEfX8/DN8sKD+4DK9g/GiVlOJBG/d
bSoZmcRv3HXpnSKTvCydkxbBliD/8H7jRvkCOi2VcYT9+08rucwXc6v+q9wiQ/b7
hPdBn6KqDKRO9HPZYVkztlsdHXnthq16QyNPOk2umggfyXMIPhYBcW/dZ5xNqxBD
KiInqjbBAoGBAP3s/FS8GvFJ80pwtA0AUtB7Lo3ZASSs9dq+Dr8binPpp/1qz3hJ
Q/gRC9EP63MOWA2PK22D4qsjTfrBpqxLaalZCbCJGDGT+2knTN+qsOJ3//qI5zjj
cFEcnWcJ3bI5eLAU/2GKViyXzdGlZxBbc4zKBUSyxMAUewr3PsqEO0SJAoGBAN6C
vEYAbNuCZagtDrhhGYb+8rbSKZA7K4QjJcLTyZZ+z4ohWRW9tVyEXkszIwCRrs3y
rhHJU+z1bJOXxIin1i5tCzVlG6iNLct9dm2Z5v4hbDGNw4HhIV0l+tXrVGhkM/1v
vbRhldQA0H9iwWx+bKNS3lVLeUvYu4udmzrY74idAoGBAJ/8zQ9mZWNZwJxKXmdC
qOsKcc6Vx46gG1dzID9wzs8xjNKylX2oS9bkhpl2elbH1trUNfyOeCZz3BH+KVGt
QimdG+nKtx+lqWYbiOfz1/cYvIPR9j11r7KrYNEm+jPs2gm3cSC31IvMKbXJjSJV
PHycXK1oJWcQgGXsWfenUOBhAoGBAKezvRa9Z04h/2A7ZWbNuCGosWHdD/pmvit/
Ggy29q54sQ8Yhz39l109Xpwq1GyvYCJUj6FULe7gIo8yyat9Y83l3ZbGt4vXq/Y8
fy+n2RMcOaE3iWywMyczYtQr45gyPYT73OzAx93bJ0l7MvEEb/jAklWS5r6lgOR/
SumVayN5AoGBALLaG16NDrV2L3u/xzxw7uy5b3prpEi4wgZd4i72XaK+dMqtNVYy
KlBs7O9y+fc4AIIn6JD+9tymB1TWEn1B+3Vv6jmtzbztuCQTbJ6rTT3CFcE6TdyJ
8rYuG3/p2VkcG29TWbQARtj5ewv9p5QNfaecUzN+tps89YzawWQBanwI
-----END RSA PRIVATE KEY-----`  // #nosec G101 -- ignore hardcoded test certs
)

// Internal script that exists on the host for running the DP in Docker.
const fileRunSh = "run.sh"

func RunDP(ctx context.Context, input DockerInput) error {
	dockerInput, err := addDefaults(&input)
	if err != nil {
		return err
	}

	dir, cleanup, err := createFiles(dockerInput)
	if err != nil {
		return err
	}
	defer cleanup() //nolint:errcheck

	// Ideally the container should always be cleaned up, but it may not always be, especially when
	// `go tool test2json` is being used, as it can't properly catch signals as of this writing.
	//
	// Read more:
	// - https://github.com/golang/go/pull/53506
	// - https://go-review.googlesource.com/c/go/+/419295
	if err := removeDockerContainer(input.ContainerName); err != nil {
		return err
	}

	cmd := &exec.Cmd{Path: path.Join(dir, fileRunSh)}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	const goroutineCount = 3
	wg.Add(goroutineCount)

	// Redirect logs from container to relevant output.
	go streamLogs(&wg, stdout, false)
	go streamLogs(&wg, stderr, true)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = cmd.Process.Signal(os.Interrupt)
	}()

	if err := cmd.Wait(); err != nil {
		// Ignore any errors due to the script exiting by result of a signal.
		if err, ok := err.(*exec.ExitError); ok {
			if status, ok := err.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				return nil
			}
		}

		// Script produced a non-zero exit code.
		return fmt.Errorf("unexpected error when running the DP, read logs for more detail: %w", err)
	}

	wg.Wait()
	return nil
}

func createFiles(input *DockerInput) (dir string, cleanup func() error, err error) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "run", &input); err != nil {
		return "", nil, err
	}

	// We're forcing the `koko-` prefix, as this is to ensure that there is
	// a common glob for all directories that are created via the tests.
	const tmpDirPrefix = "koko-"
	tmpDir := input.ContainerName + "-*"
	if !strings.HasPrefix(tmpDir, tmpDirPrefix) {
		tmpDir = tmpDirPrefix + tmpDir
	}

	if dir, err = os.MkdirTemp("", tmpDir); err != nil {
		return "", nil, err
	}

	cleanupFn := func() error { return os.RemoveAll(dir) }
	defer func() {
		if err != nil {
			_ = cleanupFn()
		}
	}()

	for v := range input.Internal.Volumes {
		if err = os.MkdirAll(path.Join(dir, filepath.Base(v)), os.ModePerm); err != nil {
			return "", nil, err
		}
	}

	for p, content := range map[string][]byte{
		path.Join(dir, fileAdminConf):     adminServerConf,
		path.Join(dir, fileClusterCACert): input.CACert,
		path.Join(dir, fileClusterCert):   input.ClientCert,
		path.Join(dir, fileClusterKey):    input.ClientKey,
		path.Join(dir, fileRunSh):         buf.Bytes(),
	} {
		if err = os.WriteFile(p, content, os.ModePerm); err != nil {
			return "", nil, err
		}
	}

	return dir, cleanupFn, err
}

// addDefaults generates the required input to the `docker run` command.
//
// By default, for the given host OSes, the Kong DP will run as follows:
// - macOS
//   - Environment variables:
//     1. `KONG_CLUSTER_CONTROL_PLANE=host.docker.internal:3100`
//   - Ports bound to host OS:
//     1. 8000:8000
//     2. 8001:8001
//
// - Linux
//   - Environment variables:
//     1. `KONG_CLUSTER_CONTROL_PLANE=localhost:3100`
//   - Networking:
//     Attached to the host interface's (meaning, all ports on the Kong DP are exposed to the host).
//     Additionally, a host record for `host.docker.internal` is created, so that you may hit the
//     host OS from the container without binding the container to the host's networking interface.
//
// See the DockerInput type for more environment variables that can be passed in.
//
// Environment variables passed to the DP will be set in the following order of precedence:
// 1. Environment variables set on DockerInput.EnvVars.
// 2. `KONG_*` environment variables provided on process (only when DockerInput.EnableAutoDPEnv is true).
// 3. Default value based on host OS
// 4. Default values defined in defaultEnvVars.
//
// For both environment variables provided on the process & DockerInput.EnvVars, such environment variables
// will be dropped if they contain an empty value.
//
// In the event a Go test is being executed with verbose logging, the `KONG_LOG_LEVEL` environment variable
// will be set to `DEBUG`.
func addDefaults(input *DockerInput) (*DockerInput, error) {
	res := input

	if res.Internal.Hosts == nil {
		res.Internal.Hosts = make(map[string]string)
	}

	// In the event the control plane address is being provided part of the input & not an
	// environment variable, we'll want to assume they want it to route to the host.
	//
	// This is required as some tests specifically set DockerInput.CPAddr & must be done before
	// `env.Parse()` is called (as else an environment variable on the process can override it).
	if res.CPAddr != "" {
		cpHostname, _, err := net.SplitHostPort(res.CPAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid DockerInput.CPAddr address %q: %w", res.CPAddr, err)
		}
		res.Internal.Hosts[cpHostname] = dockerHostGWName
	}

	// Set up the environment variables that will be passed to the Kong DP container.
	if err := setupEnv(res); err != nil {
		return nil, err
	}

	// Set up the control plane address, ports, etc. that will be configured for the Kong DP container.
	if err := setupNetworking(res); err != nil {
		return nil, err
	}

	// Allow the Kong DP access to the control plane certs & Nginx admin config directories.
	res.Internal.Volumes = map[string]string{
		"certs": "/certs",
		"conf":  "/conf",
	}

	return res, nil
}

func GetKongConfForShared() DockerInput {
	return DockerInput{
		EnvVars: map[string]string{
			"KONG_CLUSTER_MTLS": "shared",
		},
		ClientKey:  certs.DefaultSharedKey,
		ClientCert: certs.DefaultSharedCert,
		CACert:     certs.DefaultSharedCert,
	}
}

// GetKongConfForSharedWithSecrets is the same as GetKongConfForShared
// but with the addition of KONG_MY_SECRET_* environment variabled
// to be used for secrets management testing, as well as the additional
// exposure of a port for ssl requests.
func GetKongConfForSharedWithSecrets() DockerInput {
	res := GetKongConfForShared()
	res.EnvVars["KONG_MY_SECRET_CERT"] = secretCert
	res.EnvVars["KONG_MY_SECRET_KEY"] = secretKey
	res.EnvVars["KONG_PROXY_LISTEN"] = "0.0.0.0:8000, 0.0.0.0:8443 ssl"
	return res
}

func GetKongConf() DockerInput {
	res := GetKongConfForShared()
	res.EnvVars["KONG_CLUSTER_SERVER_NAME"] = "cp.example.com"
	res.EnvVars["KONG_CLUSTER_MTLS"] = "pki"
	res.CACert = certs.CPCACert
	res.ClientKey = certs.DPTree1Key
	res.ClientCert = certs.DPTree1Cert
	return res
}

func removeDockerContainer(containerName string) error {
	c := exec.Command("docker", "rm", "-f", containerName)
	var err error
	if err = c.Start(); err == nil {
		if err = c.Wait(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to remove Docker container: %w", err)
}

func streamLogs(wg *sync.WaitGroup, r io.ReadCloser, isStdErr bool) {
	defer wg.Done()

	prefix, f := "stdout", os.Stdout
	if isStdErr {
		prefix, f = "stderr", os.Stderr
	}
	prefix = "dp-" + prefix + ": "

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		_, _ = f.WriteString(prefix + sc.Text() + "\n")
	}
}

func setupEnv(input *DockerInput) error {
	if input.EnvVars == nil {
		input.EnvVars = make(map[string]string)
	}

	// Fetch & validate the environment variables from the Koko process.
	input.Internal.HostEnvVars = getHostEnv()
	if err := validateHostEnv(input.Internal.HostEnvVars); err != nil {
		return err
	}

	if testing.Verbose() {
		input.EnvVars["KONG_LOG_LEVEL"] = "debug"
	}

	// Ignore the provided `KONG_*` from the host environment when required.
	if val, _ := strconv.ParseBool(input.Internal.HostEnvVars[envEnableAutoDPEnv]); !val {
		input.Internal.HostEnvVars = lo.PickBy(input.Internal.HostEnvVars, func(key, _ string) bool {
			return !isKongEnvVar(key)
		})
	}

	// Parse the given environment variables & set them for use on the Kong DP container.
	newEnv := lo.Assign(defaultEnvVars, input.Internal.HostEnvVars, input.EnvVars)
	input.EnvVars = lo.PickBy(newEnv, filterEnvVars(true, isKongEnvVar))
	return env.Parse(input, env.Options{Environment: newEnv})
}

func setupNetworking(input *DockerInput) error {
	// Depending on the host OS, we'll set different defaults for the CP listen address.
	//
	// TODO(tjasko): By default, we should be setting the CP & DP ports (that are published to the host) to be free,
	//  random ports. However due to all the hard-coding going on, this is easier said than done. Once those changes
	//  are in, we'll be able to then update this tooling to return the info needed to access the DP.
	var defaultCPAddr, defaultNetwork string
	var ports map[int]int
	switch runtime.GOOS {
	case "linux":
		// As we're already running on Linux, just give the container access to the host's networking.
		defaultCPAddr, defaultNetwork = "localhost:3100", dockerHostNetwork

		// Let Linux users be able to talk to the host in case they don't want to bind to the host's
		// networking interface. This hostname is automatically added by Docker for non-Linux hosts.
		input.Internal.Hosts[dockerHostGWHostname] = dockerHostGWName
	case "darwin":
		// By default, let the DP reach out to the CP running on the host system.
		defaultCPAddr, ports = net.JoinHostPort(dockerHostGWHostname, "3100"), defaultDockerPorts
	default:
		// Yes, we can likely support Windows w/ Docker+WSL2 just fine, however it's currently untested.
		return fmt.Errorf("unsupported host OS: %s", runtime.GOOS)
	}

	// The control plane address that's passed to the DP will be set in the following order of precedence:
	// 1. `KONG_CLUSTER_CONTROL_PLANE` environment variable provided on process.
	// 2. Control plane address set on DockerInput.
	// 3. Default value based on host OS.
	if input.CPAddr == "" {
		input.CPAddr = defaultCPAddr
	}

	// There's no easy way with the `env` library we're using to go back to a slice of
	// environment variables. As this is the only environment variable that we override
	// part of this logic, we don't need to do anything too sophisticated right now.
	input.EnvVars["KONG_CLUSTER_CONTROL_PLANE"] = input.CPAddr

	// Any environment variables that were
	// When no explicit Docker network has been set via the provided
	// environment variable, set the default based on the host OS.
	if input.Network == "" {
		input.Network = defaultNetwork
	}

	// When no explicit Docker ports have been set, provide the default based on the host OS.
	// On Linux, providing ports when the DP container is being bounded to the host is unnecessary.
	if input.Network != dockerHostNetwork {
		input.Internal.Ports = ports
	}

	return nil
}

func validateHostEnv(e map[string]string) error {
	if e["DOCKER_HOST"] != "" {
		return errors.New("using DOCKER_HOST is unsupported until DP hostnames & ports are customizable")
	}

	return nil
}

func getHostEnv() map[string]string {
	return lo.Associate(os.Environ(), func(val string) (string, string) {
		parts := strings.SplitN(val, "=", 2) //nolint:gomnd
		return strings.ToUpper(parts[0]), parts[1]
	})
}

func isKongEnvVar(name string) bool {
	return strings.HasPrefix(name, "KONG_")
}

func filterEnvVars(filterEmpty bool, keyFilter func(string) bool) func(key, val string) bool {
	return func(key, val string) bool {
		if filterEmpty && val == "" {
			return false
		}
		if !keyFilter(key) {
			return false
		}
		return true
	}
}
