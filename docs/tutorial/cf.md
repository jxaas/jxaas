# CloudFoundry integration

JXaaS can provide services directly to CloudFoundry, acting as a service broker.

## Install CloudFoundry

There are two ways to install CloudFoundry - either using a Juju charm, or via the BOSH installer.

The Juju charm is much easier!

### Installing: The Juju way

Make sure you have Juju running, and it is recommended to use AWS for your cloud provider.  Then:

```
mkdir -p ~/cf/trusty
cd ~/cf/
bzr branch lp:~cf-charmers/charms/trusty/cloudfoundry/trunk trusty/cloudfoundry
cd trusty/cloudfoundry
./cfdeploy <cf_admin_password>
```

### Installing: The BOSH way

This technique uses the official CloudFoundry installation process, which is a pretty slow process.
Everything runs inside of VirtualBox.  This downloads a lot of data, needs a lot of disk space, and
takes a long time.  We're also getting a non-production environment.

The one advantage is that this method seems to work better if you want to run everything locally, instead
of using AWS (i.e. if you want a free configuration).

First install virtualbox and vagrant:

```
which virtualbox || sudo apt-get install --yes virtualbox
wget -O /tmp/vagrant.deb https://dl.bintray.com/mitchellh/vagrant/vagrant_1.7.2_x86_64.deb
sudo dpkg -i /tmp/vagrant.deb
```

Now install the bosh command line tool:

```
which gem || sudo apt-get install --yes ruby
sudo apt-get install --yes ruby-dev
which bosh || sudo gem install bosh_cli
```

Now install bosh-lite:

```
mkdir ~/cf
cd ~/cf
git clone https://github.com/cloudfoundry/bosh-lite.git
cd ~/cf/bosh-lite
vagrant up --provider=virtualbox
bosh target 192.168.50.4 lite
bosh login admin admin
sudo bin/add-route
```

Download spiff:

```
cd ~/cf
wget -O /tmp/spiff.zip https://github.com/cloudfoundry-incubator/spiff/releases/download/v1.0.3/spiff_linux_amd64.zip
mkdir -p ~/cf/spiff
cd ~/cf/spiff
unzip /tmp/spiff.zip
```

Check out the cloudfoundry code:

```
cd ~/cf
git clone https://github.com/cloudfoundry/cf-release
```

Edit the CF configuration to allow access to 10.0.3.x, using `vim ~/cf/cf-release/templates/cf-properties.yml`

Add these lines to default_security_group_definitions
(and this is YAML, so be careful about spaces, and don't use tabs).

```
    - protocol: all
      destination: 10.0.3.0-10.0.3.255
```

Now install CloudFoundry (this step takes a _long_ time):

```
cd ~/cf/bosh-lite
export PATH=~/cf/spiff:$PATH
bin/provision_cf
```

Install the CloudFoundry CLI:

```
wget -O /tmp/cf.deb "https://cli.run.pivotal.io/stable?release=debian64&source=github"
sudo dpkg -i /tmp/cf.deb
rm /tmp/cf.deb
```

Configure the command line tools:

```
cf api --skip-ssl-validation https://api.10.244.0.34.xip.io
cf auth admin admin
```

## Use CloudFoundry

Finally, you can use CloudFoundry!

```
cf create-org me
cf target -o me
cf create-space development
cf target -s development

```

Let's run an example app; spring music is a simple CRUD webapp that shows some music albums.  By default it runs
with an in-memory database.
 
```
mkdir -p ~/cf/apps
cd ~/cf/apps
git clone https://github.com/cloudfoundry-samples/spring-music.git

cd spring-music
./gradlew assemble
cf push spring-music -n spring-music

APP_HOST=`cf app spring-music | grep urls | cut -f 2 -d ' '`
x-www-browser http://${APP_HOST}
```

You can see that the app is currently bound to an in-memory data store:

```
# Should include 'in-memory', not 'mysql'
APP_HOST=`cf app spring-music | grep urls | cut -f 2 -d ' '`
curl http://${APP_HOST}/info | grep memory
```

## Use CloudFoundry with JXaaS

Let's have it persist to a MySQL database instead.  And let's create that MySQL database using Juju and JXaas.

First we install JXaaS:

```
juju deploy cs:~justin-fathomdb/trusty/jxaas jxaas

API_SECRET=`grep admin-secret ~/.juju/environments/local.jenv | cut -f 2 -d ':' | tr -d ' '`
echo "API_SECRET=${API_SECRET}"
juju set jxaas api-password=${API_SECRET}

juju expose jxaas

PUBLIC_ADDRESS=`juju status jxaas | grep public-address | cut -f 2 -d ':' | tr -d ' '`
echo "JXaaS is listening at http://${PUBLIC_ADDRESS}:8080"
```


Now, we tell CloudFoundry about JXaaS.  In the CloudFoundry vocabulary, JXaaS is called a service broker, because it
provides services to CloudFoundry.  First we add the service-broker, by specifying the '/cf/mysql' URL for JXaaS:

```
export JXAAS_URL_CF=http://${PUBLIC_ADDRESS}:8080/cf/mysql
echo "JXAAS_URL_CF is ${JXAAS_URL_CF}"

cf service-brokers
cf create-service-broker jxaas-mysql admin admin ${JXAAS_URL_CF}
cf service-brokers
```

Next, we tell CloudFoundry to allow users to access the MySQL service:

```
cf service-access
cf enable-service-access mysql
cf service-access
```

Let's create a mysql service named "mysql1".  This step is a little slower, because CF operations are
generally synchronous:

```
cf create-service mysql nano mysql1
cf services
```

What happened here was that CloudFoundry asked JXaaS to create the mysql service.  JXaaS implements the CloudFoundry
service-broker interface, and the manifest specifies several plans which can be chosen - here we chose the nano
plan.  Because we created it through CloudFoundry, that means CloudFoundry knows about the MySQL service and 
we can easily bind our app to it:

```
cf bind-service spring-music mysql1
cf restage spring-music
```

And now the database will be mysql:

```
# Should include 'mysql', not 'in-memory':
APP_HOST=`cf app spring-music | grep urls | cut -f 2 -d ' '`
curl http://${APP_HOST}/info | grep mysql
```

You can see the configuration:

```
cf env spring-music
```

You can see that the app still works:
```
x-www-browser http://${APP_HOST}
```


# Summary

JXaaS can act as a CloudFoundry service-broker, so Juju charms can easily be bound
to CloudFoundry applications.  This opens up the whole world of Juju services to CloudFoundry!