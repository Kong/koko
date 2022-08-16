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
