## Openstack (Keystone) authentication

### Set up Keystone

If you don't have an existing Keystone installation, you'll need to install one.


```
ADMIN_PASSWORD=secret
KEYSTONE_HOST=127.0.0.1

export OS_SERVICE_ENDPOINT=http://127.0.0.1:35357/v2.0
export OS_SERVICE_TOKEN=ADMIN

sudo apt-get install --yes keystone
keystone user-list

# Create initial accounts
keystone role-create --name admin 
keystone user-create --name=admin --pass=${ADMIN_PASSWORD}
keystone tenant-create --name admin --description "Admin Tenant" --enabled true
keystone user-role-add --user=admin --tenant=admin --role=admin
keystone user-role-add --user=admin --tenant=admin --role=_member_ 

keystone tenant-create --name=service --description="Service Tenant"

keystone service-create --name=keystone --type=identity --description="OpenStack Identity"
keystone endpoint-create \
  --service-id=$(keystone service-list | awk '/ identity / {print $2}') \
  --publicurl=http://${KEYSTONE_HOST}:5000/v2.0 \
  --internalurl=http://${KEYSTONE_HOST}:5000/v2.0 \
  --adminurl=http://${KEYSTONE_HOST}:35357/v2.0

```

### Use OpenStack to authenticate

Let's install Juju & JXaaS (locally):

```
juju init
juju switch local
juju bootstrap
juju status

juju deploy cs:~justin-fathomdb/trusty/jxaas jxaas

API_SECRET=`grep admin-secret ~/.juju/environments/local.jenv | cut -f 2 -d ':' | tr -d ' '`
echo "API_SECRET=${API_SECRET}"
juju set jxaas api-password=${API_SECRET}

juju set jxaas openstack-auth="http://10.0.3.1:5000/v2.0"

juju expose jxaas

PUBLIC_ADDRESS=`juju status jxaas | grep public-address | cut -f 2 -d ':' | tr -d ' '`
echo "JXaaS is listening at http://${PUBLIC_ADDRESS}:8080"

export JXAAS_URL=http://${PUBLIC_ADDRESS}:8080/xaas
echo "export JXAAS_URL=http://${PUBLIC_ADDRESS}:8080/xaas"
```


### Create a service account in OpenStack for jxaas

Now, we register the JXaaS server with OpenStack Keystone.  This allows
users to authenticate with KeyStone and discover services available to them,
including JXaaS.

```
IP=10.0.3.1

JXAAS_PASSWORD=secret
JXAAS_BASE=http://${PUBLIC_ADDRESS}:8080/xaas

keystone user-create --name=jxaas --pass=${JXAAS_PASSWORD}
keystone user-role-add --user=jxaas --tenant=service --role=admin

keystone service-create --name=jxaas --type=jxaas \
  --description="Juju XaaS"
keystone endpoint-create \
  --service-id=$(keystone service-list | awk '/ jxaas / {print $2}') \
  --publicurl=${JXAAS_BASE}/%\(tenant_id\)s \
  --internalurl=${JXAAS_BASE}/%\(tenant_id\)s \
  --adminurl=${JXAAS_BASE}/%\(tenant_id\)s
```


### Use the JXaaS CLI

To use JXaaS with Openstack authentication, you point it to the Keystone server.

After authenticating, the CLI uses the JXaaS server you registered in Keystone.

```
export JXAAS_AUTH=openstack
export JXAAS_URL=http://10.0.3.1:5000/v2.0

jxaas list-instances mysql

jxaas create-instance mysql m1

jxaas list-instances mysql

```

JXaaS now creates Juju instances based on your OpenStack project:

```
juju status *mysql-m1-mysql
```

The juju service id will include your OpenStack project ID.