package singleflight_cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockCache[R string, V any] struct {
	v V
}

func (m mockCache[R, V]) InvalidateCache(r R) error {
	fmt.Println("InvalidateCache:", r)
	return nil
}

func (m mockCache[R, V]) LoadCacheData(ctx context.Context, r R) (V, error) {
	time.Sleep(time.Second)
	//fmt.Println("LoadCacheData:", key)
	if strings.Contains(string(r), "1") {
		return m.v, errors.New("LoadCacheData Error")
	}
	return m.v, nil
}

func (m mockCache[R, V]) LoadSourceData(ctx context.Context, r R) (V, error) {
	time.Sleep(3 * time.Second)
	//fmt.Println("LoadSourceData:", key)
	if strings.Contains(string(r), "2") {
		return m.v, errors.New("LoadSourceData Error")
	}
	return m.v, nil
}

func (m mockCache[R, V]) LoadUniqueIdentity(r R) string {
	return string("mock_" + r)
}

func TestNewSingleFlightCache(t *testing.T) {
	cache, err := NewSingleFlightCache[string, string](context.Background(), time.Second*10, 1, mockCache[string, string]{
		"surprise!",
	})
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err = cache.Do(context.Background(), strconv.Itoa(i))
			if err != nil {
				//t.Log("Err index", i)
			}
			//t.Log(v)
		}(i)
	}
	wg.Wait()
	t.Log(">>>>>")
	t.Log("ErrorCount:", cache.ObservationIndications().GetLastErrorCount())
	t.Log("MissCount:", cache.ObservationIndications().GetLastMissCount())
	t.Log("HitCount", cache.ObservationIndications().GetLastHitCount())
	t.Log("HitRate:", cache.ObservationIndications().GetLastHitRate())
	t.Log("ErrorRate:", cache.ObservationIndications().GetLastErrorRate())
	t.Log("HitAvg", cache.ObservationIndications().GetLastHitAvgTimeConsuming())
	t.Log("MissAvg", cache.ObservationIndications().GetLastMissAvgTimeConsuming())
}
