package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/influxdata/influxdb/tsdb/index/inmem"
)

func main() {
	tmpDir, err := ioutil.TempDir("", "shard_test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	tmpShard := filepath.Join(tmpDir, "shard")
	tmpWal := filepath.Join(tmpDir, "wal")

	sfile := NewSeriesFile()
	if err := sfile.Open(); err != nil {
		panic(err)
	}
	defer sfile.Close()

	opts := tsdb.NewEngineOptions()
	opts.WALEnabled = false
	opts.CompactionDisabled = true
	opts.MonitorDisabled = true
	opts.InmemIndex = inmem.NewIndex(filepath.Base(tmpDir), sfile.SeriesFile)
	opts.SeriesIDSets = seriesIDSets([]*tsdb.SeriesIDSet{})

	sh := tsdb.NewShard(1, tmpShard, tmpWal, sfile.SeriesFile, opts)
	if err := sh.Open(); err != nil {
		panic(err)
	}
	defer sh.Close()

	points := make([]models.Point, 0, 1000)
	for i := 0; i < cap(points); i++ {
		points = append(points, models.MustNewPoint(
			"cpu",
			models.NewTags(map[string]string{"host": "server"}),
			map[string]interface{}{"value": int64(1)},
			time.Unix(int64(i), 0),
		))
	}

	eng, err := sh.Engine()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 50; i++ {
		if err := eng.DeleteMeasurement([]byte("cpu")); err != nil {
			panic(err)
		}

		_ = sh.WritePoints(points[500:])
		_ = eng.(*tsm1.Engine).WriteSnapshot()
	}
}

// SeriesFile is a test wrapper for tsdb.SeriesFile.
type SeriesFile struct {
	*tsdb.SeriesFile
}

// NewSeriesFile returns a new instance of SeriesFile with a temporary file path.
func NewSeriesFile() *SeriesFile {
	dir, err := ioutil.TempDir("", "tsdb-series-file-")
	if err != nil {
		panic(err)
	}
	return &SeriesFile{SeriesFile: tsdb.NewSeriesFile(dir)}
}

// MustOpenSeriesFile returns a new, open instance of SeriesFile. Panic on error.
func MustOpenSeriesFile() *SeriesFile {
	f := NewSeriesFile()
	if err := f.Open(); err != nil {
		panic(err)
	}
	return f
}

// Close closes the log file and removes it from disk.
func (f *SeriesFile) Close() error {
	defer os.RemoveAll(f.Path())
	return f.SeriesFile.Close()
}

// Reopen close & reopens the series file.
func (f *SeriesFile) Reopen() error {
	if err := f.SeriesFile.Close(); err != nil {
		return err
	}
	f.SeriesFile = tsdb.NewSeriesFile(f.SeriesFile.Path())
	return f.SeriesFile.Open()
}

// ForceCompact executes an immediate compaction across all partitions.
func (f *SeriesFile) ForceCompact() error {
	for _, p := range f.Partitions() {
		if err := tsdb.NewSeriesPartitionCompactor().Compact(p); err != nil {
			return err
		}
	}
	return nil
}

type seriesIDSets []*tsdb.SeriesIDSet

func (a seriesIDSets) ForEach(f func(ids *tsdb.SeriesIDSet)) error {
	for _, v := range a {
		f(v)
	}
	return nil
}
