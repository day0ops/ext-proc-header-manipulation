# Ext Proc Header Manipulation

This project demonstrates how to use [External Processor filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) to manipulate the request/response headers processed by Envoy.

It looks for the header `instructions` and the value in the following format. This header payload will allow you to change the request or response headers. 

```
{
  "addHeaders": {
     "header1": "value1"
  }
  "removeHeaders": [
    "headers3"
  ]
}
```

## Build

Setting `REPO` will override the container image location. Otherwise by default it is set to `australia-southeast1-docker.pkg.dev/solo-test-236622/apac`.

```bash
make all
```

## Test

Run the full test suite with,

```bash
make test
```

The test suite uses [testcontainers](https://golang.testcontainers.org/)

### Fedora & Podman

On Fedora need to set `DOCKER_HOST` to point to the socket

```bash
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
```

When integrating with Podman, follow the [guide](https://github.com/containers/podman/blob/main/docs/tutorials/socket_activation.md) for socket activation and to start use,

```bash
systemctl --user start podman.socket

ls $XDG_RUNTIME_DIR/podman/podman.sock
```