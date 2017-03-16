usage
----

gonewsyslog [/etc/gosyslog.conf] [-v] [-t]

* -t: test mode
* -v: verbose mode


features
----

* logging use github.com/uber-go/zap
* config TOML
* support log compression
  * gzip(use go package)
  * bz2(require installation bz2 and link path)
  * xz(require installation xz and link path)
* support cross device rotation
* support rename by go time format( https://golang.org/pkg/time/#pkg-constants )
