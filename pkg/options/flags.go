package options

import (
	"os"

	"github.com/containers/common/pkg/retry"
	"github.com/spf13/pflag"
)

// DockerImageFlags prepares a collection of docker-transport specific CLI flags
// writing into imageOptions, and the managed imageOptions structure.
func DockerImageFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageOptions) {
	flags := ImageOptions{
		DockerImageOptions: DockerImageOptions{
			Global: global,
		},
	}

	fs := pflag.FlagSet{}
	if flagPrefix != "" {
		// the non-prefixed flag is handled by a shared flag.
		fs.StringVar(&flags.Global.AuthFilePath, flagPrefix+"authfile", os.Getenv("REGISTRY_AUTH_FILE"), "path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json")
	}
	fs.BoolVar(&flags.NoCreds, flagPrefix+"no-creds", false, "Access the registry anonymously")
	fs.StringVar(&flags.CredsOption, flagPrefix+"creds", "", "Use `USERNAME[:PASSWORD]` for accessing the registry")
	fs.StringVar(&flags.DockerCertPath, flagPrefix+"cert-dir", "", "use certificates at `PATH` (*.crt, *.cert, *.key) to connect to the registry or daemon")
	fs.StringVar(&flags.Password, flagPrefix+"password", "", "Password for accessing the registry")
	fs.StringVar(&flags.RegistryToken, flagPrefix+"registry-token", "", "Provide a Bearer token for accessing the registry")
	fs.StringVar(&flags.CredType, flagPrefix+"cred-type", "docker", "Registry type. Defaults to docker")
	fs.StringVar(&flags.UserName, flagPrefix+"username", "", "Username for accessing the registry")
	return fs, &flags
}

// ImageFlags prepares a collection of CLI flags writing into imageOptions, and the managed imageOptions structure.
func ImageFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageOptions) {
	dockerFlags, opts := DockerImageFlags(global, flagPrefix, credsOptionAlias)

	fs := pflag.FlagSet{}
	fs.StringVar(&opts.SharedBlobDir, flagPrefix+"shared-blob-dir", "", "`DIRECTORY` to use to share blobs across OCI repositories")
	fs.StringVar(&opts.DockerDaemonHost, flagPrefix+"daemon-host", "", "use docker daemon host at `HOST` (docker-daemon: only)")
	fs.AddFlagSet(&dockerFlags)
	return fs, opts
}

func RetryFlags() (pflag.FlagSet, *retry.RetryOptions) {
	opts := retry.RetryOptions{}
	fs := pflag.FlagSet{}
	fs.IntVar(&opts.MaxRetry, "retry-times", 0, "the number of times to possibly retry")
	return fs, &opts
}
func MirrorFlags(global *GlobalOptions, srcOpts *ImageOptions, destOpts *ImageDestOptions, retryOpts *retry.RetryOptions) (pflag.FlagSet, *MirrorOptions) {
	flags := MirrorOptions{
		Global:           global,
		RetryOpts:        retryOpts,
		SrcImage:         srcOpts,
		DestImage:        destOpts,
		RemoveSignatures: false, // Do not copy signatures from the source image
		Quiet:            false, // Suppress output information when copying images
		AdditionalTags:   []string{},
	}
	fs := pflag.FlagSet{}
	fs.BoolVar(&flags.Debug, "debug", false, "enable debug output")
	fs.BoolVar(&flags.RenderTable, "render-table", false, "Render tables")
	fs.BoolVar(&flags.DryRun, "dry-run", false, "Run without actually copying data")

	return fs, &flags
}

// ImageDestFlags prepares a collection of CLI flags writing into imageDestOptions, and the managed imageDestOptions structure.
func ImageDestFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageDestOptions) {
	_, genericOptions := ImageFlags(global, flagPrefix, credsOptionAlias)
	opts := ImageDestOptions{ImageOptions: genericOptions}
	fs := pflag.FlagSet{}
	fs.BoolVar(&opts.precomputeDigests, flagPrefix+"precompute-digests", true, "Precompute digests to prevent uploading layers already on the registry using the 'docker' transport.")
	return fs, &opts
}
