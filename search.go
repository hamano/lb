package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/ldap.v3"
	"log"
	"math/rand"
	"reflect"
	"strings"
)

type SearchJob struct {
	BaseJob
	attributes  []string
	baseDN  string
	scope   int
	filter  string
	first   int
	last    int
	idRange int
}

var searchFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "s",
		Value: "sub",
		Usage: "scope",
	},
	&cli.StringFlag{
		Name:  "a, filter",
		Value: "(objectClass=*)",
		Usage: "filter",
	},
	&cli.StringFlag{
		Name:  "attributes",
		Value: "dn",
		Usage: "space-separated list of attributes as a single string",
	},
	&cli.IntFlag{
		Name:  "first",
		Value: 1,
		Usage: "first id",
	},
	&cli.IntFlag{
		Name:  "last",
		Value: 0,
		Usage: "last id",
	},
}

func Search(c *cli.Context) error {
	runBenchmark(c, reflect.TypeOf(SearchJob{}))
	return nil
}

func (job *SearchJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	err := job.conn.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
		return false
	}
	job.baseDN = c.String("b")
	job.filter = c.String("a")
	job.attributes = strings.Fields(c.String("t"))
	job.first = c.Int("first")
	job.last = c.Int("last")
	switch c.String("s") {
	case "base":
		job.scope = ldap.ScopeBaseObject
	case "one":
		job.scope = ldap.ScopeSingleLevel
	case "sub":
		job.scope = ldap.ScopeWholeSubtree
	case "children":
		job.scope = 3
	default:
		job.scope = ldap.ScopeWholeSubtree
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
	req := ldap.SearchRequest{
		BaseDN:       job.baseDN,
		Scope:        job.scope,
		DerefAliases: 0,
		Filter:       filter,
		Attributes:   job.attributes,
	}
	res, err := job.conn.Search(&req)
	if err != nil {
		return false
	}
	if len(res.Entries) == 0 {
		return false
	}
	return true
}
