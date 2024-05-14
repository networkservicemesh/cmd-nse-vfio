# Intro

This repo contains 'cmd-nse-vfio' a VFIO NSE application for Network Service Mesh. It provides `Network Service -> {
MAC Address, VLAN tag }` mappings for the list of registered Network Services.

This README will provide directions for building, testing, and debugging that container.

# Usage

`cmd-nse-vfio` accept following environment variables:

* `NSM_NAME` - A string value of network service endpoint name (default "vfio-server")
* `NSM_BASE_DIR` - A base directory to create a unix socker for listening incoming requests (default "./")
* `NSM_CONNECT_TO` - A Network service Manager connectTo URL (default "unix:///var/lib/networkservicemesh/nsm.io.sock")
* `NSM_MAX_TOKEN_LIFETIME` - A token lifetime duration (default 24h)
* `NSM_SERVICE_NAMES` - A list of supported Network Services in inner format:
    Name@Domain: { addr: MACAddr; vlan: VLANTag; labels: Labels; }
    MACAddr = xx:xx:xx:xx:xx:xx
    Labels = label_1=value_1&label_2=value_2
        - Name - a Network Service name
        - Domain - a Network Service domain (don't confuse it with interdomain domains)
        - MACAddr - a MAC address for the Network Service
        - VLANTag - a VLAN tag for the Network Service
        - labelN=valueN - pairs of labels supported by the Network Service
    - Examples:
        - pingpong@worker.domain: { addr: 0a:55:44:33:22:11 }
            - **pingpong** Network Service
            - **worker.domain** Network Service domain
            - **0a:55:44:33:22:11** MAC address
* `NSM_CIDR_PREFIX`              - List of CIDR Prefix to assign IPv4 and IPv6 addresses from (default: "169.254.0.0/16")
* `NSM_LABELS`                   - Endpoint labels
* `NSM_LOG_LEVEL`                - Log level (default: "INFO")
* `NSM_METRICS_EXPORT_INTERVAL`  - interval between mertics exports (default: "10s")
* `NSM_OPEN_TELEMETRY_ENDPOINT`  - OpenTelemetry Collector Endpoint (default: "otel-collector.observability.svc.cluster.local:4317")
* `NSM_PAYLOAD`                  - Name of provided service payload (default: "ETHERNET")
* `NSM_REGISTER_SERVICE`         - if true then registers network service on startup (default: "true")
* `NSM_REGISTRY_CLIENT_POLICIES` - paths to files and directories that contain registry client policies (default: "etc/nsm/opa/common/.*.rego,etc/nsm/opa/registry/.*.rego,etc/nsm/opa/client/.*.rego")


# Build

## Build cmd binary locally

You can build the locally by executing

```bash
go build ./...
```

## Build Docker container

You can build the docker container by running:

```bash
docker build .
```

# Testing

## Testing Docker container

Testing is run via a Docker container.  To run testing run:

```bash
docker run --rm $(docker build -q --target test .)
```

# Debugging

## Debugging the tests
If you wish to debug the test code itself, that can be acheived by running:

```bash
docker run --rm -p 40000:40000 $(docker build -q --target debug .)
```

This will result in the tests running under dlv.  Connecting your debugger to localhost:40000 will allow you to debug.

```bash
-p 40000:40000
```
forwards port 40000 in the container to localhost:40000 where you can attach with your debugger.

```bash
--target debug
```

Runs the debug target, which is just like the test target, but starts tests with dlv listening on port 40000 inside the container.

## Debugging the cmd

When you run 'cmd' you will see an early line of output that tells you:

```Setting env variable DLV_LISTEN_FORWARDER to a valid dlv '--listen' value will cause the dlv debugger to execute this binary and listen as directed.```

If you follow those instructions when running the Docker container:
```bash
docker run -e DLV_LISTEN_FORWARDER=:50000 -p 50000:50000 --rm $(docker build -q --target test .)
```

```-e DLV_LISTEN_FORWARDER=:50000``` tells docker to set the environment variable DLV_LISTEN_FORWARDER to :50000 telling
dlv to listen on port 50000.

```-p 50000:50000``` tells docker to forward port 50000 in the container to port 50000 in the host.  From there, you can
just connect dlv using your favorite IDE and debug cmd.

## Debugging the tests and the cmd

```bash
docker run -e DLV_LISTEN_FORWARDER=:50000 -p 40000:40000 -p 50000:50000 --rm $(docker build -q --target debug .)
```

Please note, the tests **start** the cmd, so until you connect to port 40000 with your debugger and walk the tests
through to the point of running cmd, you will not be able to attach a debugger on port 50000 to the cmd.
