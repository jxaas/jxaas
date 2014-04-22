## How to create a new service

There are a number of steps, most of which are simple copy-and-paste.

An overview of what happens: you configure a bundle template which configures the Juju services which make up your JXaaS instance.
In your template, you include the charm you're wrapping, along with a stub-client that exposes the relation properties to JXaaS.
You likely also include a monitoring subordinate charm.  Finally, you tell JXaaS about your new service, which currently involves declaring
a class, but will likely become a simple YAML file in future versions!

1. First, make sure you have a charm for the service (that works!)
1. Choose a "bundle-type": this is a short string by which your service will be known (e.g. mysql, pg).  It can't be too long, or else you'll hit filename limitations.
1. Copy one of the existing template files to <bundletype>.yaml.  Change it to reference the correct charm.  Change the root element and your primary Juju service to be <bundletype>
1. Fix up any relations in the template, and generally expose the options you want to allow to be configurable

Configure the service in the JXaaS code (for now, this involves code, in future it likely will involve a YAML file):

1. In the bundletype package, copy one of the existing .go services, and name in <bundletype>.go
1. Rename the class and the New function, and set the appropriate values.  Most are obvious.  IsStarted should test for a relation property that is set when the service is ready.  GetRelationJujuInterface should return the Juju relation contract name (as you're about to configure in the stub charm).
1. In main.go, add your charm to the system.BundleTypes map

Make sure the stub charm supports your interface:

1. In metadata.yaml add a requires definition with your new interface:
```
  pgsql:
    interface: pgsql
```
1. Symlink each of the relation hooks (hopefully we can get rid of this step in future):
```
cd hooks
ln -s relation-broken pgsql-relation-broken
ln -s relation-changed pgsql-relation-changed
ln -s relation-joined pgsql-relation-joined
```

1. That's it.  bzr commit, bzr push and juju publish

...


## Create a proxy charm

1. Copy one of the existing proxy-charms
1. Change config.yaml to expose whatever configuration options your service exposes.
1. Copy the icon.svg from the main charm
1. Edit proxy.yaml to include the bundletype for your service
1. Edit metadata.yaml to include your new proxy name (recommend "<normal-charm-name>-proxy").  Change the provides section to include the primary interface.
1. bzr push --remember to a new repository, bzr commit, bzr push, juju publish


