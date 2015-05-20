package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"runtime"
//	"reflect"
	"github.com/codegangsta/cli"
	"github.com/satori/go.uuid"
	"github.com/mqu/openldap"
)

type Result struct {
	wid int
	count int
	success int
	startTime time.Time
	endTime time.Time
	elapsedTime float64
}

type Job interface {
	init(int, *cli.Context) bool
	request() bool
	getVerbose() int
	IncCount()
	getCount() int
	IncSuccess()
	getSuccess() int
}

type BaseJob struct {
	ldap *openldap.Ldap
	wid int
	count int
	success int
	verbose int
}

func (job *BaseJob) request() bool {
	if job.verbose >= 3 {
		log.Printf("[%d]: %d\n", job.wid, job.count)
	}
	time.Sleep(1000 * time.Millisecond)
	return true
}

func (job *BaseJob) getVerbose() int {
	return job.verbose
}

func (job *BaseJob) IncCount() {
	job.count++
}

func (job *BaseJob) getCount() int {
	return job.count
}

func (job *BaseJob) IncSuccess() {
	job.success++
}

func (job *BaseJob) getSuccess() int {
	return job.success
}

type AddJob struct {
	BaseJob
}

func (job *BaseJob) init(wid int, c *cli.Context) bool {
	job.wid = wid
	job.verbose = c.Int("verbose")
	url := c.Args().First()

	if job.verbose >= 2 {
		log.Printf("worker[%d]: initialize %s\n", job.wid, url)
	}
	var err error
	job.ldap, err = openldap.Initialize(url)
	if err != nil {
		log.Fatal("initialize err=%d\n", err)
		return false
	}
	job.ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	//defer ldap.Close()
	err = job.ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Printf("bind err: %s", err)
		return false
	}
	return true
}

func (job *AddJob) request() bool {
	cn := uuid.NewV1().String()
	dn := fmt.Sprintf("cn=%s,dc=example,dc=com", cn)
	attrs := map[string][]string{
		"objectClass": {"person"},
		"cn": {cn},
		"sn": {"test"},
		"userPassword": {"secret"},
	}
	err := job.ldap.Add(dn, attrs)
	if err != nil {
		log.Printf("add err: %s", err)
		return false
	}
	return true
}

func worker(wid int,
	c *cli.Context,
	rx chan string,
	tx chan Result,
	job Job) {
	num := c.Int("n")
	job.init(wid, c)
	var result Result
	tx <- result
	<- rx
	if job.getVerbose() >= 2 {
		log.Printf("worker[%d]: starting job\n", wid)
	}
	result.startTime = time.Now()

	for i := 0; i < num; i++ {
		res := job.request()
		if res {
			job.IncSuccess()
		}
		job.IncCount()
	}
	result.endTime = time.Now()
	result.elapsedTime = result.endTime.Sub(result.startTime).Seconds()
	result.wid = wid
	result.count = job.getCount()
	result.success = job.getSuccess()
	tx <- result
}

func waitReady(ch chan Result, n int){
	for i := 0; i < n; i++ {
		<- ch
	}
}

func waitResult(ch chan Result, n int) []Result {
	results := make([]Result, n)
	for i := 0; i < n; i++ {
		result := <- ch
		results[result.wid] = result
	}
	return results
}

func reportResult(ctx *cli.Context, results []Result) {
	var firstTime time.Time
	var lastTime time.Time
	var totalRequest int
	var successRequest int
	for i := range results {
		rpq := float64(results[i].count) / results[i].elapsedTime
		if ctx.Int("v") >= 2 {
			log.Printf("worker[%d]: %.2f [#/sec] time=%.3f\n",
				results[i].wid, rpq, results[i].elapsedTime)
		}

		totalRequest += results[i].count
		successRequest += results[i].success
		if firstTime.IsZero() || firstTime.After(results[i].startTime) {
			firstTime = results[i].startTime
		}
		if lastTime.Before(results[i].endTime) {
			lastTime = results[i].endTime
		}
	}
	takenTime := lastTime.Sub(firstTime).Seconds()
	rpq := float64(totalRequest) / takenTime
	concurrency := ctx.Int("c")

	log.Printf("Concurrency Level: %d\n", concurrency)
	log.Printf("Time taken for tests: %.3f seconds\n", takenTime)
	log.Printf("Requests per second: %.2f [#/sec] (mean)\n", rpq)
	log.Printf("Time per request: %.3f [ms] (mean)\n",
		float64(concurrency) * takenTime * 1000 / float64(totalRequest))
	log.Printf("Time per request: %.3f [ms] " +
		"(mean, across all concurrent requests)\n",
		takenTime * 1000 / float64(totalRequest))

}

func add(c *cli.Context) {
	if len(c.Args()) < 1 {
		cli.ShowAppHelp(c)
		return
	}
	workerNum := c.Int("c");
	tx := make(chan string)
	rx := make(chan Result)

	for i := 0; i < workerNum; i++ {
		job := &AddJob{}
		go worker(i, c, tx, rx, job)
	}
	waitReady(rx, workerNum)
	// all worker are ready
	for i := 0; i < workerNum; i++ {
		tx <- "start"
	}
	results := waitResult(rx, workerNum)
	reportResult(c, results)
}

func bind(c *cli.Context) {

}

var commonFlags = []cli.Flag {
	cli.IntFlag {
		Name: "verbose, v",
		Value: 0,
		Usage: "How much troubleshooting info to print",
	},
	cli.IntFlag {
		Name: "n",
		Value: 1,
		Usage: "Number of requests to perform",
	},
	cli.IntFlag {
		Name: "c",
		Value: 1,
		Usage: "Number of multiple requests to make",
	},
	cli.StringFlag {
		Name: "D",
		Value: "cn=Manager,dc=example,dc=com",
		Usage: "Bind DN",
	},
	cli.StringFlag {
		Name: "w",
		Value: "secret",
		Usage: "Bind Secret",
	},
}

func main() {
	log.Printf("CPU Number: %d\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	app := cli.NewApp()
	app.Name = "lb"
	app.Commands = []cli.Command{
		{
			Name: "add",
			Usage: "LDAP ADD Benchmarking",
			Action: add,
			Flags: commonFlags,
		},
		{
			Name: "bind",
			Usage: "LDAP BIND Benchmarking",
			Action: bind,
			Flags: commonFlags,
		},
	}
	app.Run(os.Args)
}
