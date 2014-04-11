Juju XaaS allows a juju charm to be exposed easily as-a-service.

A simple manifest file, similar to a Juju bundle, determines which charms and relationships are deployed when API clients create JXaaS services.

For example, a MySQL JXaaS service consists of a MySQL charm, and two subordinate charms: one to collect metrics, and the other a stub client charm.  The stub client charm allows reuse of an existing charm's logic for exposing a service.  For MySQL, it consumes the mysql relationship interface, and it captures the database/username/password that the MySQL charm creates.

This allows the MySQL charm to produce a MySQL-aaS.  An API is automatically created that allows for creation, deletion, scaling, monitoring etc of MySQL instances.  The back-end lifting benefits from the existing robust infrastructure of Juju and its charms.

## The Manifest Template

The manifest template specifies the charms (and their relations) that are deployed when an API service is created.  The format is similar to the Juju bundle format, but with some changes to allow for aaS functionality.  (And also because the Juju bundle code is in Python, but JXaaS is in Go).

The manifest is itself a Go template, which allows for some simple logic.

In addition, the @ shortcut allows for the 'default mapping' for a field.  For example, if you specify:

```
mysql:
  num_units: @
  options:
    memory: @
```

The memory option will be mapped to the option of the service.  num_units will be mapped to the number of units of the service.


## Isolation

Although the bundle specifies the names of services using simple names, these are transformed by JXaaS to allow consumption by multiple clients.

For example, a service "web" that is part of a wordpress-as-a-service bundle will likely become a Juju service name  like "u1234567890abcdef-wordpress-clientname-web".

The long hex string is a per-tenant UUID.  'wordpress' is the name of the service, so that multiple services can run against the same Juju instance.  'clientname' is a name specified by the client.  And (finally!) 'web' is the name of the service in the bundle.

Currently, all tenants are deployed into a single shared Juju system, and JXaaS is responsible for enforcing e.g. security.

Services can be shared; for example we use a shared ElasticSearch Juju service as the backend for log & metrics data.

## Proxy Client

The Juju XaaS can be used by API clients: using a binding to the API, services can be created and connected to.  But, for clients that are using Juju, a more complete solution is available.  A proxy charm can be easily created for each service; it creates a service through the XaaS API, and then fulfills the contract for consumption by Juju.  For example, the mysql-proxy charm creates a MySQL service in the JXaaS, and then maps the properties so that it fulfills the mysql juju interface contract.