package singleflight_cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewObservationIndications(t *testing.T) {
	obs, err := NewObservationIndications(context.Background(), 10*time.Second, 2)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		obs.AddHitCount()
		obs.AddHitTimeConsuming(100)
	}
	assert.Equal(t, int64(0), obs.GetCycleHitCount())
	assert.Equal(t, int64(10), obs.GetLastHitCount())

	for i := 0; i < 6; i++ {
		obs.AddMissCount()
		obs.AddMissTimeConsuming(200)
	}

	assert.Equal(t, int64(0), obs.GetCycleMissCount())
	assert.Equal(t, int64(6), obs.GetLastMissCount())

	for i := 0; i < 3; i++ {
		obs.AddErrorCount()
	}

	assert.Equal(t, int64(0), obs.GetCycleErrorCount())
	assert.Equal(t, int64(3), obs.GetLastErrorCount())

	for i := 0; i < 2; i++ {
		obs.AddManualMissCount()
	}

	assert.Equal(t, int64(0), obs.GetCycleManualMissCount())
	assert.Equal(t, int64(2), obs.GetLastManualMissCount())

	// hit 10 miss 6 err 3 manual 2

	assert.Equal(t, float64(0), obs.GetCycleHitRate())
	assert.Equal(t, float64(0), obs.GetCycleErrorRate())
	assert.Equal(t, float64(0), obs.GetCycleHitAvgTimeConsuming())
	assert.Equal(t, float64(0), obs.GetCycleMissAvgTimeConsuming())

	// hit 10 + miss 6 + err 3 - manual 2 = 17
	// hit 10 / 17 = 0.5882352941176471
	assert.Equal(t, float64(10)/float64(17), obs.GetLastHitRate())

	// err 3 / 17 = 0.17647058823529413
	assert.Equal(t, float64(3)/float64(17), obs.GetLastErrorRate())

	// hit 10 * 100 / 10 = 100
	assert.Equal(t, float64(100), obs.GetLastHitAvgTimeConsuming())

	// miss 6 * 200 / 6 = 200
	assert.Equal(t, float64(200), obs.GetLastMissAvgTimeConsuming())

	time.Sleep(11 * time.Second)

	assert.Equal(t, int64(10), obs.GetCycleHitCount())
	assert.Equal(t, int64(0), obs.GetLastHitCount())

	assert.Equal(t, int64(6), obs.GetCycleMissCount())
	assert.Equal(t, int64(0), obs.GetLastMissCount())

	assert.Equal(t, int64(3), obs.GetCycleErrorCount())
	assert.Equal(t, int64(0), obs.GetLastErrorCount())

	assert.Equal(t, int64(2), obs.GetCycleManualMissCount())
	assert.Equal(t, int64(0), obs.GetLastManualMissCount())

	assert.Equal(t, float64(0), obs.GetLastHitRate())
	assert.Equal(t, float64(0), obs.GetLastErrorRate())
	assert.Equal(t, float64(0), obs.GetLastHitAvgTimeConsuming())
	assert.Equal(t, float64(0), obs.GetLastMissAvgTimeConsuming())

	// hit 10 + miss 6 + err 3 - manual 2 = 17
	// hit 10 / 17 = 0.5882352941176471
	assert.Equal(t, float64(10)/float64(17), obs.GetCycleHitRate())

	// err 3 / 17 = 0.17647058823529413
	assert.Equal(t, float64(3)/float64(17), obs.GetCycleErrorRate())

	// hit 10 * 100 / 10 = 100
	assert.Equal(t, float64(100), obs.GetCycleHitAvgTimeConsuming())

	// miss 6 * 200 / 6 = 200
	assert.Equal(t, float64(200), obs.GetCycleMissAvgTimeConsuming())

	time.Sleep(11 * time.Second)

	assert.Equal(t, int64(0), obs.GetCycleHitCount())
	assert.Equal(t, int64(0), obs.GetLastHitCount())

	assert.Equal(t, int64(0), obs.GetCycleMissCount())
	assert.Equal(t, int64(0), obs.GetLastMissCount())

	assert.Equal(t, int64(0), obs.GetCycleErrorCount())
	assert.Equal(t, int64(0), obs.GetLastErrorCount())

	assert.Equal(t, int64(0), obs.GetCycleManualMissCount())
	assert.Equal(t, int64(0), obs.GetLastManualMissCount())

	assert.Equal(t, float64(0), obs.GetCycleHitRate())
	assert.Equal(t, float64(0), obs.GetCycleErrorRate())
	assert.Equal(t, float64(0), obs.GetCycleHitAvgTimeConsuming())
	assert.Equal(t, float64(0), obs.GetCycleMissAvgTimeConsuming())

	assert.Equal(t, float64(0), obs.GetLastHitRate())
	assert.Equal(t, float64(0), obs.GetLastErrorRate())
	assert.Equal(t, float64(0), obs.GetLastHitAvgTimeConsuming())
	assert.Equal(t, float64(0), obs.GetLastMissAvgTimeConsuming())

}
