package main

import (
	"crypto/tls"
	"github.com/urfave/cli"
	"gopkg.in/ldap.v3"
	"log"
	"time"
)

type Result struct {
	wid         int
	count       int
	success     int
	startTime   time.Time
	endTime     time.Time
	elapsedTime float64
}

type Job interface {
	Init(int, *cli.Context) bool
	Prep(*cli.Context) bool
	Finish()
	Request() bool
	GetVerbose() int
	IncCount()
	GetCount() int
	IncSuccess()
	GetSuccess() int
}

type BaseJob struct {
	conn    *ldap.Conn
	baseDN  string
	wid     int
	count   int
	success int
	verbose int
}

func (job *BaseJob) Request() bool {
	if job.verbose >= 3 {
		log.Printf("[%d]: %d\n", job.wid, job.count)
	}
	time.Sleep(1000 * time.Millisecond)
	return true
}

func (job *BaseJob) GetVerbose() int {
	return job.verbose
}

func (job *BaseJob) IncCount() {
	job.count++
}

func (job *BaseJob) GetCount() int {
	return job.count
}

func (job *BaseJob) IncSuccess() {
	job.success++
}

func (job *BaseJob) GetSuccess() int {
	return job.success
}

func (job *BaseJob) Init(wid int, c *cli.Context) bool {
	job.wid = wid
	job.verbose = c.Int("verbose")
	job.baseDN = c.String("b")
	url := c.Args().First()
	if job.verbose >= 2 {
		log.Printf("worker[%d]: initialize %s\n", job.wid, url)
	}
	var err error
	job.conn, err = ldap.DialURL(url)
	if err != nil {
		log.Fatal("initialize error: ", err)
		return false
	}

	if c.Bool("Z") {
		job.conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
	}
	return true
}

func (job *BaseJob) Prep(c *cli.Context) bool {
	if job.GetVerbose() >= 2 {
		log.Printf("worker[%d]: prepare\n", job.wid)
	}
	return true
}

func (job *BaseJob) Finish() {
	if job.verbose >= 2 {
		log.Printf("worker[%d]: finalize\n", job.wid)
	}
	job.conn.Close()
}
