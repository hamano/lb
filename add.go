package main

import (
	"fmt"
	"log"
	"reflect"
	"github.com/codegangsta/cli"
	"github.com/satori/go.uuid"
)

type AddJob struct {
	BaseJob
}

func (job *AddJob) Request() bool {
	cn := uuid.NewV1().String()
	dn := fmt.Sprintf("cn=%s,dc=example,dc=com", cn)
	attrs := map[string][]string{
		"objectClass": {"person"},
		"cn": {cn},
		"sn": {"sn"},
		"userPassword": {"secret"},
	}
	err := job.ldap.Add(dn, attrs)
	if err != nil {
		log.Printf("add error: %s", err)
		return false
	}
	return true
}

func add(c *cli.Context) {
	runBenchmark(c, reflect.TypeOf(AddJob{}))
}
