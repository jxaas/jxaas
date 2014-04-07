Juju should pass JUJU_ACTION, rather than making us parse it.  Just like JUJU_RELATION.

Juju should support uninstall (aka remove) hook

How do we do 'zero-unit' services?
* Maybe we could create a parent service with the credentials, and then have our services be subordinate charms.  Reduces overhead to 1.