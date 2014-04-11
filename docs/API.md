#XaaS API

The JXaaS API aims to follow the REST conventions.  It currently supports only JSON, and (because it uses Go's JSON parser) is strict - in particular field names must be quoted using double-quotes:

```
{ "Message": "Hello world" }
```

This documents the _public_ API for Juju XaaS (i.e. the calls that clients will make.)

It does not document the _private_ API that is used for internal operations (for example, we use the API so that charms can report the settings of relationships)  That API is considered a private implementation detail, and (more importantly) is in flux. 

## Service operations (CRUD on services)

These are the primary operations on the API.

A JXaaS instance has the following URL path:

```
http[s]://<host>/xaas/<tenant>/services/<service>/<name>
```

`<tenant>` is your tenant ID

`<service>` is your service you are consuming (e.g. mysql)

`<name>` is an identifier you choose to assign to the service (e.g. my-mysql).


The following operations are available:

### GET

Returns the current service.

### PUT

Creates or updates the properties of the service.  This operation is idempotent.

### DELETE

Deletes the service

## Metrics

Located under the subpath `/metrics`

### GET


## Logs

Located under the subpath `/logs`

### GET



