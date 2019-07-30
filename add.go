package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/urfave/cli"
	"gopkg.in/ldap.v3"
	"log"
	"reflect"
)

type AddJob struct {
	BaseJob
	uuid bool
}

var addFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "uuid",
		Usage: "use UUID",
	},
}

func Add(c *cli.Context) error {
	runBenchmark(c, reflect.TypeOf(AddJob{}))
	return nil
}

func (job *AddJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	job.uuid = c.Bool("uuid")
	err := job.conn.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
		return false
	}
	return true
}

func (job *AddJob) Request() bool {
	var cn string
	id, err := uuid.NewV1()
	if job.uuid {
		cn = id.String()
	} else {
		cn = fmt.Sprintf("%d-%d", job.wid, job.count)
	}
	dn := fmt.Sprintf("cn=%s,%s", cn, job.baseDN)
	attrs := []ldap.Attribute{
		ldap.Attribute{"objectClass", []string{"person"}},
		ldap.Attribute{"cn", []string{cn}},
		ldap.Attribute{"sn", []string{"sn"}},
		ldap.Attribute{"userPassword", []string{"secret"}},
	}
	req := ldap.AddRequest{
		DN:         dn,
		Attributes: attrs,
	}
	err = job.conn.Add(&req)
	if err != nil {
		log.Printf("add error: %s", err)
		return false
	}
	return true
}
