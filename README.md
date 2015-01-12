# JXaaS: Juju as-a-service

JXaaS turns any Juju charm into a XaaS.  For example, running JXaaS with the
MySQL charm can power a MySQL-aaS, similar to Amazon RDS.

These XaaS can be consumed through a simple RESTful API, or from another Juju
instance, or from CloudFoundry.

Authentication is pluggable; JXaaS currently supports OpenStack authentication.

JXaaS is written in Golang.  It comprises a server that exposes a RESTful
interface for creating, modifying & destroying services backed by Juju charms.
It launches those services in a private Juju cluster, and exposes the service to
the caller.

Extra features:

* Health-checking and monitoring of the services
* Supports TLS encryption for the exposed connections
* Metrics & logs are collected and can be exposed
* Simple manifest-based configuration
* Can act as a CloudFoundry service broker
* Can authenticate using OpenStack Identity

# Next steps

There is a simple guided tour in docs/tutorial.

If you want to create a service, you can do that using

# Related projects

Check out all the official [JXaaS projects on Github](https://github.com/jxaas)


