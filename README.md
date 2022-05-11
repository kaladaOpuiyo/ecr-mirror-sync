# ECR-MIRROR-SYNC

## **Overview**

The `ecr-mirror-sync` tool allows you to sync publicly accessible images into ECR. `ecr-mirror-sync` extracts a list of ecr repositories with unique identifier tags `upstream-image` and `upstream-tags`.It then preforms a check to determine if an image already exist in the repository followed by another to determine if the digest found matches the currently available public image digest. By default request made to public registries are anonymously, you however can pass credentials to authenticate.This tool was meant to run in Kubernetes as a Cronjob, scheduled for nightly repository syncing.You will need to set up an iam role with the proper resource policies permissions. This is needed so `ecr-mirror-sync` can interact with AWS when running in Kubernetes (see [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html)). If you decide not to trust the robots, the tool can run locally, displaying table output if desired and uses the aws standard credentials mechanism for authenticating.This leverages ideas and patterns from [Skopeo](https://github.com/containers/skopeo).

## Features

-  List repositories with unique identifier tags in ECR.
-  Copy a single image:tag into an ECR repository
-  Sync all images with unique identifier tags with their corresponding public image


## Installation

Prior to running this, you'll need to ensure there is are ecr repositories the correct resource tags. Upsteam tags can be a `/` seperated list.  

Example 
```
upstream_image = "ghcr.io/kedacore/keda"
upstream_tags  = "2.4.0/2.5.0"
```

Set the `ECR_REGISTRY` in Makefile before running 

### Local (MAC)
```
brew install gpgme
make build 
```
### Image 

```
make image 
```
### helm 

Set `eks.amazonaws.com/role-arn` and `repository` in values.yaml file before running. If you are not using ISRA you can pass creds via env and app should work. We only support using IRSA here.

```
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
```
ecr-mirror-sync list  
```

```
List ECR repositories and tags marked for mirroring

Usage:
  ecr-mirror-sync list [flags]

Flags:
  -h, --help   help for list
```

### **ecr-mirror-sync copy**

```
ecr-mirror-sync copy --src ghcr.io/kedacore/keda:2.4.0 --dest $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/external/ghcr.io/kedacore/keda --insecure-policy --render-table --dry-run
```

```
Copy image:tag from public source to ECR

Usage:
  ecr-mirror-sync copy [flags]

Flags:
      --debug                           enable debug output
  -d, --dest string                     ecr destingation repository
      --dest-precompute-digests         Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dry-run                         Run without actually copying data
  -h, --help                            help for copy
      --insecure-policy                 run the tool without any policy check
      --override-arch ARCH              use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS                  use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT        use VARIANT instead of the running architecture variant for choosing images
      --policy string                   Path to a trust policy file
      --render-table                    Render tables
      --retry-times int                 the number of times to possibly retry
  -s, --src string                      public Docker hub image:tag source
      --src-authfile string             path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --src-cert-dir PATH               use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --src-cred-type string            Registry type. Defaults to docker (default "docker")
      --src-creds USERNAME[:PASSWORD]   Use USERNAME[:PASSWORD] for accessing the registry
      --src-daemon-host HOST            use docker daemon host at HOST (docker-daemon: only)
      --src-no-creds                    Access the registry anonymously
      --src-password string             Password for accessing the registry
      --src-registry-token string       Provide a Bearer token for accessing the registry
      --src-shared-blob-dir DIRECTORY   DIRECTORY to use to share blobs across OCI repositories
      --src-username string             Username for accessing the registry

```
### **ecr-mirror-sync sync**

```
ecr-mirror-sync sync --debug --render-table --src-creds=$DOCKER_USERNAME:$DOCKER_PASSWORD --insecure-policy --dry-run
```

```
Sync all ECR repositories tagged to be mirror with public repositories

Usage:
  ecr-mirror-sync sync [flags]

Flags:
      --debug                           enable debug output
      --dest-precompute-digests         Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dry-run                         Run without actually copying data
  -h, --help                            help for sync
      --insecure-policy                 run the tool without any policy check
      --override-arch ARCH              use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS                  use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT        use VARIANT instead of the running architecture variant for choosing images
      --policy string                   Path to a trust policy file
      --render-table                    Render tables
      --retry-times int                 the number of times to possibly retry
      --src-authfile string             path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --src-cert-dir PATH               use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --src-cred-type string            Registry type. Defaults to docker (default "docker")
      --src-creds USERNAME[:PASSWORD]   Use USERNAME[:PASSWORD] for accessing the registry
      --src-daemon-host HOST            use docker daemon host at HOST (docker-daemon: only)
      --src-no-creds                    Access the registry anonymously
      --src-password string             Password for accessing the registry
      --src-registry-token string       Provide a Bearer token for accessing the registry
      --src-shared-blob-dir DIRECTORY   DIRECTORY to use to share blobs across OCI repositories
      --src-username string             Username for accessing the registry
```