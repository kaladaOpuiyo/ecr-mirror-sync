package options

import (
	"os"

	"github.com/containers/common/pkg/retry"
	"github.com/spf13/pflag"
)

func DockerImageFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageOptions) {
	flags := ImageOptions{
		DockerImageOptions: DockerImageOptions{
			Global: global,
		},
	}

	fs := pflag.FlagSet{}
	fs.StringVar(&flags.Global.AuthFilePath, flagPrefix+"authfile", os.Getenv("REGISTRY_AUTH_FILE"), "path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json")
	fs.BoolVar(&flags.NoCreds, flagPrefix+"no-creds", false, "Access the registry anonymously")
	fs.StringVar(&flags.CredsOption, flagPrefix+"creds", "", "Use `USERNAME[:PASSWORD]` for accessing the registry")
	fs.StringVar(&flags.DockerCertPath, flagPrefix+"cert-dir", "", "use certificates at `PATH` (*.crt, *.cert, *.key) to connect to the registry or daemon")
	fs.StringVar(&flags.Password, flagPrefix+"password", "", "Password for accessing the registry")
	fs.StringVar(&flags.RegistryToken, flagPrefix+"registry-token", "", "Provide a Bearer token for accessing the registry")
	fs.StringVar(&flags.UserName, flagPrefix+"username", "", "Username for accessing the registry")
	return fs, &flags
}

func ImageFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageOptions) {
	dockerFlags, opts := DockerImageFlags(global, flagPrefix, credsOptionAlias)

	fs := pflag.FlagSet{}
	fs.AddFlagSet(&dockerFlags)
	return fs, opts
}

func GlobalFlags() (pflag.FlagSet, *GlobalOptions) {
	opts := GlobalOptions{}
	fs := pflag.FlagSet{}

	fs.BoolVar(&opts.InsecurePolicy, "insecure-policy", false, "run the tool without any policy check")
	fs.StringVar(&opts.OverrideArch, "override-arch", "amd64", "use `ARCH` instead of the architecture of the machine for choosing images")
	fs.StringVar(&opts.OverrideOS, "override-os", "linux", "use `OS` instead of the running OS for choosing images")
	fs.StringVar(&opts.OverrideVariant, "override-variant", "", "use `VARIANT` instead of the running architecture variant for choosing images")
	fs.StringVar(&opts.PolicyPath, "policy", "", "Path to a trust policy file")

	return fs, &opts
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
	fs.BoolVar(&flags.DryRun, "dry-run", false, "Run without actually copying data")
	fs.BoolVar(&flags.RenderTable, "render-table", false, "Render tables")
	fs.StringVar(&flags.UpstreamImageKey, "image-key", "upstream-image", "aws resource tag for upstream image")
	fs.StringVar(&flags.UpstreamTagsKey, "tag-key", "upstream-tags", "aws resource tag for upstream tags")

	return fs, &flags
}

// ImageDestFlags prepares a collection of CLI flags writing into ImageDestOptions, and the managed ImageDestOptions structure.
func ImageDestFlags(global *GlobalOptions, flagPrefix, credsOptionAlias string) (pflag.FlagSet, *ImageDestOptions) {
	_, genericOptions := ImageFlags(global, flagPrefix, credsOptionAlias)
	opts := ImageDestOptions{ImageOptions: genericOptions}
	fs := pflag.FlagSet{}
	fs.BoolVar(&opts.precomputeDigests, flagPrefix+"precompute-digests", true, "Precompute digests to prevent uploading layers already on the registry using the 'docker' transport.")
	return fs, &opts
}
