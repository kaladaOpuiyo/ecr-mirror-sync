package mirror

import (
	"ecr-mirror-sync/pkg/options"

	"github.com/aws/aws-sdk-go/aws/session"
)

type MirrorRepository struct {
	ECRRespository string
	Status         string
	SyncImage      bool
	UpstreamImage  string
	UpstreamTag    string
}
type MirrorProvider struct {
	AWSClientSession *session.Session
	DefaultECRRegion *string
	ECRAuthToken     []byte
	ECRTypeFilter    []*string
	Options          *options.MirrorOptions
	UpstreamImageKey *string
	UpstreamTagsKey  *string
}
