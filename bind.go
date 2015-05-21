package main

import (
	"log"
	"time"
	"reflect"
	"github.com/codegangsta/cli"
)

type BindJob struct {
	BaseJob
}

func (job *BindJob) Init(wid int, c *cli.Context) bool {
	log.Printf("bind init:\n")
	return true
}

func (job *BindJob) Request() bool {
	log.Printf("bind request: %d\n", job.count)
	time.Sleep(1000 * time.Millisecond)
	return true
}

func Bind(c *cli.Context) {
	runBenchmark(c, reflect.TypeOf(BindJob{}))
}
