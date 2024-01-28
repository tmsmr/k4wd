# K4wd
*`kubectl port-forward` on steroids*

**Work in Progress! Docs are incomplete and this tool is not properly tested yet**

## Purpose
K4wd allows to make multiple Resources in a Kubernetes cluster available locally for development and debugging purposes in a pleasant way.

## Quickstart / Demo
- Have some [Services](https://kubernetes.io/docs/concepts/services-networking/service/) deployed, e.g.:
```bash
cd docs
kubectl apply -f example.yaml
```
- Write a `Forwardfile`, e.g.:
```toml
relaxed = true

[forwards.upstream-a]
service = "upstream-a"
namespace = "k4wd"
remote = "http-alt"
local = "8080"

[forwards.upstream-b]
service = "upstream-b"
namespace = "k4wd"
remote = "8080"
local = "0.0.0.0:8081"

[forwards.upstream-c]
service = "upstream-c"
namespace = "k4wd"
remote = "http-alt"
```
- Start `k4wd`, e.g.:
```bash
$ k4wd
INFO[0000] loaded Forwardfile (relaxed=true) containing 3 entries: (upstream-b, upstream-c, upstream-a) 
INFO[0000] created Kubeclient for https://127.0.0.1:6443 (/home/thomas/.kube/config) 
INFO[0000] initialized Envfile for /tmp/k4wd_env_9c2bed59bac37ab739fca89c25e8cddfc25dd0568537c7c864751d3948962afb 
Forwarding from 127.0.0.1:8080 -> 1234
INFO[0000] upstream-a ready: 127.0.0.1:8080 -> k4wd, upstream-a-5bcdb8b947-m2f9z, 1234 
Forwarding from 0.0.0.0:8081 -> 1234
INFO[0000] upstream-b ready: 0.0.0.0:8081 -> k4wd, upstream-b-577855f5c7-frwjm, 1234 
Forwarding from 127.0.0.1:57079 -> 1234
INFO[0000] upstream-c ready: 127.0.0.1:57079 -> k4wd, upstream-c-6c658678ff-w6cf2, 1234 
```
*Note that for upstream-c a random free port was assigned, since no value is defined in the `Forwardfile`*
- (Optional) Get a new shell in the same context and request the active forwards as env variables, e.g.:
```bash
$ k4wd -e
# Name: 'upstream-a', Type: 1, Namespace: 'k4wd', Pod: 'upstream-a-5bcdb8b947-m2f9z', Service: 'upstream-a', Remote: 'http-alt', Local: '8080', LocalAddr: '127.0.0.1', LocalPort: 8080, TargetPort: 1234, Active: true
UPSTREAM_A_ADDR=127.0.0.1:8080

# Name: 'upstream-b', Type: 1, Namespace: 'k4wd', Pod: 'upstream-b-577855f5c7-frwjm', Service: 'upstream-b', Remote: '8080', Local: '0.0.0.0:8081', LocalAddr: '0.0.0.0', LocalPort: 8081, TargetPort: 1234, Active: true
UPSTREAM_B_ADDR=0.0.0.0:8081

# Name: 'upstream-c', Type: 1, Namespace: 'k4wd', Pod: 'upstream-c-6c658678ff-w6cf2', Service: 'upstream-c', Remote: 'http-alt', Local: '', LocalAddr: '127.0.0.1', LocalPort: 57079, TargetPort: 1234, Active: true
UPSTREAM_C_ADDR=127.0.0.1:57079
```
- (Optional) Or write the env variables to a file, e.g.:
```bash
$ k4wd -e -p .env
$ cat .env
# Name: 'upstream-a', Type: 1, Namespace: 'k4wd', Pod: 'upstream-a-5bcdb8b947-m2f9z', Service: 'upstream-a', Remote: 'http-alt', Local: '8080', LocalAddr: '127.0.0.1', LocalPort: 8080, TargetPort: 1234, Active: true
UPSTREAM_A_ADDR=127.0.0.1:8080

# Name: 'upstream-b', Type: 1, Namespace: 'k4wd', Pod: 'upstream-b-577855f5c7-frwjm', Service: 'upstream-b', Remote: '8080', Local: '0.0.0.0:8081', LocalAddr: '0.0.0.0', LocalPort: 8081, TargetPort: 1234, Active: true
UPSTREAM_B_ADDR=0.0.0.0:8081

# Name: 'upstream-c', Type: 1, Namespace: 'k4wd', Pod: 'upstream-c-6c658678ff-w6cf2', Service: 'upstream-c', Remote: 'http-alt', Local: '', LocalAddr: '127.0.0.1', LocalPort: 57079, TargetPort: 1234, Active: true
UPSTREAM_C_ADDR=127.0.0.1:57079
```
- Access the upstream services, e.g.:
```bash
$ . .env && curl $UPSTREAM_A_ADDR -I
HTTP/1.0 200 OK
Server: SimpleHTTP/0.6 Python/3.12.1
Date: Sat, 13 Jan 2024 14:55:28 GMT
Content-type: text/html; charset=utf-8
Content-Length: 840
```
- Stop the `k4wd` process
- Clean up, e.g.:
```bash
kubectl delete -f example.yaml
```

## k3d setup for GH actions...TBD
- `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash`
- `k3d cluster create mycluster`
