package kong

const (
	// On non-Linux systems, this is Docker's hostname that allows routing to the host gateway.
	// Read more: https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
	dockerHostGWHostname = "host.docker.internal"

	// Used with `--add-host`, which is Docker's special string to tell it
	// to re-write it to the host's gateway IP. This is not a hostname.
	dockerHostGWName = "host-gateway"

	// The `host` networking interface, which binds the container to the host's networking interface on Linux only.
	// Read more: https://docs.docker.com/network/host
	dockerHostNetwork = "host"
)

// Environment variable backing DockerInput.EnableAutoDPEnv.
const envEnableAutoDPEnv = "KOKO_TEST_ENABLE_AUTO_DP_ENV"

// See DockerInput.EnvVars for usage info.
var defaultEnvVars = map[string]string{
	"KONG_CLUSTER_CA_CERT":             fileClusterCACert,
	"KONG_CLUSTER_CERT":                fileClusterCert,
	"KONG_CLUSTER_CERT_KEY":            fileClusterKey,
	"KONG_LUA_SSL_TRUSTED_CERTIFICATE": fileClusterCACert,
	"KONG_NGINX_HTTP_INCLUDE":          fileAdminConf,

	"KONG_ANONYMOUS_REPORTS":      "off",
	"KONG_DATABASE":               "off",
	"KONG_NGINX_WORKER_PROCESSES": "1",
	"KONG_ROLE":                   "data_plane",
	"KONG_VITALS":                 "off",
}

// See dockerInternalInput.Ports for usage info.
//
//nolint:gomnd
var defaultDockerPorts = map[int]int{
	8000: 8000, // Kong DP HTTP proxy port (`KONG_PROXY_LISTEN`).
	8443: 8443, // Kong DP HTTPS proxy port (`KONG_PROXY_LISTEN`).
	8001: 8001, // Kong DP HTTP admin API port (`KONG_ADMIN_LISTEN`).
}

// DockerInput defines the required values in order to spin up a DP via Docker.
type DockerInput struct {
	// Internal values that are required for use in Go template generation, but must not be set
	// by other callers. It is exported to ensure the underlying Go template works as intended.
	Internal dockerInternalInput

	// Environment variables to be passed to the DP (`-e`/`--env`).
	EnvVars map[string]string

	// The name of the Docker container created for the DP.
	ContainerName string `env:"KOKO_TEST_KONG_DP_CONTAINER_NAME" envDefault:"koko-dp"`

	// The Docker image passed to `docker run ... [IMAGE]`.
	Image string `env:"KOKO_TEST_KONG_DP_IMAGE,notEmpty"`

	// The existing Docker network to attach the DP to (`--network`).
	Network string `env:"KOKO_TEST_KONG_DP_NETWORK"`

	// When set to true, all `KONG_*` environment variables are passed to the DP
	// when creating the container, and these override any defaulted values.
	//
	// NOTE: Here be dragons enabling this, as Koko may be setting `KONG_*` environment
	// variables during integration tests, which are used both by Koko itself along with
	// the DP. An example of this is the `KONG_LICENSE_DATA` environment variable.
	//
	// This setting is mostly meant for local development only, and not for use in CI.
	EnableAutoDPEnv bool `env:"KOKO_TEST_ENABLE_AUTO_DP_ENV"`

	// The address to the control plane.
	// When not provided, this will automatically be set based on the host OS.
	CPAddr string `env:"KONG_CLUSTER_CONTROL_PLANE"`

	// The DP cluster certificate contents.
	ClientCert []byte `env:"KONG_CLUSTER_CERT_RAW"`

	// The DP cluster certificate key contents.
	ClientKey []byte `env:"KONG_CLUSTER_CERT_KEY_RAW"`

	// The DP cluster CA certificate key contents.
	CACert []byte `env:"KONG_CLUSTER_CA_CERT_RAW"`
}

type dockerInternalInput struct {
	// Sources in environment variables from the Koko process.
	HostEnvVars map[string]string

	// Volumes to bind mount (`-v`/`--volume`).
	// (host path)->(container path)
	//
	// Host paths are relative to the temp directory that is created on the host system.
	Volumes map[string]string

	// The host-to-IP mappings that will be created on the container's network interface (`--add-host`).
	// (hostname)->(IP address)
	Hosts map[string]string

	// The ports from the container to publish to the host (`-p`/`--publish`).
	// (Host port)->(Container port)
	Ports map[int]int
}
