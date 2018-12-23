```
   _____ __       ____
  / ___// /____  / / /___ ______
  \__ \/ __/ _ \/ / / __ `/ ___/
 ___/ / /_/  __/ / / /_/ / /
/____/\__/\___/_/_/\__,_/_/

```

Simplified Container Runtime Cluster

Stellar is designed to provide simple container runtime clustering.  One
or more nodes are joined together to create a cluster.  The cluster
is eventually consistent making it ideal for transient workloads or edge
computing where nodes are not always guaranteed to have high bandwidth, low
latency connectivity.

# Why
There are several container platforms and container orchestrators out there.
However, they are too complex for my use.  I like simple infrastructure that
is easy to deploy and manage, tolerates failure cases and is easy to debug
when needed.  With the increased tolerance in failure modes, this comes at
a consistency cost.  It may not be for you.  Use the best tool for your use
case.  Enjoy :)

# Features

- Container execution via [containerd](https://github.com/containerd/containerd)
- Multihost Networking via [CNI](https://github.com/containernetworking/cni)
- Service Discovery via DNS
- Cluster event system via [NATS](https://github.com/nats-io/gnatsd)
- Builtin Proxy using [Radiant](https://github.com/stellarproject/radiant) (zero downtime reloads, canary deploys, health checks, automatic HTTPS)
- Masterless design
- Efficient use of system resources
- Simple daemon deployment

# Downloads
For official releases, see the [Releases](https://github.com/ehazlett/stellar/releases)

You can also grab the latest [Master Build](https://s3.us-east-2.amazonaws.com/stellar-release/latest/stellar-linux-amd64.tar.gz)

# Building
In order to build Stellar you will need the following:

- A [working Go environment](https://golang.org/doc/code.html) (1.11+)
- Protoc 3.x compiler and headers (download at the [Google releases page](https://github.com/google/protobuf/releases))

Once you have the requirements you can build.

If you change / update the protobuf definitions you will need to generate:

`make generate`

To build the binaries (client and server) run:

`make`

## Docker
Alternatively you can use [Docker](https://www.docker.com) to build:

To generate protobuf:

`make docker-generate`

To build binaries:

`make docker-build`

# Running
To run Stellar, once you have a working containerd installation follow these steps:

- Install [Containerd](https://github.com/containerd/containerd#getting-started) version >=1.1
- Build binaries or get a release
- Copy `/bin/sctl` to `/usr/local/bin/`
- Copy `/bin/stellar` to `/usr/local/bin/`
- Copy `/bin/stellar-cni-ipam` to `/opt/containerd/bin/` or `/opt/cni/bin`

First, we will generate a config:

```
$> stellar config > stellar.conf
```

This will produce a default configuration.  Edit the addresses to match your environment.  For
this example we will use the IP `10.0.1.70`.

```
{
    "ConnectionType": "local",
    "ClusterAddress": "10.0.1.70:7946",
    "AdvertiseAddress": "10.0.1.70:7946",
    "Debug": false,
    "NodeID": "dev",
    "GRPCAddress": "10.0.1.70:9000",
    "TLSServerCertificate": "",
    "TLSServerKey": "",
    "TLSClientCertificate": "",
    "TLSClientKey": "",
    "TLSInsecureSkipVerify": false,
    "ContainerdAddr": "/run/containerd/containerd.sock",
    "Namespace": "default",
    "DataDir": "/var/lib/stellar",
    "StateDir": "/run/stellar",
    "Bridge": "stellar0",
    "UpstreamDNSAddr": "8.8.8.8:53",
    "ProxyHTTPPort": 80,
    "ProxyHTTPSPort": 443,
    "ProxyTLSEmail": "",
    "GatewayAddress": "127.0.0.1:9001",
    "EventsAddress": "10.0.1.70:4222",
    "EventsClusterAddress": "10.0.1.70:5222",
    "EventsHTTPAddress": "10.0.1.70:4322",
    "CNIBinPaths": [
        "/opt/containerd/bin",
        "/opt/cni/bin"
    ],
    "Peers": [],
    "Subnet": "172.16.0.0/12"
}
```

To start the initial node run:

```
$> stellar -D server --config stellar.conf
```

To join additional nodes simply add the `AdvertiseAddress` of the first node to the `Peers`
config option of the second node:

For example:

```
{
    "ConnectionType": "local",
    "ClusterAddress": "10.0.1.71:7946",
    "AdvertiseAddress": "10.0.1.71:7946",
    "Debug": false,
    "NodeID": "dev",
    "GRPCAddress": "10.0.1.71:9000",
    "TLSServerCertificate": "",
    "TLSServerKey": "",
    "TLSClientCertificate": "",
    "TLSClientKey": "",
    "TLSInsecureSkipVerify": false,
    "ContainerdAddr": "/run/containerd/containerd.sock",
    "Namespace": "default",
    "DataDir": "/var/lib/stellar",
    "StateDir": "/run/stellar",
    "Bridge": "stellar0",
    "UpstreamDNSAddr": "8.8.8.8:53",
    "ProxyHTTPPort": 80,
    "ProxyHTTPSPort": 443,
    "ProxyTLSEmail": "",
    "GatewayAddress": "127.0.0.1:9001",
    "EventsAddress": "10.0.1.71:4222",
    "EventsClusterAddress": "10.0.1.71:5222",
    "EventsHTTPAddress": "10.0.1.71:4322",
    "CNIBinPaths": [
        "/opt/containerd/bin",
        "/opt/cni/bin"
    ],
    "Peers": ["10.0.1.70:7946"],
    "Subnet": "172.16.0.0/12"
}
```

You will now have a two node cluster.  To see node information, use `sctl`.

```
$> sctl --addr 10.0.1.70:9000 cluster nodes
NAME                ADDR                OS                       UPTIME              CPUS                MEMORY (USED)
stellar-00          10.0.1.70:9000      Linux (4.17.0-3-amd64)   7 seconds           2                   242 MB / 2.1 GB
stellar-01          10.0.1.71:9000      Linux (4.17.0-3-amd64)   6 seconds           2                   246 MB / 2.1 GB
```

# Deploying an Application
To deploy an application, create an application config.  For example, create the following as `example.conf`:

```json
{
    "name": "example",
    "labels": [
        "env=prod",
        "region=us-east"
    ],
    "services": [
        {
            "name": "redis",
            "image": "docker.io/library/redis:alpine",
            "runtime": "io.containerd.runtime.v1.linux",
            "process": {
                "uid": 0,
                "gid": 0,
                "args": ["redis-server"]
            },
            "labels": [
                "env=prod"
            ],
            "network": true
        }
    ]
}

```

Then run the following to deploy:

```
$> sctl --addr 10.0.1.70:9000 apps create -f ./example.conf
```

You should now see the application deployed:

```
$> sctl --addr 10.0.1.70:9000 apps list
NAME                SERVICES
example             1

$> sctl --addr 10.0.1.70:9000 apps inspect example
Name: example
Services:
  - Name: example.redis
    Image: docker.io/library/redis:alpine
    Runtime: io.containerd.runtime.v1.linux
    Snapshotter: overlayfs
    Labels:
      containerd.io/restart.status=running
      stellar.io/application=example
      stellar.io/network=true

```

By default all applications that have networking enabled will have a corresponding nameserver record
created.  To view the records use the following:

```
$> sctl --addr 10.0.1.70:9000 nameserver list
NAME                    TYPE                VALUE                                            OPTIONS
example.redis.stellar   A                   172.16.0.4
example.redis.stellar   TXT                 node=stellar-00; updated=2018-09-08T10:71:02-04:00
```
