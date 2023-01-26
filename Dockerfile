FROM quay.io/cybozu/ubuntu:20.04

COPY build/grpcserver /bin/grpcserver
COPY build/grpcclient /bin/grpcclient

ENTRYPOINT ["/bin/grpcserver"]
