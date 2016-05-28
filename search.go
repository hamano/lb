package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"math/rand"
	"github.com/urfave/cli"
	openldap "github.com/hamano/golang-openldap"
)

type SearchJob struct {
	BaseJob
	baseDN string
	scope int
	filter string
	first int
	last int
	idRange int
}

var searchFlags = []cli.Flag {
	cli.StringFlag {
		Name: "s",
		Value: "sub",
		Usage: "scope",
	},
	cli.StringFlag {
		Name: "a, filter",
		Value: "(objectClass=*)",
		Usage: "filter",
	},
	cli.IntFlag {
		Name: "first",
		Value: 1,
		Usage: "first id",
	},
	cli.IntFlag {
		Name: "last",
		Value: 0,
		Usage: "last id",
	},
}

func Search(c *cli.Context) {
	runBenchmark(c, reflect.TypeOf(SearchJob{}))
}

func (job *SearchJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	err := job.ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
		return false
	}
	job.baseDN = c.String("b")
	job.filter = c.String("a")
	job.first = c.Int("first")
	job.last = c.Int("last")
	switch c.String("s") {
	case "base":
		job.scope = openldap.LDAP_SCOPE_BASE
	case "one":
		job.scope = openldap.LDAP_SCOPE_ONE
	case "sub":
		job.scope = openldap.LDAP_SCOPE_SUBTREE
	case "children":
		job.scope = openldap.LDAP_SCOPE_CHILDREN
	default:
		job.scope = openldap.LDAP_SCOPE_SUBTREE
	}

	if strings.Contains(job.filter, "%") {
		job.idRange = job.last - job.first + 1
	}
	return true
}

func (job *SearchJob) Request() bool {
	var filter string
	if job.idRange > 0 {
		id := rand.Intn(job.idRange) + job.first
		filter = fmt.Sprintf(job.filter, id)
	} else {
		filter = job.filter
	}
	res, err := job.ldap.SearchAll(job.baseDN, job.scope, filter, []string{"dn"})
	if err != nil {
		return false
	}
	if res.Count() == 0 {
		return false
	}
	return true
}
