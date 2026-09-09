FROM gcr.io/oss-fuzz-base/base-builder-go:v1
RUN git clone --depth 1 https://github.com/NguyenHien-8/tcq-network-protocol $SRC/tcq-network-protocol
WORKDIR $SRC/tcq-network-protocol
COPY build.sh $SRC/