package sqlswitch

import (
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	sd Driver
)

type fakeDriver struct{}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

type fakeConn struct{}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{}, nil
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return &fakeTx{}, nil
}

type fakeTx struct{}

func (t *fakeTx) Commit() error {
	return nil
}

func (t *fakeTx) Rollback() error {
	return nil
}

type fakeStmt struct{}

func (s *fakeStmt) Close() error {
	return nil
}

func (s *fakeStmt) NumInput() int {
	return 1
}

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &fakeResult{}, nil
}

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &fakeRows{}, nil
}

type fakeRows struct{}

func (r *fakeRows) Columns() []string {
	return []string{}
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	return nil
}

type fakeResult struct{}

func (r *fakeResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *fakeResult) RowsAffected() (int64, error) {
	return 2, nil
}

/*
func init() {
	fake := &fakeDriver{}
	sd = New(fake.Open)
	sql.Register("sqlswitch", sd)
}

func TestDriver(t *testing.T) {
	db, _ := sql.Open("sqlswitch", "dsn1")

	fmt.Printf("%+v\n", sd)

	fmt.Println(db)
	fmt.Println(sd.GetDSN())
}
*/

type mockFuncs struct {
	mock.Mock
}

func (m *mockFuncs) TimeNow() time.Time {
	ret := m.Called()
	return ret.Get(0).(time.Time)
}

type driverSuite struct {
	suite.Suite
	mockFuncs *mockFuncs
}

func (d *driverSuite) SetupTest() {
	d.mockFuncs = &mockFuncs{}
	timeNow = d.mockFuncs.TimeNow
}

func (d *driverSuite) TestOpenSrc() {
	fake := &fakeDriver{}
	sd := New(fake.Open)

	dsn1 := "dsn1"
	sd.Open(dsn1)

	d.Equal(dsn1, sd.GetDSN())
}

func (d *driverSuite) TestOpenDst() {
	fake := &fakeDriver{}
	sd := New(fake.Open)

	dsn1 := "dsn1"
	dsn2 := "dsn2"
	dsn := fmt.Sprintf("%s%s%s", dsn1, sep, dsn2)
	sd.Open(dsn)

	d.Equal(dsn1, sd.GetDSN())

	loc, _ := time.LoadLocation("Asia/Taipei")
	t0 := time.Date(2018, 3, 29, 12, 10, 0, 0, loc)
	t1 := t0.Unix()
	t2 := t0.Add(1 * time.Minute).Unix()
	t3 := t0.Add(5 * time.Minute).Unix()
	t4 := t0.Add(6 * time.Minute).Unix()

	c1 := Config{
		Target:           "dst",
		SrcPoolEndTime:   t1,
		DstPoolStartTime: t2,
		DstPoolEndTime:   t3,
		BakPoolStartTime: t4,
	}
	sd.ApplyConfig(c1)
	d.mockFuncs.On("TimeNow").Return(t0.Add(-2 * time.Minute))

	sd.Open(dsn)
	d.Equal(dsn1, sd.GetDSN())

	// Reset mock
	d.mockFuncs = &mockFuncs{}
	timeNow = d.mockFuncs.TimeNow
	d.mockFuncs.On("TimeNow").Return(t0.Add(2 * time.Minute))

	sd.Open(dsn)
	d.Equal(dsn2, sd.GetDSN())
}

func (d *driverSuite) TestOpenBak() {
	fake := &fakeDriver{}
	sd := New(fake.Open)

	dsn1 := "dsn1"
	dsn2 := "dsn2"
	dsn3 := "dsn3"
	dsn := fmt.Sprintf("%s%s%s%s%s", dsn1, sep, dsn2, sep, dsn3)
	sd.Open(dsn)

	d.Equal(dsn1, sd.GetDSN())

	loc, _ := time.LoadLocation("Asia/Taipei")
	t0 := time.Date(2018, 3, 29, 12, 10, 0, 0, loc)
	t1 := t0.Unix()
	t2 := t0.Add(1 * time.Minute).Unix()
	t3 := t0.Add(5 * time.Minute).Unix()
	t4 := t0.Add(6 * time.Minute).Unix()

	c1 := Config{
		Target:           "bak",
		SrcPoolEndTime:   t1,
		DstPoolStartTime: t2,
		DstPoolEndTime:   t3,
		BakPoolStartTime: t4,
	}
	sd.ApplyConfig(c1)
	d.mockFuncs.On("TimeNow").Return(t0.Add(2 * time.Minute))

	sd.Open(dsn)
	d.Equal(dsn2, sd.GetDSN())

	// Reset mock
	d.mockFuncs = &mockFuncs{}
	timeNow = d.mockFuncs.TimeNow
	d.mockFuncs.On("TimeNow").Return(t0.Add(8 * time.Minute))

	sd.Open(dsn)
	d.Equal(dsn3, sd.GetDSN())
}

func (d *driverSuite) TestOpenShouldPanic() {
	fake := &fakeDriver{}
	sd := New(fake.Open)

	dsn1 := "dsn1"
	loc, _ := time.LoadLocation("Asia/Taipei")
	t0 := time.Date(2018, 3, 29, 12, 10, 0, 0, loc)
	d.mockFuncs.On("TimeNow").Return(t0)

	sd.ApplyConfig(Config{Target: "gg"})
	d.Panics(
		func() {
			sd.Open(dsn1)
		},
		"Open should panic",
	)
}

func (d *driverSuite) TestOpenWithError() {
	fake := &fakeDriver{}
	sd := New(fake.Open)

	dsn1 := "dsn1"
	dsn2 := "dsn2"
	dsn := fmt.Sprintf("%s%s%s", dsn1, sep, dsn2)
	sd.Open(dsn)

	d.Equal(dsn1, sd.GetDSN())

	loc, _ := time.LoadLocation("Asia/Taipei")
	t0 := time.Date(2018, 3, 29, 12, 10, 0, 0, loc)
	t1 := t0.Unix()
	t2 := t0.Add(2 * time.Minute).Unix()
	t3 := t0.Add(5 * time.Minute).Unix()
	t4 := t0.Add(6 * time.Minute).Unix()

	c1 := Config{
		Target:           "dst",
		SrcPoolEndTime:   t1,
		DstPoolStartTime: t2,
		DstPoolEndTime:   t3,
		BakPoolStartTime: t4,
	}
	sd.ApplyConfig(c1)
	d.mockFuncs.On("TimeNow").Return(t0.Add(1 * time.Minute))

	_, err := sd.Open(dsn)
	d.Equal(ErrGapTime, err)
}

func TestDriverSuite(t *testing.T) {
	suite.Run(t, new(driverSuite))
}
