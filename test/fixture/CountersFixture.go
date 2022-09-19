package test_fixture

import (
	"testing"
	"time"

	ccount "github.com/pip-services3-gox/pip-services3-components-gox/count"
	"github.com/stretchr/testify/assert"
)

type CountersFixture struct {
	counters *ccount.CachedCounters
}

func NewCountersFixture(counters *ccount.CachedCounters) *CountersFixture {
	cf := CountersFixture{}
	cf.counters = counters
	return &cf
}

func (c *CountersFixture) TestSimpleCounters(t *testing.T) {
	c.counters.Last("Test.LastValue", 123)
	c.counters.Last("Test.LastValue", 123456)

	var counter = c.counters.Get("Test.LastValue", ccount.LastValue)
	assert.NotNil(t, counter)
	assert.NotNil(t, counter.Last)
	assert.Equal(t, counter.Last, float32(123456), 3)

	c.counters.IncrementOne("Test.Increment")
	c.counters.Increment("Test.Increment", 3)

	counter = c.counters.Get("Test.Increment", ccount.Increment)
	assert.NotNil(t, counter)
	assert.Equal(t, counter.Count, 4)

	c.counters.TimestampNow("Test.Timestamp")
	c.counters.TimestampNow("Test.Timestamp")

	counter = c.counters.Get("Test.Timestamp", ccount.Timestamp)
	assert.NotNil(t, counter)
	assert.NotNil(t, counter.Time)

	c.counters.Stats("Test.Statistics", 1)
	c.counters.Stats("Test.Statistics", 2)
	c.counters.Stats("Test.Statistics", 3)

	counter = c.counters.Get("Test.Statistics", ccount.Statistics)
	assert.NotNil(t, counter)
	assert.Equal(t, counter.Average, float32(2), 3)

	c.counters.Dump()

	select {
	case <-time.After(1000 * time.Millisecond):
		{
		}
	}
}

func (c *CountersFixture) TestMeasureElapsedTime(t *testing.T) {
	timer := c.counters.BeginTiming("Test.Elapsed")

	time.AfterFunc(100*time.Millisecond, func() {
		timer.EndTiming()

		counter := c.counters.Get("Test.Elapsed", ccount.Interval)
		assert.True(t, counter.Last > 50)
		assert.True(t, counter.Last < 5000)

		c.counters.Dump()
		select {
		case <-time.After(1000 * time.Millisecond):
			{
			}
		}
	})
}
