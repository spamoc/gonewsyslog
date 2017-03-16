package gonewsyslog

import "github.com/spamoc/gonewsyslog/internal/rotate"

type Evaluate func(info rotate.LogInfo, rule string) bool
