# sqlswitch

[![Build Status](https://travis-ci.org/17media/go-sql-switch.svg?branch=master)](https://travis-ci.org/17media/go-sql-switch) [![Coverage Status](https://coveralls.io/repos/github/17media/go-sql-switch/badge.svg?branch=master)](https://coveralls.io/github/17media/go-sql-switch?branch=master)

A specific purpose golang `database/sql` driver wrapper that helps [goapi](https://github.com/17media/api) switch between different MySQL sources while minimizing down time. It's **NOT** for general use.

### Usage

`sqlswitch` provides a wrapper for a `database/sql/driver.Driver` to support multiple data source names (DSNs) in the `Open` method. All DSNs should be put in the same string and separate by `&&&`. It takes up to **3** DSNs which will be marked as `src`, `dst` ,and `bak` in internal implementation.

- `src`: source database
- `dst`: destination database
- `bak`: backup database

The time of switching dsn can be determined in runtime by `ApplyConfig` method. The switch target and corresponding time can be set in `Config`.

```go
import (
	"database/sql"

	"github.com/17media/go-sql-switch"
)

var (
	sd sqlswitch.Driver
)

func init() {
	sd = sqlswitch.New(driver.Open)
	sql.Register("sqlswitch", sd)
}

func main() {
	db, err := sql.Open("sqlswitch", "dsn1&&&dsn2&&&dsn3")

	loc, _ := time.LoadLocation("Asia/Taipei")
	t0 := time.Date(2018, 3, 29, 12, 10, 0, 0, loc)
	t1 := t0.Unix()
	t2 := t0.Add(1 * time.Minute).Unix()
	t3 := t0.Add(5 * time.Minute).Unix()
	t4 := t0.Add(6 * time.Minute).Unix()

	c := sqlswitch.Config{
		Target:           "dst",
		SrcPoolEndTime:   t1,
		DstPoolStartTime: t2,
		DstPoolEndTime:   t3,
		BakPoolStartTime: t4,
	}

	sd.ApplyConfig(c)
}
```

