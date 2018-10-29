Helm convert plugin
===================

> Charts are curated application definitions for Helm, this plugin let you
convert existing charts into [Kustomize](https://github.com/kubernetes-sigs/kustomize)
compatible package.

[![Build Status](https://travis-ci.org/ContainerSolutions/helm-convert.svg?branch=master)](https://travis-ci.org/ContainerSolutions/helm-convert)
[![Go Report Card](https://goreportcard.com/badge/github.com/ContainerSolutions/helm-convert)](https://goreportcard.com/report/github.com/ContainerSolutions/helm-convert)


## Install

### Helm plugin

```bash
$ helm plugin install https://github.com/ContainerSolutions/helm-convert
```

### Binary without Helm

If you don't have Helm installed, you can just download the binary from the
[release page](https://github.com/ContainerSolutions/helm-convert/releases).

## Usage

See `helm convert --help` for usage. Example:

```bash
# convert the stable/mongodb chart into Kustomize compatible package
helm convert --destination mongodb --name mongodb stable/mongodb

# convert chart from a url
helm convert https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/prometheus-operator

# convert the stable/mongodb chart with a given values.yaml file
helm convert -f values.yaml stable/mongodb

# convert the stable/mongodb chart and override values using --set flag:
helm convert --set persistence.enabled=true stable/mongodb
```

## Docker

You can also execute Helm convert from Docker:

```bash
$ docker run -ti containersol/helm-convert convert --help
```

## Development

```bash
# clone the repo
$ git clone git@github.com:ContainerSolutions/helm-convert.git

# add a symlink in the Helm plugin directory targeting the repository
$ ln -s $PWD ~/.helm/plugins/helm-convert

# build the binary
$ make build

# run
$ helm convert --help

# run lint, vet and tests
$ make test-all
```

## Features

The conversion is currently quite basic and has the following features:

- get image tags and store them in kustomization.yaml
- get common labels and store them in kustomization.yaml
- get resources and store them in kustomization.yaml
- remove helm specific labels from manifests
- remove helm specific annotations from manifests
- get namespace and store it in kustomization.yaml
- create secretGenerator based on secret resources (type Opaque and TLS)
- create secretGenerator based on secret type TLS
- create configGenerator from multiline files
