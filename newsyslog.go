package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spamoc/gonewsyslog/internal/conf"
	"github.com/spamoc/gonewsyslog/internal/gonewsyslog"
	"github.com/spamoc/gonewsyslog/internal/log"
	"github.com/spamoc/gonewsyslog/internal/rotate"
)

type Param struct {
	t bool
	v bool
	c string
}

func readArgs() (Param, error) {
	// ignore over arguments
	t := flag.Bool("t", false, "test mode(do not execute)")
	v := flag.Bool("v", false, "output verbose")
	flag.Parse()
	if len(flag.Args()) == 0 {
		return Param{}, errors.New("usage: this [ -t(test mode) -v(log verbose) ] configPath")
	}
	c := flag.Args()[0]
	if c == "" {
		return Param{}, errors.New("usage: this [ -t(test mode) -v(log verbose) ] configPath")
	}
	params := Param{t: *t, v: *v, c: c}
	return params, nil
}

func parse(c conf.Config, ctx context.Context) ([]rotate.RotateJob, error) {
	logger := log.Get(ctx)
	if c.Common.Include != "" {
		paths, err := filepath.Glob(c.Common.Include)
		if err != nil {
			logger.Error("include file read error. through.", log.Error(err))
		} else {
			c.Projects = append(c.Projects, walk(paths, ctx)...)
		}

	}
	return createJobs(c.Projects, ctx)
}

func walk(paths []string, ctx context.Context) []conf.ProjectConfig {
	logger := log.Get(ctx)
	var projects []conf.ProjectConfig = make([]conf.ProjectConfig, 0)
	for _, path := range paths {
		in, err := conf.New(path)
		if err != nil {
			logger.Error("include file read error. through.", log.Error(err))
		} else if len(in.Projects) > 0 {
			projects = append(projects, in.Projects...)
		}
	}
	return projects
}

func createJobs(projects []conf.ProjectConfig, ctx context.Context) ([]rotate.RotateJob, error) {
	jobs := make([]rotate.RotateJob, 0)
	for _, p := range projects {
		job, err := gosyslog.New(p.Name, p.From, p.To, string(p.Compress.Type), p.Compress.Ext, p.Success, p.Failed, p.Pid, p.Rotate.Term, p.Rotate.Size, p.Rotate.Count)
		if err != nil {
			return make([]rotate.RotateJob, 0), err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	params, err := readArgs()
	if err != nil {
		panic(err)
	}

	logger := log.Create(params.v)
	ctx := context.Background()
	ctx = log.Set(ctx, logger)
	c, err := conf.New(params.c)
	if err != nil {
		logger.Fatal("config parse error.", log.Error(err))
		os.Exit(1)
	}
	jobs, err := parse(c, ctx)
	if err != nil {
		logger.Fatal("job create error.", log.Error(err))
		os.Exit(1)
	}
	wg := &sync.WaitGroup{}
	for _, j := range jobs {
		wg.Add(1)
		go func(params Param, j rotate.RotateJob) {
			logger.Debug("running job info > ", log.Any("job", j))
			if params.t {
				result := j.Test(ctx)
				logger.Info("test result: ", log.Any("result", result))
			} else {
				result := j.Run(ctx)
				logger.Info("result: ", log.Any("result", result))
			}
			wg.Done()
		}(params, j)
	}
	wg.Wait()
}
