package main

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/urfave/cli"
	"log"
	"reflect"
)

type DeleteJob struct {
	BaseJob
}

func Delete(c *cli.Context) error {
	runBenchmark(c, reflect.TypeOf(DeleteJob{}))
	return nil
}

func (job *DeleteJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 1 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	err := job.conn.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
		return false
	}
	return true
}

func (job *DeleteJob) Request() bool {
	dn := fmt.Sprintf("cn=%d-%d,%s", job.wid, job.count, job.baseDN)
	req := ldap.DelRequest{
		DN: dn,
	}
	err := job.conn.Del(&req)
	if err != nil {
		log.Printf("delete error: %s", err)
		return false
	}
	return true
}
