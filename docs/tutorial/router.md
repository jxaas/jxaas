# JXaaS routing for scale-out

JXaaS has built-in scale-out support; a front-end server can direct traffic
to one of a cluster of back-end JXaaS servers.

## Install Juju and JXaaS

First we install Juju & JXaaS (locally):

```
juju init
juju switch local
juju bootstrap
juju status

juju deploy cs:~justin-fathomdb/trusty/jxaas jxaas

API_SECRET=`grep admin-secret ~/.juju/environments/local.jenv | cut -f 2 -d ':' | tr -d ' '`
echo "API_SECRET=${API_SECRET}"
juju set jxaas api-password=${API_SECRET}

juju expose jxaas
```


## Install the JXaaS router

The JXaaS router uses etcd to store the routing configuration.

Install etcd:
```
curl -L  https://github.com/coreos/etcd/releases/download/v0.4.6/etcd-v0.4.6-linux-amd64.tar.gz -o /tmp/etcd.tar.gz
mkdir -p ~/etcd
cd ~/etcd
tar -x -z -v --strip-components 1 -f /tmp/etcd.tar.gz
~/etcd/etcd &
```

Assuming you've installed JXaaS from source (so you have it on your machine):
```
git clone https://github.com/jxaas/jxaas.git ~/jxaas

cd ~/jxaas
./build.sh
```

We can now start the JXaaS front-end router:
```
.build/bin/jxaas-routerd &
```

Now we will direct traffic for the mysql service to the JXaaS server running in Juju:
```
PUBLIC_ADDRESS=`juju status jxaas | grep public-address | cut -f 2 -d ':' | tr -d ' '`
echo "JXaaS back-end is listening at http://${PUBLIC_ADDRESS}:8080"

.build/bin/jxaas-admin set-service-backend mysql ${PUBLIC_ADDRESS}:8080
```

Now we can see the router configuration:
```
.build/bin/jxaas-admin list-service-backends
```

So now we can use the JXaaS mysql service, the router will route requests to the JXaaS server running in Juju:
```
jxaas list-instances mysql
jxaas create-instance mysql m1
jxaas wait mysql m1
jxaas connect mysql m1
```

You will see log messages with the JXaaS router proxying to the back-end server.


# Summary

For large-scale deployments, JXaaS has a front-end router than can split the load across
several back-end JXaaS servers and even different Juju clusters.
