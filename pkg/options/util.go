package options

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

var (
	UpstreamImage    = aws.String("upstream-image")
	UpstreamTags     = aws.String("upstream-tags")
	ECRypeFilter     = []*string{aws.String("ecr:repository")}
	DefaultECRRegion = aws.String("us-east-1")
)

func GetDockerAuth(creds string) (*types.DockerAuthConfig, error) {
	username, password, err := ParseCreds(creds)
	if err != nil {
		return nil, err
	}
	return &types.DockerAuthConfig{
		Username: username,
		Password: password,
	}, nil
}

func ParseCreds(creds string) (string, string, error) {
	if creds == "" {
		return "", "", errors.New("credentials can't be empty")
	}
	up := strings.SplitN(creds, ":", 2)
	if len(up) == 1 {
		return up[0], "", nil
	}
	if up[0] == "" {
		return "", "", errors.New("username can't be empty")
	}
	return up[0], up[1], nil
}

// parseImageSource converts image URL-like string to an ImageSource.
// The caller must call .Close() on the returned ImageSource.
func ParseImageSource(ctx context.Context, opts *ImageOptions, name string) (types.ImageSource, error) {
	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return nil, err
	}
	sys, err := opts.NewSystemContext()
	if err != nil {
		return nil, err
	}
	return ref.NewImageSource(ctx, sys)
}

func GetDefaultAwsClient() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: DefaultECRRegion}, SharedConfigState: session.SharedConfigEnable}))
}

func ECRRepofilters() *resourcegroupstaggingapi.GetResourcesInput {

	upstreamImageFilter := &resourcegroupstaggingapi.TagFilter{
		Key: UpstreamImage,
	}
	upstreamTagsFilter := &resourcegroupstaggingapi.TagFilter{
		Key: UpstreamTags,
	}

	return &resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: ECRypeFilter,
		TagFilters: []*resourcegroupstaggingapi.TagFilter{
			upstreamImageFilter, upstreamTagsFilter},
	}
}

func GetECRAuthToken() (*string, error) {
	svc := ecr.New(GetDefaultAwsClient())
	input := &ecr.GetAuthorizationTokenInput{}
	ecrToken, err := svc.GetAuthorizationToken(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				log.Println(ecr.ErrCodeServerException, aerr.Error())
				return nil, err
			case ecr.ErrCodeInvalidParameterException:
				log.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
				return nil, err
			default:
				log.Println(aerr.Error())
				return nil, err
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())

			return nil, err
		}
	}

	return ecrToken.AuthorizationData[0].AuthorizationToken, nil
}
