package containers

import (
	"ecr-mirror-sync/pkg/options"

	"github.com/containers/common/pkg/retry"
)

type Copy struct {
	additionalTags   []string // For docker-archive: destinations, in addition to the name:tag specified as destination, also add these
	destImage        *options.ImageDestOptions
	global           *options.GlobalOptions
	quiet            bool // Suppress output information when copying images
	removeSignatures bool // Do not copy signatures from the source image
	retryOpts        *retry.RetryOptions
	srcImage         options.ImageOptions
}

type Inspect struct {
	global        *options.GlobalOptions
	image         options.ImageOptions
	retryOpts     *retry.RetryOptions
	doNotListTags bool // Do not list all tags available in the same repository
}
