ARG BASE_IMAGE=golang:1.16


FROM ${BASE_IMAGE} as build


ARG VERSION

RUN apt-get update \
    && apt-get install -y \
    git \
    libgpgme-dev \
    libassuan-dev \
    libbtrfs-dev \
    libdevmapper-dev \
    libostree-dev \
    libvshadow-utils \
    pkg-config


ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED=0

WORKDIR /go/src/ecr-mirror-sync
COPY . .

RUN go build -v -o /ecr-mirror-sync -ldflags="-X 'github.com/kaladaOpuiyo/ecr-mirror-sync/version.Version=${VERSION}'" -tags 'btrfs_noversion libdm_no_deferred_remove containers_image_openpgp' ./main.go 

FROM  alpine:3.15.4

COPY --from=build /ecr-mirror-sync /usr/local/bin/ecr-mirror-sync


ENTRYPOINT [ "ecr-mirror-sync" ]
