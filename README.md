# Test timeout config of contour for gRPC

## Environment

- os: Linux
- arch: amd64
- go: 1.19 or later

## Setup kind cluster

```sh
# Create cluster and setup latest contour
make cluster-up

# Forwarding service/envoy's 80 port to 8888.
# Contour requires service whose `type` is `LoadBalancer`.
# However, kind does not have avility to achieve it.
make port-forward
```

## Run gRPC service

```sh
# Build application providing gRPC service
make

# Create container image and load it to kind cluster
make load-image

# Apply manifests
make apply
```

Now, you can access the service via `localhost:8888`.

```sh
./build/grpcclient liveness --server localhost:8888
```

### Rollout

You should re-build container image and rollout the deployment.

```sh
# Build application providing gRPC service
make


# Create container image and load it to kind cluster
make load-image

# Rollout the deployment
make rollout
```

## Configure timeout

This can be easily changed by editing the following.

https://github.com/pddg/grpc-contour-timeout/blob/master/kubernetes/httpproxy.libsonnet#L19-L21

Then, run `make apply`

```
make apply
```

## Author

- pddg
