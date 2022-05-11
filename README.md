# ECR-MIRROR-SYNC

## **Overview**

The `ecr-mirror-sync` tool allows you to sync publicly accessible images into ECR. `ecr-mirror-sync` extracts a list of ecr repositories with unique identifier tags `upstream-image` and `upstream-tags`.It then preforms a check to determine if an image already exist in the repository followed by another to determine if the digest found matches the currently available public image digest. By default request made to public registries are anonymously, you however can pass credentials to authenticate.This tool was meant to run in Kubernetes as a Cronjob, scheduled for nightly repository syncing.You will need to set up an iam role with the proper resource policies permissions. This is needed so `ecr-mirror-sync` can interact with AWS when running in Kubernetes (see [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html)). If you decide not to trust the robots, the tool can run locally, displaying table output if desired and uses the aws standard credentials mechanism for authenticating.This leverages ideas and patterns from [Skopeo](https://github.com/containers/skopeo).

## Features

-  List repositories with unique identifier tags in ECR.
-  Copy a single image:tag into an ECR repository
-   Sync all images with unique identifier tags with their corresponding public image


## Installation

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

Set `eks.amazonaws.com/role-arn` and `repository` in values.yaml file before running. If you are not using ISRA you can pass creds via env and app should work. We're only support using IRSA here.

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
  ecr-mirror-sync list 

Flags:
  -h, --help   help for list

Global Flags:
      --insecure-policy            run the tool without any policy check
      --override-arch ARCH         use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS             use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT   use VARIANT instead of the running architecture variant for choosing images
      --policy string              Path to a trust policy file
```

### **ecr-mirror-sync copy**

```
ecr-mirror-sync copy --src ghcr.io/kedacore/keda:2.4.0 --dest 928314642453.dkr.ecr.us-east-1.amazonaws.com/external/ghcr.io/kedacore/keda --insecure-policy --render-table --dry-run
```

```
Copy image from public source to ECR

Usage:
  ecr-mirror-sync copy [flags]

Flags:
      --debug                           enable debug output
  -d, --dest string                     ecr destingation repository
      --dest-precompute-digests         Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dry-run                         Run without actually copying data
  -h, --help                            help for copy
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

Global Flags:
      --insecure-policy            run the tool without any policy check
      --override-arch ARCH         use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS             use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT   use VARIANT instead of the running architecture variant for choosing images
      --policy string              Path to a trust policy file
```
### **ecr-mirror-sync sync**

```
ecr-mirror-sync sync --debug --render-table --src-creds=$DOCKER_USERNAME:$DOCKER_PASSWORD --insecure-policy --dry-run
```

```
Sync all ECR repositories tagged to be mirror with public repository

Usage:
  ecr-mirror-sync sync [flags]

Flags:
      --authfile string                       path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --credtype string                       Registry credentials can be used with.Only one can be specified. (default "docker")
      --debug                                 enable debug output
      --dest-authfile string                  path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --dest-cert-dir PATH                    use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --dest-compress                         Compress tarball image layers when saving to directory using the 'dir' transport. (default is same compression type as source)
      --dest-compress-format FORMAT           FORMAT to use for the compression
      --dest-compress-level LEVEL             LEVEL to use for the compression
      --dest-creds USERNAME[:PASSWORD]        Use USERNAME[:PASSWORD] for accessing the registry
      --dest-daemon-host HOST                 use docker daemon host at HOST (docker-daemon: only)
      --dest-decompress                       Decompress tarball image layers when saving to directory using the 'dir' transport. (default is same compression type as source)
      --dest-no-creds                         Access the registry anonymously
      --dest-oci-accept-uncompressed-layers   Allow uncompressed image layers when saving to an OCI image using the 'oci' transport. (default is to compress things that aren't compressed)
      --dest-password string                  Password for accessing the registry
      --dest-precompute-digests               Precompute digests to prevent uploading layers already on the registry using the 'docker' transport. (default true)
      --dest-registry-token string            Provide a Bearer token for accessing the registry
      --dest-shared-blob-dir DIRECTORY        DIRECTORY to use to share blobs across OCI repositories
      --dest-username string                  Username for accessing the registry
      --dry-run                               Run without actually copying data
  -h, --help                                  help for sync
      --render-table                          Render tables
      --retry-times int                       the number of times to possibly retry
      --src-authfile string                   path of the authentication file. Default is ${XDG_RUNTIME_DIR}/containers/auth.json
      --src-cert-dir PATH                     use certificates at PATH (*.crt, *.cert, *.key) to connect to the registry or daemon
      --src-creds USERNAME[:PASSWORD]         Use USERNAME[:PASSWORD] for accessing the registry
      --src-daemon-host HOST                  use docker daemon host at HOST (docker-daemon: only)
      --src-no-creds                          Access the registry anonymously
      --src-password string                   Password for accessing the registry
      --src-registry-token string             Provide a Bearer token for accessing the registry
      --src-shared-blob-dir DIRECTORY         DIRECTORY to use to share blobs across OCI repositories
      --src-username string                   Username for accessing the registry

Global Flags:
      --insecure-policy            run the tool without any policy check
      --override-arch ARCH         use ARCH instead of the architecture of the machine for choosing images (default "amd64")
      --override-os OS             use OS instead of the running OS for choosing images (default "linux")
      --override-variant VARIANT   use VARIANT instead of the running architecture variant for choosing images
      --policy string              Path to a trust policy file
      --registries-conf string     path to the registries.conf file
      --registries.d DIR           use registry configuration files in DIR (e.g. for container signature storage)
```