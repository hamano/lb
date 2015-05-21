package main

import (
	"os"
	"fmt"
	"log"
	"time"
//	"errors"
	"runtime"
//	"reflect"
	"github.com/codegangsta/cli"
	"github.com/mqu/openldap"
	"github.com/satori/go.uuid"
	"./job"
)

type Result struct {
	wid int
	count int
	success int
	startTime time.Time
	endTime time.Time
	elapsedTime float64
}

type AddJob struct {
	job.BaseJob
}

func (j *AddJob) Request() bool {
	cn := uuid.NewV1().String()
	dn := fmt.Sprintf("cn=%s,dc=example,dc=com", cn)
	attrs := map[string][]string{
		"objectClass": {"person"},
		"cn": {cn},
		"sn": {"test"},
		"userPassword": {"secret"},
	}
	err := j.Ldap.Add(dn, attrs)
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
	j job.Job) {
	num := c.Int("n")
	j.Init(wid, c)
	var result Result
	tx <- result
	<- rx
	if j.GetVerbose() >= 2 {
		log.Printf("worker[%d]: starting job\n", wid)
	}
	result.startTime = time.Now()

	for i := 0; i < num; i++ {
		res := j.Request()
		if res {
			j.IncSuccess()
		}
		j.IncCount()
	}
	result.endTime = time.Now()
	result.elapsedTime = result.endTime.Sub(result.startTime).Seconds()
	result.wid = wid
	result.count = j.GetCount()
	result.success = j.GetSuccess()
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

	fmt.Printf("Concurrency Level: %d\n", concurrency)
	fmt.Printf("Total Requests: %d\n", totalRequest)
	fmt.Printf("Success Requests: %d\n", successRequest)
	fmt.Printf("Success Rate: %d%%\n", successRequest / totalRequest * 100)
	fmt.Printf("Time taken for tests: %.3f seconds\n", takenTime)
	fmt.Printf("Requests per second: %.2f [#/sec] (mean)\n", rpq)
	fmt.Printf("Time per request: %.3f [ms] (mean)\n",
		float64(concurrency) * takenTime * 1000 / float64(totalRequest))
	fmt.Printf("Time per request: %.3f [ms] " +
		"(mean, across all concurrent requests)\n",
		takenTime * 1000 / float64(totalRequest))
	fmt.Printf("CPU Number: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

func add(c *cli.Context) {
	workerNum := c.Int("c");
	tx := make(chan string)
	rx := make(chan Result)

	for i := 0; i < workerNum; i++ {
		j := &AddJob{}
		go worker(i, c, tx, rx, j)
	}
	waitReady(rx, workerNum)
	// all worker are ready
	for i := 0; i < workerNum; i++ {
		tx <- "start"
	}
	results := waitResult(rx, workerNum)
	reportResult(c, results)
}

func checkArgs(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowAppHelp(c)
		return nil
	}

	fmt.Printf("This is LDAPBench, Version %s\n", c.App.Version)
	fmt.Printf("This software is released under the MIT License.\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	return nil
}

func setup(c *cli.Context) {
	baseDN := c.String("b")
	fmt.Printf("Adding base entry: %s\n", baseDN)
	ldap, err := openldap.Initialize(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}
	attrs := map[string][]string{
		"objectClass": {"dcObject", "organization"},
		"dc": {"example"},
		"o": {"example"},
	}
	err = ldap.Add(baseDN, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	fmt.Printf("Added base entry: %s\n", baseDN)
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
	cli.StringFlag {
		Name: "b",
		Value: "dc=example,dc=com",
		Usage: "BaseDN",
	},
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	app.Name = "lb"
	app.Usage = "LDAP Benchmarking Tool"
	app.Version = "0.1.0"
	app.Author = "HAMANO Tsukasa <hamano@osstech.co.jp>"
	app.Commands = []cli.Command{
		{
			Name: "bind",
			Usage: "LDAP BIND Test",
			Before: checkArgs,
			Action: job.Bind,
			Flags: commonFlags,
		},
		{
			Name: "add",
			Usage: "LDAP ADD Test",
			Before: checkArgs,
			Action: add,
			Flags: commonFlags,
		},
		{
			Name: "setup",
			Usage: "Add Base Entry",
			Before: checkArgs,
			Action: setup,
			Flags: commonFlags,
		},
	}
	app.Run(os.Args)
}
