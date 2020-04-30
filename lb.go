package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/urfave/cli"
)

func worker(wid int,
	c *cli.Context,
	rx chan string,
	tx chan Result,
	job Job) {

	num := c.Int("n")
	num_per_worker := int(math.Ceil(float64(num) / float64(c.Int("c"))))

	job.Init(wid, c)
	job.Prep(c)

	tx <- Result{}
	<-rx
	if job.GetVerbose() >= 2 {
		log.Printf("worker[%d]: starting job\n", wid)
	}
	var result Result
	result.startTime = time.Now()

	for i := 0; i < num_per_worker; i++ {
		res := job.Request()
		if res {
			job.IncSuccess()
		}
		job.IncCount()
	}
	result.endTime = time.Now()
	job.Finish()
	result.elapsedTime = result.endTime.Sub(result.startTime).Seconds()
	result.wid = wid
	result.count = job.GetCount()
	result.success = job.GetSuccess()
	tx <- result
}

func waitReady(ch chan Result, n int) {
	for i := 0; i < n; i++ {
		<-ch
	}
}

func waitResult(ch chan Result, n int) []Result {
	results := make([]Result, n)
	for i := 0; i < n; i++ {
		result := <-ch
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

	if ctx.Bool("short") {
		printShortResult(concurrency,
			int(float64(successRequest)/float64(totalRequest)*100), rpq)
	} else {
		printResult(concurrency, totalRequest, successRequest,
			int(float64(successRequest)/float64(totalRequest)*100),
			takenTime, rpq,
			float64(concurrency)*takenTime*1000/float64(totalRequest),
			takenTime*1000/float64(totalRequest))
	}
}

func printResult(
	concurrency int,
	totalRequest int,
	successRequest int,
	successRate int,
	takenTime float64,
	rpq float64,
	tpr float64,
	tpr_all float64) {
	fmt.Printf("Concurrency Level: %d\n", concurrency)
	fmt.Printf("Total Requests: %d\n", totalRequest)
	fmt.Printf("Success Requests: %d\n", successRequest)
	fmt.Printf("Success Rate: %d%%\n", successRate)
	fmt.Printf("Time taken for tests: %.3f seconds\n", takenTime)
	fmt.Printf("Requests per second: %.2f [#/sec] (mean)\n", rpq)
	fmt.Printf("Time per request: %.3f [ms] (mean)\n", tpr)
	fmt.Printf("Time per request: %.3f [ms] "+
		"(mean, across all concurrent requests)\n", tpr_all)
	fmt.Printf("CPU Number: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

func printShortResult(
	concurrency int,
	successRate int,
	rpq float64) {
	fmt.Printf("%d %.2f %d%%\n", concurrency, rpq, successRate)
}

func checkArgs(c *cli.Context) error {
	if c.Args().Len() < 1 {
		cli.ShowAppHelp(c)
		return errors.New("few args")
	}

	if !c.Bool("q") {
		fmt.Printf("This is LDAPBench, Version %s\n", c.App.Version)
		fmt.Printf("Copyright 2015 Open Source Solution Technology Corporation\n")
		fmt.Printf("This software is released under the MIT License.\n")
		fmt.Printf("\n")
	}
	return nil
}

func runBenchmark(c *cli.Context, jobType reflect.Type) {
	if !c.Bool("q") {
		fmt.Printf("%s Benchmarking: %s\n",
			jobType.Name(), c.Args().First())
	}

	workerNum := c.Int("c")
	tx := make(chan string)
	rx := make(chan Result)

	for i := 0; i < workerNum; i++ {
		job := reflect.New(jobType).Interface().(Job)
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

var commonFlags = []cli.Flag{
	&cli.IntFlag{
		Name:  "verbose, v",
		Value: 0,
		Usage: "How much troubleshooting info to print",
	},
	&cli.BoolFlag{
		Name:  "quiet, q",
		Usage: "Quiet flag",
	},
	&cli.IntFlag{
		Name:  "n",
		Value: 1,
		Usage: "Number of requests to perform",
	},
	&cli.IntFlag{
		Name:  "c",
		Value: 1,
		Usage: "Number of multiple requests to make",
	},
	&cli.StringFlag{
		Name:  "D",
		Value: "cn=Manager,dc=example,dc=com",
		Usage: "Bind DN",
	},
	&cli.StringFlag{
		Name:  "w",
		Value: "secret",
		Usage: "Bind Secret",
	},
	&cli.StringFlag{
		Name:  "b",
		Value: "dc=example,dc=com",
		Usage: "BaseDN",
	},
	&cli.BoolFlag{
		Name:  "starttls, Z",
		Usage: "Use StartTLS",
	},
	&cli.BoolFlag{
		Name:  "short",
		Usage: "short result",
	},
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	app.Name = "lb"
	app.Usage = "LDAP Benchmarking Tool"
	app.Version = Version
	app.Authors = []*cli.Author{
		{
			Email: "hamano@osstech.co.jp",
			Name:  "HAMANO Tsukasa",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:   "add",
			Usage:  "LDAP ADD Benchmarking",
			Before: checkArgs,
			Action: Add,
			Flags:  append(commonFlags, addFlags...),
		},
		{
			Name:   "bind",
			Usage:  "LDAP BIND Benchmarking",
			Before: checkArgs,
			Action: Bind,
			Flags:  append(commonFlags, bindFlags...),
		},
		{
			Name:   "delete",
			Usage:  "LDAP DELETE Benchmarking",
			Before: checkArgs,
			Action: Delete,
			Flags:  commonFlags,
		},
		{
			Name:   "modify",
			Usage:  "LDAP MODIFY Benchmarking",
			Before: checkArgs,
			Action: Modify,
			Flags:  append(commonFlags, modifyFlags...),
		},
		{
			Name:   "search",
			Usage:  "LDAP SEARCH Benchmarking",
			Before: checkArgs,
			Action: Search,
			Flags:  append(commonFlags, searchFlags...),
		},
		{
			Name:  "setup",
			Usage: "Setup SubCommands",
			Subcommands: []*cli.Command{
				{
					Name:   "base",
					Usage:  "Add Base Entry",
					Before: checkArgs,
					Action: setupBase,
					Flags:  commonFlags,
				},
				{
					Name:   "person",
					Usage:  "Add User Entry",
					Before: checkArgs,
					Action: setupPerson,
					Flags:  append(commonFlags, setupPersonFlags...),
				},
			},
		},
	}
	app.Run(os.Args)
}
