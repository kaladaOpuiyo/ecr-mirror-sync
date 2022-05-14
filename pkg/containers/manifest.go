package containers

import (
	"ecr-mirror-sync/pkg/options"
	"fmt"
	"regexp"
	"strings"

	"github.com/containers/common/pkg/retry"
	"github.com/containers/image/v5/types"
	"github.com/pkg/errors"
)

func NewManifestProvider(options options.ManifestOptions) *Manifest {
	return &Manifest{

		global:        options.Global,
		retryOpts:     options.RetryOpts,
		image:         options.Image,
		doNotListTags: options.DoNotListTags,
	}
}

func (opts *Manifest) Manifest(args []string) (rawManifest []byte, err error) {

	// When Syncing a combinationn of images from multiple repositories, we favor dockerhub when using command line flags to pass credentials
	// we expect that the other repositories are accessible anonymously
	anonymous, _ := regexp.MatchString(`([^\s]+)\.([^\s]+)\/([^\s]+)`, strings.TrimPrefix(args[0], "docker://"))

	if anonymous && !strings.Contains(args[0], "docker.io") && opts.image.DockerImageOptions.Transport == "docker" {

		opts.image.DockerImageOptions.CredsOption = ""
	}

	var (
		src types.ImageSource
	)

	ctx, cancel := opts.global.TimeoutContext()
	defer cancel()

	if len(args) != 1 {
		return rawManifest, errors.New("Exactly one argument expected")
	}

	imageName := args[0]

	if err := retry.RetryIfNecessary(ctx, func() error {
		src, err = options.ParseImageSource(ctx, &opts.image, imageName)
		return err
	}, opts.retryOpts); err != nil {
		return rawManifest, errors.Wrapf(err, "Error parsing image name %q", imageName)
	}

	defer func() {
		if err := src.Close(); err != nil {
			err = errors.Wrapf(err, fmt.Sprintf("(could not close image: %v) ", err))
		}
	}()

	if err := retry.RetryIfNecessary(ctx, func() error {
		rawManifest, _, err = src.GetManifest(ctx, nil)
		return err
	}, opts.retryOpts); err != nil {
		return rawManifest, errors.Wrapf(err, "Error retrieving manifest for image")
	}

	return rawManifest, nil
}
