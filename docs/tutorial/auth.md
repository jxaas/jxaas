## Openstack (Keystone) authentication

### Set up Keystone

If you don't have an existing Keystone installation, you'll need to install one.



```
ADMIN_PASSWORD=secret
KEYSTONE_HOST=127.0.0.1

export OS_SERVICE_ENDPOINT=http://127.0.0.1:35357/v2.0
export OS_SERVICE_TOKEN=ADMIN

apt-get install keystone
keystone user-list

# Undifferentiated heavy lifting
#  (and the primary reason why Openstack is hard to install!)
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
 
 
### Create a service account for jxaas

```
JXAAS_PASSWORD=secret
JXAAS_BASE=http://127.0.0.1:8080/xaas

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


