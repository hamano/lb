lb - LDAP benchmarking tool like an Apache Bench
================================================

lb is simple benchmarking tool for LDAP Server.
This tool is designed to allow perform by command line such as Apache Bench.

## Benchmarking scenario

*There are some specification changes from version 2.0.0 onwards. Please check the usage.*

ADD benchmarking adds entries sequentially with `cn=0, cn=1, cn=2, ...`, distributed across concurrent tasks.
Each task handles a specific range of entries to ensure exact total request counts.

For example, with `-n 10 -c 2`:
- Task 0 adds: cn=0, cn=1, cn=2, cn=3, cn=4
- Task 1 adds: cn=5, cn=6, cn=7, cn=8, cn=9

DELETE benchmarking follows the same pattern. Therefore:
- ADD benchmarking
- Various operation benchmarking
- DELETE benchmarking

The above scenario cleans up the entry database, allowing repeated execution.

## Build

~~~
$ cargo build -r
~~~

## Usage

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

This command adds 1000 entries with 10 concurrent tasks.
Each task handles a specific range of sequential entries.

For example, with `-c 2 -n 10`:
- Task 0 (tid=0) adds: cn=0, cn=1, cn=2, cn=3, cn=4
- Task 1 (tid=1) adds: cn=5, cn=6, cn=7, cn=8, cn=9

Entry format:
~~~
dn: cn=${INDEX},dc=example,dc=com
cn: ${INDEX}
sn: ${TASK_ID}
objectClass: person
userPassword: password
~~~

Use --uuid option if you want to use UUID for cn.
~~~
$ lb add -c 10 -n 1000 --uuid ldap://localhost/
~~~

~~~
dn: cn=${UUID},dc=example,dc=com
cn: ${UUID}
sn: ${TASK_ID}
objectClass: person
userPassword: password
~~~

### DELETE Benchmarking

~~~
$ lb delete -c 10 -n 1000 ldap://localhost/
~~~

This command makes 1000 delete requests with 10 concurrent tasks.
Each task handles a specific range of sequential entries.

For example, with `-c 2 -n 10`:
- Task 0 deletes: cn=0, cn=1, cn=2, cn=3, cn=4
- Task 1 deletes: cn=5, cn=6, cn=7, cn=8, cn=9

DNs being deleted:
~~~
cn=0,dc=example,dc=com
cn=1,dc=example,dc=com
...
cn=999,dc=example,dc=com
~~~

### BIND Benchmarking

* BIND Benchmarking with single entry

~~~
$ lb bind -c 10 -n 1000 ldap://localhost/
~~~
This command makes 1000 bind requests with 10 concurrent tasks (100 requests per task).

* BIND Benchmarking with random entries

Each task randomly selects entries from cn=0 to cn=(n-1) for binding.

### SEARCH Benchmarking

* Search Benchmarking with random filters

~~~
$ lb search -c 10 -n 1000 -s sub ldap://localhost/
~~~

This command makes 1000 search requests with 10 concurrent tasks.
Each task randomly selects entries from cn=0 to cn=(n-1) for search filters.

For example, with `-n 1000`, searches are performed with random filters:
~~~
(cn=0)
(cn=125)
(cn=573)
...
(cn=999)
~~~

### MODIFY Benchmarking

~~~
$ lb modify -c 10 -n 1000 --attr sn --value modified ldap://localhost/
~~~

This command makes 1000 modify requests with 10 concurrent tasks.
Each task handles a specific range of sequential entries.

For example, with `-c 2 -n 10`:
- Task 0 modifies: cn=0, cn=1, cn=2, cn=3, cn=4
- Task 1 modifies: cn=5, cn=6, cn=7, cn=8, cn=9

DNs being modified:
~~~
cn=0,dc=example,dc=com
cn=1,dc=example,dc=com
...
cn=999,dc=example,dc=com
~~~

## TODO
* modrdn benchmarking
