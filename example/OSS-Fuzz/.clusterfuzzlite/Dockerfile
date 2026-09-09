FROM gcr.io/oss-fuzz-base/base-builder-go:v1

COPY . $SRC/tcq-network-protocol
WORKDIR $SRC/tcq-network-protocol
COPY .clusterfuzzlite/build.sh $SRC/
