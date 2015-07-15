lb - LDAP benchmarking tool like an Apache Bench
================================================

lb is simple benchmarking tool for LDAP Server.
This tool is designed to allow perform by command line such as Apache Bench.

## Build

### Install dependencies

* for Debian or Ubuntu
~~~
# apt-get install build-essential golang libldap2-dev
~~~

* Setting GOPATH
~~~
$ export GOPATH=~/go
$ export PATH=$GOPATH/bin:$PATH
~~~

### Install lb command
~~~
$ go get github.com/hamano/lb
~~~

## Usage

lb have setup sub-command that preparing for benchmark.

### Setup subcommand

* Add base entry
~~~
$ lb setup base -b 'dc=example,dc=com' ldap://localhost/
~~~
This command add base entry.

* Add single entry
~~~
$ lb setup person --cn 'test' ldap://localhost/
~~~

* Add range entries
~~~
$ lb setup person --cn 'user%d' --last 10 ldap://localhost/
~~~

### ADD Benchmarking

~~~
$ lb add -n 1000 -c 10 ldap://localhost/
~~~

This command add following entries 1000 times with 10 threads.

~~~
dn: cn=${THREADID}-${COUNT},dc=example,dc=com
cn: ${THREADID}-${COUNT}
sn: sn
userPassword: secret
~~~

Use --uuid option if you want to use UUID for cn.
~~~
$ lb add -n 1000 -c 10 --uuid ldap://localhost/
~~~

~~~
dn: cn=${UUID},dc=example,dc=com
cn: ${UUID}
sn: sn
userPassword: secret
~~~

### DELETE Benchmarking

~~~
$ lb delete -n 1000 -c 10 ldap://localhost/
~~~

This command delete following DNs:

~~~
cn=0-0,dc=example,dc=com
...
cn=9-999,dc=example,dc=com
~~~

### BIND Benchmarking

* BIND Benchmarking with single entry

~~~
$ lb bind -n 1000 -c 10 -D cn=user,dc=example,dc=com -w secret ldap://localhost/
~~~
This command make 1000 times bind request with 10 threads.

* BIND Benchmarking with ranged random entries
~~~
$ lb bind -D 'cn=user%d,dc=example,dc=com' -w secret --last 10 ldap://localhost/
~~~

### SEARCH Benchmarking

* Search Benchmarking with random filters
~~~
$ lb search -n 1000 -c 10 -a "(cn=user%d)" --last 1000 -s sub ldap://localhost/
~~~
This command make 1000 times search request with following random filters:

~~~
(cn=user1)
...
(cn=user1000)
~~~

## TODO
* DELETE benchmarking
