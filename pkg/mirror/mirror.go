package mirror

import (
	"ecr-mirror-sync/pkg/containers"
	"ecr-mirror-sync/pkg/options"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TwiN/go-color"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/containers/image/manifest"
	"github.com/gammazero/workerpool"
	"github.com/jedib0t/go-pretty/table"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func New(opts *options.MirrorOptions) *MirrorProvider {

	if opts.Debug {
		log.SetLevel(logrus.DebugLevel)
	}

	token, err := options.GetECRAuthToken(aws.String(opts.Region))
	if err != nil {
		return nil
	}

	authorizationToken, err := base64.StdEncoding.DecodeString(*token)
	if err != nil {
		log.Error("could not get ECR authorization token from AWS : %v", err)
	}

	return &MirrorProvider{
		AWSClientSession: options.GetDefaultAwsClient(aws.String(opts.Region)),
		DefaultECRRegion: aws.String(opts.Region),
		ECRAuthToken:     authorizationToken,
		ECRTypeFilter:    []*string{aws.String("ecr:repository")},
		Options:          opts,
		UpstreamImageKey: aws.String(opts.UpstreamImageKey),
		UpstreamTagsKey:  aws.String(opts.UpstreamTagsKey),
	}
}

func (p *MirrorProvider) List() []MirrorRepository {
	return p.getECRTaggedRepos()
}

func (p *MirrorProvider) Sync() {
	log.Info("Attempting to sync public images to private ecr repositories...")
	p.copy(p.getECRTaggedRepos())
}

func (p *MirrorProvider) Copy(upstreamImageTag, ecrRespository string) {

	if upstreamImageTag != "" && ecrRespository != "" {
		upstream := strings.Split(upstreamImageTag, ":")

		mirrorRepos := []MirrorRepository{{
			UpstreamImage:  upstream[0],
			UpstreamTag:    upstream[1],
			ECRRespository: ecrRespository,
		}}
		log.Info("Attempting to copy public image to private ecr repository...")
		p.copy(mirrorRepos)
	} else {
		log.Error("upstream image tag or ecr repository missing")
	}

}

func (p *MirrorProvider) getImageDigest(mirror MirrorRepository, tag string) (string, error) {
	log.Debugf("Get Image Digest for %s:%s...", mirror.ECRRespository, mirror.UpstreamTag)

	// We can not predict the output of calls to inspect since the image results coould be v1 or v2 image spec
	// We therefore need to store the results in an interface which we unmarshal into json
	var manifests interface{}

	var digest string
	var err error
	var raw []byte

	mirrorImageFlag := fmt.Sprintf("%s://%s:%s", options.RemoteTransport, mirror.UpstreamImage, tag)

	manifestOptions := &options.ManifestOptions{
		DoNotListTags: true,
		Global:        p.Options.Global,
		Image:         *p.Options.SrcImage,
		Raw:           true,
		RetryOpts:     p.Options.RetryOpts,
	}

	ms := containers.NewManifestProvider(*manifestOptions)
	raw, err = ms.Manifest([]string{mirrorImageFlag})

	if err != nil {
		return "", err
	}

	json.Unmarshal(raw, &manifests)

	// v1 image spec, use manifest to grab Digest
	if _, ok := manifests.(map[string]interface{})["config"]; ok {
		digestRaw, err := manifest.Digest(raw)
		if err != nil {
			return "", err
		}
		digest = string(digestRaw)

	} else {
		if manifest, ok := manifests.(map[string]interface{})["manifests"]; ok {
			for _, m := range manifest.([]interface{}) {
				if m.(map[string]interface{})["platform"].(map[string]interface{})["architecture"] == "amd64" &&
					m.(map[string]interface{})["platform"].(map[string]interface{})["os"] == "linux" {
					digest = m.(map[string]interface{})["digest"].(string)
				}

			}

		}
	}

	return digest, err
}

func (p *MirrorProvider) copy(mirrorRepos []MirrorRepository) {

	var (
		t    table.Writer
		pool int
		err  error
	)

	ecrSession := ecr.New(p.AWSClientSession)
	c := containers.NewCopyProvider(p.Options)

	if p.Options.RenderTable {
		t = table.NewWriter()
	}

	totalProcessed := 0
	totalSucceeded := 0
	totalfailed := 0

	p.Options.DestImage.CredsOption = string(p.ECRAuthToken)

	p.Options.Global.CommandTimeout = 20 * time.Minute // Hard coded by default

	if p.Options.WorkerPoolSize != "" {
		pool, err = strconv.Atoi(p.Options.WorkerPoolSize)
		if err != nil {
			log.Fatalf("%s", err)
		}
	} else {
		pool = len(mirrorRepos)
	}

	wp := workerpool.New(pool)

	log.Infof("Batch size for syncing images: %v", pool)

	for _, mirror := range mirrorRepos {

		mirror := mirror

		wp.Submit(func() {

			var (
				ecrRespositoryFlag string
				mirrorImageFlag    string
				ecrRepo            string
			)

			fromToFields := log.Fields{
				"from": fmt.Sprintf("%s:%s", mirror.UpstreamImage, mirror.UpstreamTag),
				"to":   fmt.Sprintf("%s:%s", mirror.ECRRespository, mirror.UpstreamTag),
			}

			if p.Options.MirrorRepoPrefix != "" {
				ecrRepo = fmt.Sprintf("%s/%s", p.Options.MirrorRepoPrefix, mirror.UpstreamImage)
			} else {
				ecrRepo = mirror.UpstreamImage
			}

			imageTagFilter := &ecr.ImageIdentifier{
				ImageTag: &mirror.UpstreamTag,
			}
			input := &ecr.DescribeImagesInput{
				RepositoryName: &ecrRepo,
				ImageIds:       []*ecr.ImageIdentifier{imageTagFilter},
			}

			// check if image tag was provided for in mirror.ECRRespository
			if !strings.Contains(mirror.ECRRespository, ":") {
				ecrRespositoryFlag = fmt.Sprintf("%s://%s:%s", options.RemoteTransport, mirror.ECRRespository, mirror.UpstreamTag)
			} else {
				ecrRespositoryFlag = fmt.Sprintf("%s://%s", options.RemoteTransport, mirror.ECRRespository)
			}

			mirrorImageFlag = fmt.Sprintf("%s://%s:%s", options.RemoteTransport, mirror.UpstreamImage, mirror.UpstreamTag)

			image, err := ecrSession.DescribeImages(input)

			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {

					case ecr.ErrCodeInvalidParameterException:
						log.WithFields(fromToFields).Error("%s:%s: %s", mirror.ECRRespository, mirror.UpstreamTag, aerr.Message())
						log.WithFields(fromToFields).Error("%s:%s: Will not mirror image", mirror.ECRRespository, mirror.UpstreamTag)
						mirror.Status = color.Ize(color.Red, fmt.Sprintf("Will not mirror image: %s", err.Error()))
					case ecr.ErrCodeRepositoryNotFoundException:
						log.WithFields(fromToFields).Error("%s:%s: %s", mirror.ECRRespository, mirror.UpstreamTag, aerr.Message())
						mirror.Status = color.Ize(color.Red, fmt.Sprintf("ecr repo does not exist: %s", err.Error()))

					case ecr.ErrCodeImageNotFoundException:
						log.WithFields(fromToFields).Infof("%s:%s: %s", mirror.ECRRespository, mirror.UpstreamTag, aerr.Message())
						log.WithFields(fromToFields).Infof("%s:%s: Will attempt to mirror public repository", mirror.ECRRespository, mirror.UpstreamTag)

						if p.Options.DryRun {
							log.WithFields(fromToFields).Infof("Would have copied image %s", mirror.UpstreamTag)
						} else {
							err = c.Copy([]string{mirrorImageFlag, ecrRespositoryFlag}, os.Stdout)

							if err != nil {
								mirror.Status = color.Ize(color.Red, fmt.Sprintf("failed to mirror: %s", err.Error()))
							} else {
								mirror.Status = color.Ize(color.Green, "success")

							}
							totalProcessed++

						}
					}

				}
			}

			if mirror.Status == "" {

				var (
					digest string
					err    error
				)

				if p.Options.DryRun {
					log.WithFields(fromToFields).Infof("Would get digest for public image tag %s", mirror.UpstreamTag)
					mirror.Status = color.Ize(color.Yellow, "Dry Run")
				} else {
					log.WithFields(fromToFields).Infof("Checking Digest for Upstream Image %s with Tag %s...", mirror.UpstreamImage, mirror.UpstreamTag)
					digest, err = p.getImageDigest(mirror, mirror.UpstreamTag)
				}
				if err != nil {
					log.WithFields(fromToFields).Errorf("%s:%s: %s", mirror.ECRRespository, mirror.UpstreamTag, err)
					mirror.Status = color.Ize(color.Red, fmt.Sprintf("failed to mirror: %s", err.Error()))

				} else if digest != "" {
					if *image.ImageDetails[0].ImageDigest != digest {

						log.WithFields(fromToFields).Infof("We have a diff in digest for %s:%s %s vs %s. Attempting to copy...", mirror.UpstreamImage, mirror.UpstreamTag, *image.ImageDetails[0].ImageDigest, digest)

						if p.Options.DryRun {
							log.WithFields(fromToFields).WithFields(fromToFields).Infof("Would have copied image %s", mirror.UpstreamTag)
						} else {
							err = c.Copy([]string{mirrorImageFlag, ecrRespositoryFlag}, os.Stdout)

							if err != nil {
								mirror.Status = color.Ize(color.Red, fmt.Sprintf("failed to mirror: %s", err.Error()))
							} else {
								mirror.Status = color.Ize(color.Green, "success")

							}
							totalProcessed++
						}
					} else {
						log.WithFields(fromToFields).Info("ecr image digest matches upstream image")
						mirror.Status = color.Ize(color.White, "skipping, image exists already")
					}

				} else if image != nil && !p.Options.DryRun {
					mirror.Status = color.Ize(color.Yellow, "Could not retrieve image digest from pulbic upstream. However an Image exist in ECR, skipped")
				}

			}
			if strings.Contains(mirror.Status, "failed") {
				totalfailed++
				totalProcessed++
			}

			if mirror.Status == "success" {
				totalSucceeded++
			}
			if p.Options.RenderTable {
				t.AppendRows([]table.Row{
					{mirror.UpstreamImage, mirror.ECRRespository, mirror.UpstreamTag, mirror.Status},
				})
			}
		})
	}
	wp.StopWait()

	if p.Options.RenderTable {

		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Source Image", "Destination", "Tag", "Status"})
		t.AppendFooter(table.Row{"Total Images Processed", totalProcessed})
		t.AppendFooter(table.Row{"Total Succeeded", totalSucceeded})
		t.AppendFooter(table.Row{"Total Failed", totalfailed})
		t.AppendFooter(table.Row{"Total", t.Length()})
		t.Render()
	} else {
		log.Infof("Total Images: %d", len(mirrorRepos))
		log.Infof("Total Images Processed: %d", totalProcessed)
		log.Infof("Total Mirrors Succeeded: %d", totalSucceeded)
		log.Infof("Total Mirrors Failed: %d", totalfailed)

	}

}

func (p *MirrorProvider) getECRTaggedRepos() (mirrorRepos []MirrorRepository) {

	var t table.Writer

	if p.Options.RenderTable {
		t = table.NewWriter()
	}
	resource := resourcegroupstaggingapi.New(p.AWSClientSession)

	res, err := resource.GetResources(options.ECRRepofilters())
	if err != nil {
		log.Fatalf("failed to get resource(s): %v", err)
	}

	var mirrorRepo MirrorRepository

	for _, repo := range res.ResourceTagMappingList {

		var upstreamTags []string
		parsedARN, err := arn.Parse(*repo.ResourceARN)
		if err != nil {
			log.Error(err.Error())
		}

		re := regexp.MustCompile("^repository/(.*?)$")
		repoName := re.FindStringSubmatch(parsedARN.Resource)

		mirrorRepo.ECRRespository = fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", parsedARN.AccountID, parsedARN.Region, repoName[1])

		for _, repoTag := range repo.Tags {

			switch *repoTag.Key {
			case *options.UpstreamImage:
				mirrorRepo.UpstreamImage = *repoTag.Value
			case *options.UpstreamTags:
				upstreamTags = strings.Split(strings.Replace(*repoTag.Value, "+", "*", -1), "/")
			}
		}

		for _, tag := range upstreamTags {
			mirrorRepo.UpstreamTag = tag
			mirrorRepos = append(mirrorRepos, mirrorRepo)

			if p.Options.RenderTable {
				t.AppendRows([]table.Row{
					{mirrorRepo.UpstreamImage, mirrorRepo.ECRRespository, tag},
				})
			}
		}

	}

	if p.Options.RenderTable {
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Source Image", "Destination", "Tag", "Status"})
		t.AppendFooter(table.Row{"Total Images to Mirror", len(mirrorRepos)})
		t.Render()
	}

	log.Infof("Total Images to Mirror: %d", len(mirrorRepos))
	return mirrorRepos

}
