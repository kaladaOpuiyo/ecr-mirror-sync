package containers

import (
	"ecr-mirror-sync/pkg/options"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/containers/common/pkg/retry"
	"github.com/containers/image/manifest"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/transports/alltransports"
	log "github.com/sirupsen/logrus"
)

func NewCopyProvider(options *options.MirrorOptions) *Copy {
	return &Copy{
		additionalTags:   []string{},
		destImage:        options.DestImage,
		global:           options.Global,
		quiet:            options.Quiet,
		removeSignatures: options.RemoveSignatures,
		retryOpts:        options.RetryOpts,
		srcImage:         *options.SrcImage,
	}
}

func (opts *Copy) Copy(args []string, stdout io.Writer) (retErr error) {

	// When Syncing a combinationn of images from multiple repositories, we favor dockerhub when passing credentials,
	// we expect that the other repositories are accessible anonymously.
	anonymous, _ := regexp.MatchString(`([^\s]+)\.([^\s]+)\/([^\s]+)`, strings.TrimPrefix(args[0], "docker://"))

	if anonymous && !strings.Contains(args[0], "docker.io") &&
		opts.srcImage.DockerImageOptions.Transport == "docker" &&
		opts.srcImage.DockerImageOptions.Global.AuthFilePath == "" {

		opts.srcImage.DockerImageOptions.CredsOption = ""
	}

	if len(args) != 2 {
		log.Error("Exactly two arguments expected")
	}
	imageNames := args

	policyContext, err := opts.global.GetPolicyContext()
	if err != nil {
		log.Error("Error loading trust policy: %v", err)
	}
	defer func() {
		if err := policyContext.Destroy(); err != nil {
			retErr = fmt.Errorf("(error tearing down policy context: %v): %w", err, retErr)
		}
	}()

	srcRef, err := alltransports.ParseImageName(imageNames[0])
	if err != nil {
		log.Error("Invalid source name %s: %v", imageNames[0], err)
	}
	destRef, err := alltransports.ParseImageName(imageNames[1])
	if err != nil {
		log.Error("Invalid destination name %s: %v", imageNames[1], err)
	}

	srcCtx, err := opts.srcImage.NewSystemContext()
	if err != nil {
		return err
	}
	destCtx, err := opts.destImage.NewSystemContext()
	if err != nil {
		return err
	}

	ctx, cancel := opts.global.TimeoutContext()
	defer cancel()

	if opts.quiet {
		stdout = nil
	}

	imageListSelection := copy.CopySystemImage

	return retry.RetryIfNecessary(ctx, func() error {
		_, err := copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
			DestinationCtx:        destCtx,
			ForceManifestMIMEType: manifest.DockerV2Schema2MediaType,
			ImageListSelection:    imageListSelection,
			PreserveDigests:       true,
			RemoveSignatures:      false,
			ReportWriter:          stdout,
			SourceCtx:             srcCtx,
		})
		if err != nil {
			return err
		}

		return nil
	}, opts.retryOpts)
}
