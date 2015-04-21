lb - LDAP benchmarking tool like an Apache Bench
================================================

lb is simple benchmarking tool for LDAP Server.
It is designed to allow perform by command line as with Apache Bench.

## Usage

~~
% lb bind -n 1000 -c 10 -D cn=test,dc=example,dc=com -w secret ldap://localhost:389/
~~

It will make 1000 bind request with 10 threads.

