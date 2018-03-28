package sqlswitch

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
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
}

func New(open OpenFunc) Driver {
	return &switchDriver{open: open}
}

type switchDriver struct {
	open OpenFunc
	c    Config
}

func (sd *switchDriver) Open(dsn string) (driver.Conn, error) {
	dsns := strings.Split(dsn, "&&&")
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
		return sd.open(srcDsn)
	}

	ts := time.Now().Unix()

	switch sd.c.Target {
	case "dst":
		if ts <= sd.c.SrcPoolEndTime {
			return sd.open(srcDsn)
		} else if ts >= sd.c.DstPoolStartTime {
			return sd.open(dstDsn)
		}
	case "bak":
		if ts <= sd.c.DstPoolEndTime {
			return sd.open(dstDsn)
		} else if ts >= sd.c.BakPoolStartTime {
			return sd.open(bakDsn)
		}
	default:
		panic("invalid target")
	}

	return nil, fmt.Errorf("gap time")
}

func (sd *switchDriver) ApplyConfig(c Config) {
	sd.c = c
}
