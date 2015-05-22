lb - LDAP benchmarking tool like an Apache Bench
================================================

lb is simple benchmarking tool for LDAP Server.
It is designed to allow perform by command line as with Apache Bench.

# Build

## Install dependency

* for Debian or Ubuntu
~~~
# apt-get install build-essential libldap2-dev
~~~

## Install lb
~~~
% go get github.com/hamano/lb
~~~

# Usage

## Setup subcommand

* Add base entry
~~~
% lb setup base -b 'dc=example,dc=com' ldap://localhost/
~~~

* Add single entry
~~~
% lb setup person --dn 'cn=test,dc=example,dc=com' ldap://localhost/
~~~

* Add range entry
~~~
% lb setup person --dn 'cn=test%d,dc=example,dc=com' --last 10 ldap://localhost/
~~~

## BIND Benchmarking

* BIND Benchmarking to single entry

~~~
% lb bind -n 1000 -c 10 -D cn=test,dc=example,dc=com -w secret ldap://localhost:389/
~~~

It will make 1000 bind request with 10 threads.

* BIND Benchmarking to random entry
~~~
% lb bind -D 'cn=test%d,dc=example,dc=com' -w secret --last 10 ldap://localhost/
~~~

## ADD Benchmarking

~~~
% lb add -n 1000 -c 10 ldap://localhost/
~~~
