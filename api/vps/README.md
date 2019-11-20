# Vps Service

This is the Vps service

Generated with

```
micro new github.com/John-Tonny/mnhost/api/vps --namespace=go.micro --alias=vps --type=web
```

## Getting Started

- [Configuration](#configuration)
- [Dependencies](#dependencies)
- [Usage](#usage)

## Configuration

- FQDN: go.micro.web.vps
- Type: web
- Alias: vps

## Dependencies

Micro services depend on service discovery. The default is multicast DNS, a zeroconf system.

In the event you need a resilient multi-host setup we recommend consul.

```
# install consul
brew install consul

# run consul
consul agent -dev
```

## Usage

A Makefile is included for convenience

Build the binary

```
make build
```

Run the service
```
./vps-web
```

Build a docker image
```
make docker
```