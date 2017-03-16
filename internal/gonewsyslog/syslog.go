package gonewsyslog

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spamoc/gonewsyslog/internal/archiver"
	"github.com/spamoc/gonewsyslog/internal/log"
	"github.com/spamoc/gonewsyslog/internal/rotate"
)

var unitMap = map[string]uint64{
	"K": 1024,
	"M": 1024 * 1024,
	"G": 1024 * 1024 * 1024,
	"T": 1024 * 1024 * 1024 * 1024,
	"P": 1024 * 1024 * 1024 * 1024 * 1024,
	"E": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
}

func sizeToInt(size string) (uint64, error) {
	if strings.HasSuffix(size, "B") {
		l := len(size)
		unit := size[l-2 : l-1]
		if unitMap[unit] != 0 {
			i, err := strconv.ParseInt(size[0:l-2], 10, 64)
			if err == nil {
				return uint64(i) * unitMap[unit], nil
			} else {
				return 0, err
			}
		} else {
			return 0, errors.New("configuration error: size unit is not compatible. " + size)
		}
	} else {
		i, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return 0, errors.New("configuration error: size unit is not compatible. " + size)
		}
		return uint64(i), nil
	}
}

// newsyslog struct define

type NewSyslogJob struct {
	name      string
	from      string
	to        string
	archiveIf archiver.Archive
	ext       string
	count     int
	cond      condition
	pidPath   string
	success   rotate.Callback
	failed    rotate.Callback
}

func (this NewSyslogJob) archive(from string, to string) error {
	return this.archiveIf(from, to)
}

type condition struct {
	rule   string
	evalIf Evaluate
}

func (this *condition) eval(info rotate.LogInfo) bool {
	return this.evalIf(info, this.rule)
}

func (this condition) String() string {
	return fmt.Sprintf("rule: %s", this.rule)
}

func (this NewSyslogJob) Run(ctx context.Context) rotate.Result {
	logger := log.Get(ctx)
	r := rotate.Result{}
	r.Name = this.name
	status, _ := this.LogState(ctx)
	logger.Debug("log status", log.String("name", this.name), log.Any("status", status))
	if x, err := this.CanRotate(status, ctx); x == true && err == nil {
		err := this.Rotate(ctx)
		if err == nil {
			this.success()
			r.Status = rotate.SUCCESS
		} else {
			this.failed()
			r.Status = rotate.FAIL
			r.Err = err
			return r
		}
	} else if err == nil {
		r.Status = rotate.SKIP
	} else {
		this.failed()
		r.Status = rotate.FAIL
		r.Err = err
	}
	return r
}

func (this NewSyslogJob) Test(ctx context.Context) rotate.Result {
	logger := log.Get(ctx)
	r := rotate.Result{}
	r.Name = this.name
	status, _ := this.LogState(ctx)
	logger.Debug("log status : ", log.String("name", this.name), log.Any("status", status))

	if x, err := this.CanRotate(status, ctx); x == true && err == nil {
		r.Status = rotate.SUCCESS
	} else if err == nil {
		r.Status = rotate.SKIP
	} else {
		r.Status = rotate.FAIL
		r.Err = err
	}
	return r
}

func (this NewSyslogJob) Rotate(ctx context.Context) error {
	logger := log.Get(ctx)
	stats, _ := os.Stat(this.from)
	mode := stats.Mode()
	// copy
	to := time.Now().Format(this.to)
	bt, err := ioutil.ReadFile(this.from)
	if err != nil {
		logger.Error("move error", log.Error(err), log.String("jobname", this.name))
		return err
	}
	err = ioutil.WriteFile(to, bt, mode)
	if err != nil {
		logger.Error("move error", log.Error(err), log.String("jobname", this.name))
		return err
	}
	bt = make([]byte, 0)
	// clear from file
	ioutil.WriteFile(this.from, bt, mode)

	// get pid
	if this.pidPath != "" {
		pidbytes, err := ioutil.ReadFile(this.pidPath)
		if err != nil {
			logger.Error("get pid file error", log.Error(err), log.String("jobname", this.name))
			return err
		}
		// trim line break
		pid := strings.TrimRight(string(pidbytes), "\n")
		logger.Debug("kill info", log.String("pid", pid), log.String("jobname", this.name))
		// send signal
		if err := exec.Command("kill", "-HUP", pid).Run(); err != nil {
			logger.Error("process kill error", log.Error(err), log.String("jobname", this.name))
			return err
		}
	}
	// compress
	if this.archiveIf != nil {
		// moved file to moved file+ext
		if err := this.archive(to, to+this.ext); err != nil {
			logger.Error("compress error", log.Error(err), log.String("jobname", this.name))
			return err
		}
		logger.Debug("compress succeed", log.String("from", to), log.String("to", to+this.ext), log.String("jobname", this.name))
	}

	if err := this.evict(ctx); err != nil {
		logger.Error("evict error", log.Error(err), log.String("jobname", this.name))
		return err
	}
	return nil
}

func (this NewSyslogJob) LogState(_ context.Context) (rotate.LogInfo, error) {
	stat, err := os.Stat(this.from)
	if err != nil {
		return rotate.LogInfo{}, err
	}
	mtime := stat.ModTime()

	// implement status caching and using
	return rotate.LogInfo{
		Path:     this.from,
		ChangeAt: mtime,
		Size:     uint64(stat.Size()),
	}, nil
}

func (this NewSyslogJob) CanRotate(info rotate.LogInfo, _ context.Context) (bool, error) {
	return this.cond.eval(info), nil
}

func (this NewSyslogJob) Callback(_ context.Context) {

}

func (this NewSyslogJob) evict(ctx context.Context) error {
	logger := log.Get(ctx)
	allfile, err := filepath.Glob(filepath.Dir(this.to) + "/*")
	logger.Debug("allfile", log.Int("allfile size", len(allfile)))
	if err != nil {
		logger.Error("rotate error", log.Error(err), log.String("jobname", this.name))
		return err
	}
	paths := make([]string, 0)
	for _, path := range allfile {
		if _, err = time.Parse(this.to+this.ext, path); err == nil {
			paths = append(paths, path)
		}
	}

	logs := make([]rotate.LogInfo, 0)
	for _, path := range paths {
		stat, _ := os.Stat(path)
		mtime := stat.ModTime()
		logs = append(logs, rotate.LogInfo{
			Path:     path,
			ChangeAt: mtime,
		},
		)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ChangeAt.After(logs[j].ChangeAt)
	})
	logger.Debug("rotate info", log.String("log name pattern", this.to+this.ext), log.Int("exists log count", len(logs)))
	if len(logs) > this.count {
		for _, removal := range logs[this.count:] {
			if err := os.Remove(removal.Path); err != nil {
				logger.Error("rotate error", log.Error(err), log.String("jobname", this.name))
			} else {
				logger.Info("old log file deleted.", log.Int("exists log count", len(logs)), log.Int("max count", this.count), log.String("file", removal.Path))
			}
		}
	}
	return nil
}

func (this NewSyslogJob) String() string {
	return fmt.Sprintf("Job Status:{name: %s, from: %s, to: %s, count: %d, condition: %s, pid path: %s}\n", this.name, this.from, this.to, this.count, this.cond, this.pidPath)
}
