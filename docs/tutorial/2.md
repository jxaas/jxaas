## 2 - Proxy Charms

Go to EC2, and open up port 8080 from your IP only. (JXaaS isn't yet secure,
so we need the IP restriction!)

In step #1, we forwarded the API port.  So, if you have the Python client
installed, you can run JXaaS commands locally.

```
export JXAAS_URL=http://54.204.195.171:8080/xaas
jxaas list-instances mysql
```


Now, let's demonstrate the proxy charm.  This allows you to consume JXaaS services easily
from within Juju.
   
```
export JXAAS_URL=http://54.204.195.171:8080/xaas
cat > /tmp/config <<EOF
mp1:
  jxaas-url: ${JXAAS_URL}
  jxaas-tenant: tenant1
  jxaas-user: user1
  jxaas-secret: secret1
EOF

juju deploy --config=/tmp/config cs:~justin-fathomdb/trusty/mysql-proxy mp1

juju deploy cs:~justin-fathomdb/trusty/mediawiki wiki1

juju status wiki1
juju status mp1

juju add-relation wiki1:db mp1:db
```

If you want to watch progress:

```
sudo tail -f  /var/log/juju-*-local/all-machines.log 
```