Juju should pass JUJU_ACTION, rather than making us parse it.  Just like JUJU_RELATION.

Juju should support uninstall (aka remove) hook

Make sure heka metrics are not analyzed by ES

Close anything closeable after injection?

Singleton injectors?  Pooled injectors?

Juju should return the actual default value

Don't forget to turn on SSL validation in python requests

How do we do 'zero-unit' services?
* Maybe we could create a parent service with the credentials, and then have our services be subordinate charms.  Reduces overhead to 1.


For the proxy client, could we have a subordinate charm (of the serving instance itself)

Does Juju support fallback / wildcard relation hooks, to make it easier to reuse the charm?

Some sort of bug where Juju status fails if there are no units? Seems to affect CLI also.

Juju ServiceDeploy needs to be able to deploy without NumUnits, for subordinate charms


We need to pre-download charms??