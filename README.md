# JXaaS: Juju as-a-service

JXaaS turns any Juju charm into a service - we call this XaaS.

For example, running JXaaS with the MySQL charm can power a MySQL-aaS, similar to Amazon RDS.

JXaaS automatically gives you a simple RESTful API, which can be used from a [command-line tool](https://github.com/jxaas/cli),
or from another Juju system, or from CloudFoundry.

Authentication is pluggable; JXaaS currently supports OpenStack authentication.

JXaaS is written in Golang.  It comprises a server that exposes a RESTful
interface for creating, modifying & destroying services backed by Juju charms.
It launches those services in a private Juju cluster, and exposes the service to
the caller.

Highlights:

* Health-checking and monitoring of the services
* Supports TLS encryption for the exposed connections
* Metrics & logs are collected and can be exposed
* Simple manifest-based configuration
* Can act as a CloudFoundry service broker
* Can authenticate using OpenStack Identity

# Getting started

You should probably read or try out the [tutorial](docs/tutorial).

If you want to create a service, you should read about the [manifest](docs/manifest)

# Related projects

Check out all the official [JXaaS projects on Github](https://github.com/jxaas)

Notable projects:

* [The command-line interface](https://github.com/jxaas/cli)
* [Python client bindings for the API](https://github.com/jxaas/python-client)
* [The integration test suite](https://github.com/jxaas/jxaas-tests)


