package gonewsyslog

import (
	"errors"
	"os/exec"

	"github.com/spamoc/gonewsyslog/internal/archiver"
	"github.com/spamoc/gonewsyslog/internal/rotate"
)

func canRotateBySize(info rotate.LogInfo, rule string) bool {
	size, _ := sizeToInt(rule)
	return info.Size >= size
}

func canRotateByCron(info rotate.LogInfo, rule string) bool {
	//Not Implements yet
	return false
}

type archiveType string

// if you change it, should change conf/conf const, too.
const (
	GZIP archiveType = "gzip"
	BZ2              = "bz2"
	XZ               = "xz"
	NONE             = "none"
)

func New(name string, from string, to string, archiveTypeStr string, ext string, success string, failed string, pidPath string, rTerm string, rSize string, rCount int) (NewSyslogJob, error) {
	var archive archiver.Archive = nil
	switch archiveType(archiveTypeStr) {
	case GZIP:
		archive = archiver.GZArchive
	case BZ2:
		archive = archiver.BZ2Archive
	case XZ:
		archive = archiver.XZArchive
	case NONE:
		archive = nil
	default:
		return NewSyslogJob{}, errors.New("archive type is invalid." + archiveTypeStr)
	}

	// define condition evaluator
	cond := condition{}
	if rTerm != "" {
		cond = condition{
			rule:   rTerm,
			evalIf: canRotateByCron,
		}
	} else if rSize != "" {
		if _, err := sizeToInt(rSize); err != nil {
			return NewSyslogJob{}, err
		}
		cond = condition{
			rule:   rSize,
			evalIf: canRotateBySize,
		}
	} else {
		//always return true(always rotate log)
		cond = condition{
			rule: "",
			evalIf: func(_ rotate.LogInfo, _ string) bool {
				return true
			},
		}
	}

	return NewSyslogJob{
		name:      name,
		from:      from,
		to:        to,
		archiveIf: archive,
		ext:       ext,
		cond:      cond,
		count:     rCount,
		success:   func() { exec.Command(success).Start() },
		failed:    func() { exec.Command(failed).Start() },
		pidPath:   pidPath,
	}, nil
}
