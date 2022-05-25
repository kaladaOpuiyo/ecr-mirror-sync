package options

import (
	"context"
	"errors"
	"time"

	"github.com/containers/common/pkg/retry"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
)

const (
	Version          = "1.0.0"
	DefaultUserAgent = "ecr-mirror-sync/" + Version
	RemoteTransport  = "docker"
)

// errorShouldDisplayUsage is a subtype of error used by command handlers to indicate that cli.ShowSubcommandHelp should be called.
type ErrorShouldDisplayUsage struct {
	error
}

type MirrorOptions struct {
	AdditionalTags   []string // For docker-archive: destinations, in addition to the name:tag specified as destination, also add these
	Debug            bool     // Enable debug output
	DestImage        *ImageDestOptions
	DryRun           bool // Dry run does not copy
	Global           *GlobalOptions
	MirrorRepoPrefix string
	Quiet            bool   // Suppress output information when copying images
	Region           string // aws region use for ecr repos
	RemoveSignatures bool   // Do not copy signatures from the source image
	RenderTable      bool   //
	RetryOpts        *retry.RetryOptions
	SrcImage         *ImageOptions
	UpstreamImageKey string
	UpstreamTagsKey  string
	WorkerPoolSize   string
}

type ManifestOptions struct {
	DoNotListTags bool // Do not list all tags available in the same repository
	Global        *GlobalOptions
	Image         ImageOptions
	Raw           bool // Output the raw manifest instead of parsing information about the image
	RetryOpts     *retry.RetryOptions
}

type DockerImageOptions struct {
	CredsOption    string // username[:password] for accessing a registry
	Transport      string
	DockerCertPath string         // A directory using Docker-like *.{crt,cert,key} files for connecting to a registry or a daemon
	Global         *GlobalOptions // May be shared across several imageOptions instances.
	NoCreds        bool           // Access the registry anonymously
	Password       string         // password for accessing a registry
	RegistryToken  string         // token to be used directly as a Bearer token when accessing the registry
	UserName       string         // username for accessing a registry
}

type ImageOptions struct {
	DockerDaemonHost string // docker-daemon: host to connect to
	DockerImageOptions
	SharedBlobDir string // A directory to use for OCI blobs, shared across repositories
}

type GlobalOptions struct {
	AuthFilePath    string
	CommandTimeout  time.Duration // Timeout for the command execution
	InsecurePolicy  bool          // Use an "allow everything" signature verification policy
	OverrideArch    string        // Architecture to use for choosing images, instead of the runtime one
	OverrideOS      string        // OS to use for choosing images, instead of the runtime one
	OverrideVariant string        // Architecture variant to use for choosing images, instead of the runtime one
	PolicyPath      string        // Path to a signature verification policy file
}

// GetPolicyContext returns a *signature.PolicyContext based on opts.
func (opts *GlobalOptions) GetPolicyContext() (*signature.PolicyContext, error) {
	var policy *signature.Policy // This could be cached across calls in opts.
	var err error

	if opts.InsecurePolicy {
		policy = &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	} else if opts.PolicyPath == "" {
		policy, err = signature.DefaultPolicy(nil)
	} else {
		policy, err = signature.NewPolicyFromFile(opts.PolicyPath)
	}
	if err != nil {
		return nil, err
	}
	return signature.NewPolicyContext(policy)
}

// NewSystemContext returns a *types.SystemContext corresponding to opts.
// It is guaranteed to return a fresh instance, so it is safe to make additional updates to it.
func (opts *GlobalOptions) newSystemContext() *types.SystemContext {
	ctx := &types.SystemContext{
		ArchitectureChoice:      opts.OverrideArch,
		DockerRegistryUserAgent: DefaultUserAgent,
		OSChoice:                opts.OverrideOS,
		VariantChoice:           opts.OverrideVariant,
	}

	return ctx
}

// TimeoutContext returns a context.Context and a cancellation callback based on opts.
// The caller should usually "defer cancel()" immediately after calling this.
func (opts *GlobalOptions) TimeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if opts.CommandTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.CommandTimeout)
	}
	return ctx, cancel
}

// NewSystemContext returns a *types.SystemContext corresponding to opts.
// It is guaranteed to return a fresh instance, so it is safe to make additional updates to it.
func (opts *ImageOptions) NewSystemContext() (*types.SystemContext, error) {
	// *types.SystemContext instance from globalOptions
	//  imageOptions option overrides the instance if both are present.
	ctx := opts.Global.newSystemContext()
	ctx.DockerCertPath = opts.DockerCertPath
	ctx.OCISharedBlobDirPath = opts.SharedBlobDir
	ctx.AuthFilePath = opts.Global.AuthFilePath
	ctx.DockerDaemonHost = opts.DockerDaemonHost
	ctx.DockerDaemonCertPath = opts.DockerCertPath
	if opts.Global.AuthFilePath != "" {
		ctx.AuthFilePath = opts.Global.AuthFilePath
	}

	if opts.CredsOption != "" && opts.NoCreds {
		return nil, errors.New("creds and no-creds cannot be specified at the same time")
	}
	if opts.UserName != "" && opts.NoCreds {
		return nil, errors.New("username and no-creds cannot be specified at the same time")
	}
	if opts.CredsOption != "" && opts.UserName != "" {
		return nil, errors.New("creds and username cannot be specified at the same time")
	}
	// if any of username or password is present, then both are expected to be present
	if opts.UserName != opts.Password {
		if opts.UserName != "" {
			return nil, errors.New("password must be specified when username is specified")
		}
		return nil, errors.New("username must be specified when password is specified")
	}

	if opts.CredsOption != "" {
		var err error
		ctx.DockerAuthConfig, err = GetDockerAuth(opts.CredsOption)
		if err != nil {
			return nil, err
		}
	} else if opts.UserName != "" {
		ctx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: opts.UserName,
			Password: opts.Password,
		}
	}
	if opts.RegistryToken != "" {
		ctx.DockerBearerRegistryToken = opts.RegistryToken
	}
	if opts.NoCreds {
		ctx.DockerAuthConfig = &types.DockerAuthConfig{}
	}

	return ctx, nil
}

// ImageDestOptions is a superset of imageOptions specialized for image destinations.
type ImageDestOptions struct {
	*ImageOptions
	precomputeDigests bool // Precompute digests to dedup layers when saving to the docker: transport
}
