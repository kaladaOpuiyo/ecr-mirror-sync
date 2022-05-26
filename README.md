# ECR-MIRROR-SYNC

## **Overview**

The `ecr-mirror-sync` extracts a list of ecr repositories with resource identifier tags indicating the repository is to be mirrored from an external/public image source. (by default, these resource tags are `upstream-image` and `upstream-tags`). It then performs a check to determine if an image already exists in the repository, followed by another to determine if the digest found matches the currently available public image digest. It performs a sync if any of the previous checks are not true. 

By default, requests made to public registries are anonymous; you may, however, pass credentials to authenticate ( an example of this done using the flag `--src-creds`). This tool can run in a Kubernetes cluster as a Cronjob, scheduled for nightly repository syncing (or whatever frequency you desire). You will need to set up an IAM role with the proper resource policy permissions ( an example of the need for permission can be found [here](aws/iam-policy.json)). This is needed so `ecr-mirror-sync` can interact with AWS when running in Kubernetes (see [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html)). If you decide not to trust the robots, the tool can run locally, displaying table output if desired, and uses the AWS standard credentials mechanism for authenticating.

This leverages ideas and patterns from [Skopeo](https://github.com/containers/skopeo).


## **Features**

-  List repositories with resource identifier tags indicating this repository is to be mirrored from a public image source.
-  Copy a single image:tag into an ECR repository
-  Sync ecr repositories with mirror identifier tags 


## **Installation**

Prior to running this, you'll nee d to ensure there are ecr repositories with the correct resource tags. Upsteam tags can be a `/` seperated list.  

Example 
```bash
upstream-image = "ghcr.io/kedacore/keda"
upstream-tags  = "2.4.0/2.5.0"
```

Set the `ECR_REGISTRY` in Makefile before running and associated commands

### Local (MAC)

```bash
brew install gpgme
make build 
```
### Image 

```bash
make image 
```
### helm 

Set `eks.amazonaws.com/role-arn` and `repository` in values.yaml file before running. If you are not using ISRA you can pass creds via env and app should work. The chart will need to be modified to accommodate this. Currently it only supports using IRSA.

```bash
helm upgrade \
ecr-mirror-sync \ 
./charts/ecr-mirror-sync \
--install \
--debug \
--wait \
--namespace="ibeify-ops" 
```

## **Usage**

### **ecr-mirror-sync list**

*Example*
```bash
ecr-mirror-sync list  
```

```bash
List ECR repositories and tags marked for mirroring

Usage:
  ecr-mirror-sync list [flags]

Flags:
      --batch string       batch size for syncing images, default is all
      --debug              enable debug output
      --dry-run            Run without actually copying data
  -h, --help               help for list
      --image-key string   aws resource tag for upstream image (default "upstream-image")
      --prefix string      prefix for external images in ecr
      --region string      ecr region for to interactive with (default "us-east-1")
      --render-table       Render tables
      --tag-key string     aws resource tag for upstream tags (default "upstream-tags")
```

### **ecr-mirror-sync copy**

*Example*
```bash
ecr-mirror-sync copy --src ghcr.io/kedacore/keda:2.4.0 --dest $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/external/ghcr.io/kedacore/keda --policy=./docker/default-policy.json --render-table --dry-run
```

```bash
Copy image:tag from public source to ECR

Usage:
  ecr-mirror-sync copy [flags]

Flags:
      --batch string                    batch size for syncing images, default is all
      --debug                           enable debug output
  -d, --dest string                     ecr destingation repository
      --dest-precompute-digests         Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dry-run                         Run without actually copying data
  -h, --help                            help for copy
      --image-key string                aws resource tag for upstream image (default "upstream-image")
      --insecure-policy                 run the tool without any policy check
      --override-arch ARCH              use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS                  use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT        use VARIANT instead of the running architecture variant for choosing images
      --policy string                   Path to a trust policy file
      --prefix string                   prefix for external images in ecr
      --region string                   ecr region for to interactive with (default "us-east-1")
      --render-table                    Render tables
      --retry-times int                 the number of times to possibly retry
  -s, --src string                      source image:tag
      --src-authfile string             path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --src-cert-dir PATH               use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --src-creds USERNAME[:PASSWORD]   Use USERNAME[:PASSWORD] for accessing the registry
      --src-no-creds                    Access the registry anonymously
      --src-password string             Password for accessing the registry
      --src-registry-token string       Provide a Bearer token for accessing the registry
      --src-username string             Username for accessing the registry
      --tag-key string                  aws resource tag for upstream tags (default "upstream-tags")

```
### **ecr-mirror-sync sync**

*Example*
```bash
ecr-mirror-sync sync --debug --render-table --src-creds=$DOCKER_USERNAME:$DOCKER_PASSWORD --policy=./docker/default-policy.json --dry-run
```

```bash
Sync all ECR repositories tagged to be mirror with public repositories

Usage:
  ecr-mirror-sync sync [flags]

Flags:
      --batch string                    batch size for syncing images, default is all
      --debug                           enable debug output
      --dest-precompute-digests         Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dry-run                         Run without actually copying data
  -h, --help                            help for sync
      --image-key string                aws resource tag for upstream image (default "upstream-image")
      --insecure-policy                 run the tool without any policy check
      --override-arch ARCH              use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS                  use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT        use VARIANT instead of the running architecture variant for choosing images
      --policy string                   Path to a trust policy file
      --prefix string                   prefix for external images in ecr
      --region string                   ecr region for to interactive with (default "us-east-1")
      --render-table                    Render tables
      --retry-times int                 the number of times to possibly retry
      --src-authfile string             path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --src-cert-dir PATH               use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --src-creds USERNAME[:PASSWORD]   Use USERNAME[:PASSWORD] for accessing the registry
      --src-no-creds                    Access the registry anonymously
      --src-password string             Password for accessing the registry
      --src-registry-token string       Provide a Bearer token for accessing the registry
      --src-username string             Username for accessing the registry
      --tag-key string                  aws resource tag for upstream tags (default "upstream-tags")
```
