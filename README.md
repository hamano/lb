lb - LDAP benchmarking tool like an Apache Bench
================================================

lb is simple benchmarking tool for LDAP Server.
This tool is designed to allow perform by command line such as Apache Bench.

## Benchmarking scenario

*There are some specification changes from version 2.0.0 onwards. Please check the usage.*

ADD benchmarking adds entries in the format `cn=${THREADID}-${COUNT}`.
DELETE benchmarking is similar. Therefore,
- ADD benchmarking
- Various operation benchmarking
- DELETE benchmarking

The above scenario cleans up the entry database, allowing repeated execution.

## Build

~~~
$ cargo build -r
~~~

## Usage

lb has base/person subcommands for preparing benchmarking entries.

### Base subcommand

* Add base entry
~~~
$ lb base -b 'dc=example,dc=com' ldap://localhost/
~~~
This command add base entry.

### ADD Benchmarking

~~~
$ lb add -c 10 -n 1000 ldap://localhost/
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
$ lb add -c 10 -n 1000 --uuid ldap://localhost/
~~~

~~~
dn: cn=${UUID},dc=example,dc=com
cn: ${UUID}
sn: sn
userPassword: secret
~~~

### DELETE Benchmarking

~~~
$ lb delete -c 10 -n 1000 ldap://localhost/
~~~

This command make delete request with following DNs:

~~~
cn=0-0,dc=example,dc=com
...
cn=9-999,dc=example,dc=com
~~~

### BIND Benchmarking

* BIND Benchmarking with single entry

~~~
$ lb bind -c 10 -n 1000 -D cn=user,dc=example,dc=com -w secret ldap://localhost/
~~~
This command make 1000 times bind request with 10 threads.

* BIND Benchmarking with ranged random entries
~~~
$ lb bind -D 'cn=user%d,dc=example,dc=com' -w secret --last 10 ldap://localhost/
~~~

### SEARCH Benchmarking

* Search Benchmarking with random filters
~~~
$ lb search -c 10 -n 1000 -a "(cn=user%d)" --last 1000 -s sub ldap://localhost/
~~~
This command make 1000 times search request with following random filters:

~~~
(cn=user1)
...
(cn=user1000)
~~~

### MODIFY Benchmarking

~~~
$ lb modify -c 10 -n 1000 --attr sn --value modified ldap://localhost/
~~~

This command make modify request with following DNs:

~~~
cn=0-0,dc=example,dc=com
...
cn=9-999,dc=example,dc=com
~~~

## TODO
* modrdn benchmarking
