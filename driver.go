package sqlswitch

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	sep = "&&&"
)

var (
	ErrGapTime = fmt.Errorf("gap time")
	timeNow    = time.Now
)

type Config struct {
	Target           string
	SrcPoolEndTime   int64
	DstPoolStartTime int64
	DstPoolEndTime   int64
	BakPoolStartTime int64
}

type OpenFunc func(dsn string) (driver.Conn, error)

type Driver interface {
	driver.Driver

	ApplyConfig(c Config)

	GetDSN() string
}

func New(open OpenFunc) Driver {
	return &switchDriver{open: open}
}

type switchDriver struct {
	open OpenFunc
	c    Config
	dsn  string
}

func (sd *switchDriver) Open(dsn string) (driver.Conn, error) {
	dsns := strings.Split(dsn, sep)
	var srcDsn, dstDsn, bakDsn string
	if len(dsns) == 1 {
		srcDsn = dsns[0]
	}
	if len(dsns) == 2 {
		srcDsn = dsns[0]
		dstDsn = dsns[1]
	}
	if len(dsns) == 3 {
		srcDsn = dsns[0]
		dstDsn = dsns[1]
		bakDsn = dsns[2]
	}

	if sd.c.Target == "src" || sd.c.Target == "" {
		sd.dsn = srcDsn
		return sd.open(srcDsn)
	}

	ts := timeNow().Unix()

	switch sd.c.Target {
	case "dst":
		if ts <= sd.c.SrcPoolEndTime {
			sd.dsn = srcDsn
			return sd.open(srcDsn)
		} else if ts >= sd.c.DstPoolStartTime {
			sd.dsn = dstDsn
			return sd.open(dstDsn)
		}
	case "bak":
		if ts <= sd.c.DstPoolEndTime {
			sd.dsn = dstDsn
			return sd.open(dstDsn)
		} else if ts >= sd.c.BakPoolStartTime {
			sd.dsn = bakDsn
			return sd.open(bakDsn)
		}
	default:
		panic("invalid target")
	}

	return nil, ErrGapTime
}

func (sd *switchDriver) ApplyConfig(c Config) {
	sd.c = c
}

func (sd *switchDriver) GetDSN() string {
	return sd.dsn
}
