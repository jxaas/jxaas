This has to be run against EC2, because CloudFoundry/BOSH blocks access to the 10., 192.168. and 172.16. ranges.


```

PUBLIC_ADDRESS=`juju status jxaas | grep public-address | cut -f 2 -d ':' | tr -d ' '`
echo "PUBLIC_ADDRESS is ${PUBLIC_ADDRESS}"
export JXAAS_CF_URL=http://${PUBLIC_ADDRESS}:8080/cf
echo "export JXAAS_CF_URL=http://${PUBLIC_ADDRESS}:8080/cf"

cf service-brokers

cf create-service-broker jxaas jxaasuser jxaaspassword ${JXAAS_CF_URL}
cf update-service-broker jxaas jxaasuser jxaaspassword ${JXAAS_CF_URL}

cf service-brokers
```


```
cf service-access
cf enable-service-access mysql
cf service-access
```

```
cf create-service mysql default mysql1
cf services

cf apps

cf bind-service myapp mysql1

cf env myapp
```


cf unbind-service myapp mysql2
