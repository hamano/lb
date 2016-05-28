package main

import (
	"fmt"
	"log"
	"reflect"
	"github.com/urfave/cli"
)

type ModifyJob struct {
	BaseJob
	attr string
	value string
}

var modifyFlags = []cli.Flag {
	cli.StringFlag {
		Name: "attr",
		Value: "sn",
		Usage: "attribute",
	},
	cli.StringFlag {
		Name: "value",
		Value: "modified",
		Usage: "attribute value for modify",
	},
}

func Modify(c *cli.Context) {
	runBenchmark(c, reflect.TypeOf(ModifyJob{}))
}

func (job *ModifyJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	err := job.ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
		return false
	}
	job.attr = c.String("attr")
	job.value = c.String("value")
	return true
}

func (job *ModifyJob) Request() bool {
	dn := fmt.Sprintf("cn=%d-%d,%s", job.wid, job.count, job.baseDN)
	attrs := map[string][]string{job.attr:[]string{job.value}}
	err := job.ldap.Modify(dn, attrs)
	if err != nil {
		log.Printf("modify error: %s", err)
		return false
	}
	return true
}
