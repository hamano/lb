package main

import (
	"log"
	"fmt"
	"strings"
	"reflect"
	"math/rand"
	"github.com/codegangsta/cli"
)

type BindJob struct {
	BaseJob
	dn string
	password string
	first int
	last int
	idRange int
}

var bindFlags = []cli.Flag {
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

func Bind(c *cli.Context) {
	runBenchmark(c, reflect.TypeOf(BindJob{}))
}

func (job *BindJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 2 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	job.dn = c.String("D")
	job.password = c.String("w")
	job.first = c.Int("first")
	job.last = c.Int("last")

	if strings.Contains(job.dn, "%") {
		job.idRange = job.last - job.first + 1
	}
	return true
}

func (job *BindJob) Request() bool {
	var dn string
	if job.idRange > 0 {
		id := rand.Intn(job.idRange) + job.first
		dn = fmt.Sprintf(job.dn, id)
	} else {
		dn = job.dn
	}
	err := job.ldap.Bind(dn, job.password)
	if err != nil {
		return false
	}
	return true
}
