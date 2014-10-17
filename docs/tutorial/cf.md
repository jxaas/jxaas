```
cf service-brokers

cf create-service-broker jxaas jxaasuser jxaaspassword http://172.16.2.4:8080/cf
cf update-service-broker jxaas jxaasuser jxaaspassword http://172.16.2.4:8080/cf

cf service-brokers
```


```
cf service-access
cf enable-service-access mysql
cf service-access
```

```
cf create-service mysql default mysql2
cf services

cf apps

cf bind-service myapp mysql2

cf env myapp
