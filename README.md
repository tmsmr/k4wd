# K4wd
[![Go](https://github.com/tmsmr/k4wd/actions/workflows/push.yml/badge.svg)](https://github.com/tmsmr/k4wd/actions/workflows/push.yml)
[![Go Report](https://goreportcard.com/badge/github.com/tmsmr/k4wd)](https://goreportcard.com/report/github.com/tmsmr/k4wd)

*`kubectl port-forward` on steroids*

**Work in Progress! Docs are incomplete and this tool is not properly tested yet**

## Purpose
K4wd allows to make multiple Resources in a Kubernetes cluster available locally for development and debugging purposes in a pleasant way.

## Quickstart / Demo
- Install `k4wd`, e.g.:
```bash
go install github.com/tmsmr/k4wd/cmd/k4wd@latest
```
- Have some Resources deployed, e.g.:
```bash
kubectl apply -f docs/example.yaml
```
- Write a `Forwardfile`, e.g. `docs/Forwardfile`:
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
- Start `k4wd`, e.g.:
```
$ k4wd -f docs/Forwardfile 
INFO[09:02:47] starting 3 forwards
INFO[09:02:47] nginx-service ready (127.0.0.1:8080 -> k4wd/nginx-77b4fdf86c-f4wt6:80) 
INFO[09:02:47] nginx-pod ready (127.0.0.1:1234 -> k4wd/nginx:80) 
INFO[09:02:47] nginx-deployment ready (127.0.0.1:49758 -> k4wd/nginx-77b4fdf86c-f4wt6:80)
```
*Note that for nginx-deployment a random free port was assigned, since no value is defined in the `Forwardfile`*
- (Optional) Get a new shell and request the active forwards as env variables, e.g.:
```
$ k4wd -f docs/Forwardfile -e
export NGINX_SERVICE_ADDR=127.0.0.1:8080
export NGINX_POD_ADDR=127.0.0.1:1234
export NGINX_DEPLOYMENT_ADDR=127.0.0.1:0
```
- Use the Forwards, e.g.:
```
$ k4wd -f docs/Forwardfile -e > .env
$ . .env && curl $NGINX_SERVICE_ADDR -I
HTTP/1.1 200 OK
Server: nginx/1.25.3
Date: Sun, 11 Feb 2024 08:05:56 GMT
Content-Type: text/html
Content-Length: 615
Last-Modified: Tue, 24 Oct 2023 13:46:47 GMT
Connection: keep-alive
ETag: "6537cac7-267"
Accept-Ranges: bytes
```
- Stop the `k4wd` process
- Clean up, e.g.:
```
kubectl delete -f docs/example.yaml
```
