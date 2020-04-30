package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/ldap.v3"
	"log"
	"reflect"
)

type ModifyJob struct {
	BaseJob
	attr  string
	value string
}

var modifyFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "attr",
		Value: "sn",
		Usage: "attribute",
	},
	&cli.StringFlag{
		Name:  "value",
		Value: "modified",
		Usage: "attribute value for modify",
	},
}

func Modify(c *cli.Context) error {
	runBenchmark(c, reflect.TypeOf(ModifyJob{}))
	return nil
}

func (job *ModifyJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	err := job.conn.Bind(c.String("D"), c.String("w"))
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
	mod := ldap.PartialAttribute{job.attr, []string{job.value}}
	change := ldap.Change{
		Operation:    ldap.ReplaceAttribute,
		Modification: mod,
	}
	req := ldap.ModifyRequest{
		DN:      dn,
		Changes: []ldap.Change{change},
	}
	err := job.conn.Modify(&req)
	if err != nil {
		log.Printf("modify error: %s", err)
		return false
	}
	return true
}
