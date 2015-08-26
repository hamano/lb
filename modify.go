package main

import (
	"fmt"
	"log"
	"reflect"
	"github.com/codegangsta/cli"
)

type ModifyJob struct {
	BaseJob
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
	return true
}

func (job *ModifyJob) Request() bool {
	dn := fmt.Sprintf("cn=%d-%d,%s", job.wid, job.count, job.baseDN)
	attrs := map[string][]string{"sn":[]string{"modified"}}
	err := job.ldap.Modify(dn, attrs)
	if err != nil {
		log.Printf("modify error: %s", err)
		return false
	}
	return true
}
