function()
  {
    apiVersion: 'projectcontour.io/v1',
    kind: 'HTTPProxy',
    metadata: {
      name: 'grpcserver',
    },
    spec: {
      virtualhost: {
        fqdn: "localhost",
      },
      routes: [
        {
          conditions: [
            {
              prefix: '/',
            },
          ],
          timeoutPolicy: {
            response: 'infinity',
            idle: 'infinity',
          },
          services: [
            {
              name: 'api',
              port: 80,
              protocol: 'h2c',
            },
          ],
          loadBalancerPolicy: {
            strategy: 'RoundRobin',
          },
        },
      ],
    },
  }