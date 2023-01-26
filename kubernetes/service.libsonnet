function(labels)
  {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'api',
      labels: labels.all,
    },
    spec: {
      ports: [
        {
          port: 80,
          targetPort: 8080,
        },
      ],
      selector: labels.selector,
    },
  }
