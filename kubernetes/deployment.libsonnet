function(labels)
  {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name: 'grpcserver',
      labels: labels.all,
    },
    spec: {
      replicas: 2,
      selector: {
        matchLabels: labels.selector,
      },
      template: {
        metadata: {
          labels: labels.all,
        },
        spec: {
          containers: [
            {
              name: 'grpcserver',
              image: 'grpc-contour-timeout:latest',
              imagePullPolicy: 'IfNotPresent',
              securityContext: {
                readOnlyRootFilesystem: true,
              },
              args: [
                '-server-keepalive',
                '-enforce-keepalive',
              ],
              ports: [
                {
                  containerPort: 8080,
                },
              ],
              livenessProbe: {
                exec: {
                  command: ['/bin/grpcclient', 'liveness', '--server', 'localhost:8080'],
                },
                initialDelaySeconds: 5,
                periodSeconds: 10,
              },
            },
          ],
        },
      },
    },
  }