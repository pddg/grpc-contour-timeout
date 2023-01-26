function()
  local labels = {
    all: {
      'app.kubernetes.io/name': 'grpcserver',
    },
    selector: {
      'app.kubernetes.io/name': 'grpcserver',
    },
  };
  [
    (import 'deployment.libsonnet')(labels),
    (import 'service.libsonnet')(labels),
    (import 'httpproxy.libsonnet')(),
  ]