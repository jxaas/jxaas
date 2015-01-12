# The JXaaS manifest

The JXaaS manifest controls _how_ a Juju charm is transformed into a service.

We will explain the official MySQL manifest here.

```
services:
  mysql:
    charm: "cs:~justin-fathomdb/trusty/mysql"
    num_units: <<
    exposed: true
    options:
      performance: <<
      slow-query-time: <<
      dataset-size: <<

  sc:
    charm: "cs:~justin-fathomdb/trusty/stub-client"
    exposed: true
    options:
      private-port: 3306
      public-port: {{.AssignPublicPort}}
      open-ports: 3306
      protocol: <<

  metrics:
    charm: "cs:~justin-fathomdb/trusty/heka-collector"
```

The services block specifies the Juju services that will be created for each JXaaS service instance.

In this case, when a user requests a MySQL service, we will create a mysql service in Juju, a "stub-client",
and a "heka-collector" for collecting metrics.

The MySQL service should be self-evident - this is the service we are actually going to expose.

The stub-client acts as a client to the MySQL service.  Most Juju services
require a client to actually do anything!  In addition, it sends the connection
information to JXaaS, which JXaaS then provides to the API caller.

The heka metric collector is based on Mozilla's
[Heka](http://blog.mozilla.org/services/2013/04/30/introducing-heka/), a
lightweight metric collector.

For each service, we define the charm, whether or not it is exposed, the number
of units, along with the options to pass to the Juju charm.

The magic "<<" syntax means use the "default value".  `public-port` on the
stub-client is also interesting: it is ```{{.AssignPublicPort}}```.  This is a
Go template, so this is actually a method call into JXaaS.  The
AssignPublicPort method assigns a unique TCP port to the service, which allows
multiple JXaaS services to share a single IP address.  This was, you can run
multiple MySQL services, even if you only have one or a limited number of IPs.

```
relations:
  - - "sc:juju-info"
    - "mysql:juju-info"
  - - "sc:mysql"
    - "mysql:db"

  - - "metrics:juju-info"
    - "mysql:juju-info"
  - - "metrics:elasticsearch"
    - "{{.SystemServices.elasticsearch}}:cluster"

  - - "sc:website"
    - "{{.SystemServices.haproxy}}:reverseproxy"
```

The relations section defines how these Juju services are connected; JXaaS will create the Juju services
defines in the services section, and will then add relations as defined here.

This follows the Juju syntax:

* The relations are 'from service' followed by 'to service'
* Each side of the relation is 'charm name':'interface'

We use some "system services", which are shared Juju services that JXaaS
creates.  These are defined in ```templates/shared.yaml```.  Again, we're using
a Go-template method syntax to call into JXaaS.  In particular, we send our metrics
to a shared instance of elastic search (which JXaaS then queries to allow API clients
to get their logs).  And we actually expose the service by connecting it to a shared haproxy.


```
provides:
  mysql:
    protocol: <<
    host: <<
    private-address: <<
    port: <<
    database: <<
    password: <<
    slave: <<
    user: <<
```

This specifies the properties that we will actually expose to the JXaaS API client.

Most of these are just straight pass-throughs of properties from the main service; so this block can often be omitted altogether.

TODO??

```
options:
  defaults:
    dataset-size: 256M
```

We can specify some default configuration options.

TODO: How other options get exposed

```
checks:
  service-mysql:
    service: mysql
```

The checks section specifies the health checks that we want to run.  In this case, we make sure that the mysql service is still running.

TODO: Is this only on primary?

```
meta:
  primary-relation-key: mysql
```

TODO

```
cloudfoundry:
  credentials:
    jdbcUrl: jdbc:mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
    uri: mysql://{{.user}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}
    name: {{.database}}
    hostname: {{.host}}
    port: {{.port}}
    username: {{.user}}
    password: {{.password}}

  plans:
    nano:
      options:
        dataset-size: 128M
    micro:
      options:
        dataset-size: 256M
    milli:
      options:
        dataset-size: 512M
    uni:
      options:
        dataset-size: 1024M
```

For cloudfoundry integration, we specify the credentials that we will expose to
cloudfoundry bindings.  These should follow the cloudfoundry conventions for
the service type you're creating, so that it can be compatible with existing
clients.

Optionally, we can also define some plans, which have different configuration
options.  In this case, we're defining plans of different sizes.  This is
optional - if you don't define any plans, you'll just get a plan named
'default' with the default options.
