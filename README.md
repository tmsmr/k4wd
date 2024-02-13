# K4wd [keɪ ˈfɔrwərd]
[![Go](https://github.com/tmsmr/k4wd/actions/workflows/push.yml/badge.svg)](https://github.com/tmsmr/k4wd/actions/workflows/push.yml)
[![Go Report](https://goreportcard.com/badge/github.com/tmsmr/k4wd)](https://goreportcard.com/report/github.com/tmsmr/k4wd)

*`kubectl port-forward` on steroids*

**Docs are WiP**

## General
*K4wd* allows to make multiple resources in Kubernetes clusters available locally for development and debugging purposes in a pleasant way.
While there are many similar tools available, *K4wd* might fill a niche. The primary goals are:
- No need to install additional software in the clusters
- No elevated privileges or additional software on the client
- Declarative configuration for complex setups
- Easy integration in development workflows

## Quickstart / Demo
- Install *k4wd*: `go install github.com/tmsmr/k4wd/cmd/k4wd@latest`
- Have some resources deployed: `kubectl apply -f docs/example.yaml`
- Write a *Forwardfile* (`docs/Forwardfile`):
```toml
[forwards.nginx-pod]
pod = "nginx"
namespace = "k4wd"
remote = "80"
local = "1234"

[forwards.nginx-deployment]
deployment = "nginx"
namespace = "k4wd"
remote = "80"

[forwards.nginx-service]
service = "nginx"
namespace = "k4wd"
remote = "http-alt"
local = "8080"

```
- Start *k4wd*:
```
$ k4wd -f docs/Forwardfile
INFO[09:02:47] starting 3 forwards
INFO[09:02:47] nginx-service ready (127.0.0.1:8080 -> k4wd/nginx-77b4fdf86c-f4wt6:80) 
INFO[09:02:47] nginx-pod ready (127.0.0.1:1234 -> k4wd/nginx:80) 
INFO[09:02:47] nginx-deployment ready (127.0.0.1:49758 -> k4wd/nginx-77b4fdf86c-f4wt6:80)
```
*Note that for nginx-deployment, a random free port was assigned since no value is defined in the Forwardfile*
- (Optional) Get a new shell and request the active forwards as env variables, e.g.:
```
$ k4wd -f docs/Forwardfile -e
export NGINX_SERVICE_ADDR=127.0.0.1:8080
export NGINX_POD_ADDR=127.0.0.1:1234
export NGINX_DEPLOYMENT_ADDR=127.0.0.1:0
```
- Use the forwards, e.g.:
```
$ eval $(k4wd -f docs/Forwardfile -e)
$ curl $NGINX_SERVICE_ADDR -I
HTTP/1.1 200 OK
...
```
- Stop the `k4wd` process and clean up: `kubectl delete -f docs/example.yaml`

## Installation
For every tagged version the binaries are attached to the release. You can download it and put it in your `$PATH`.
Releases are available [here](https://github.com/tmsmr/k4wd/releases). While no binaries are signed, you can verify the checksum from *SHA256SUMS.txt*.

If you have *go* installed, you can also install *k4wd* using `go install github.com/tmsmr/k4wd/cmd/k4wd@latest`.

### Usage
```
$ k4wd -h
Usage of k4wd:
  -d    enable debug logging
  -e    print environment instead of running k4wd
  -f string
        path to Forwardfile (context) (default "Forwardfile")
  -k string
        alternative path to kubeconfig
  -o string
        output format for environment (env, no-export, json, ps, cmd) (default "env")
```

## Forwardfile
__TBD__

## Limitations
__TBD__

## Disclaimer
Check *LICENSE* for details. If this tool eats your dog, it's not my fault.
