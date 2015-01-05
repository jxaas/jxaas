Install bosh lite following the instructions here: https://github.com/cloudfoundry/bosh-lite

```
mkdir ~/cf
cd ~/cf
git clone https://github.com/cloudfoundry/bosh-lite.git
cd ~/cf/bosh-lite
VBoxManage --version
vagrant up --provider=virtualbox
bosh target 192.168.50.4 lite
bosh login admin admin
bin/add-route
# May need to enter sudo password
```

Edit the CF configuration to allow access to 1.0.3.x:

```vim ../cf-release/templates/cf-properties.yml```

Add these lines to default_security_group_definitions
(and this is YAML, so be careful about spaces, and don't use tabs).

```
    - protocol: all
      destination: 10.0.3.0-10.0.3.255
```

Now install CloudFoundry (this step takes a while):

```
bin/provision_cf

#sudo ip route add 10.0.2.0/24 via 192.168.50.4 dev vboxnet0

cf api --skip-ssl-validation https://api.10.244.0.34.xip.io
cf auth admin admin
cf create-org me
cf target -o me
cf create-space development
cf target -s development

```

If you want to run an example app:

```

mkdir -p ~/cf/apps
cd ~/cf/apps
git clone https://github.com/cloudfoundry-samples/spring-music.git

cd spring-music
./gradlew assemble
cf push

cf service-brokers
cf create-service-broker jxaas admin admin http://10.0.3.1:8080/cf
cf enable-service-access mysql
cf service-access

cf create-service mysql default mysql1
cf services
cf bind-service spring-music mysql1

cf restage spring-music

cf env myapp
```


