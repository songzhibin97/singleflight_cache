# singleflight_cache

Single-flight caching for observability, using singleflight with some event burial points to provide some observability metrics (USE) on the cache


## quick start

```shell
go get github.com/songzhibin97/singleflight_cache
```


The SingleFlightCacheInterface interface needs to be implemented beforehand.

```go
package mock

import (
	"context"
	"errors"
	"fmt"
	"strings"
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


```
The following interfaces are provided

```go
package mock

import "context"


type ObservationIndications interface {
	Close()

	AddErrorCount()
	AddHitCount()
	AddMissCount()
	AddManualMissCount()
	AddHitTimeConsuming(timeConsuming int64)
	AddMissTimeConsuming(timeConsuming int64)

	GetCycleErrorCount() int64
	GetCycleHitCount() int64
	GetCycleMissCount() int64
	GetCycleHitTimeConsuming() int64
	GetCycleMissTimeConsuming() int64
	GetCycleManualMissCount() int64
	GetCycleHitRate() float64
	GetCycleErrorRate() float64
	GetCycleHitAvgTimeConsuming() float64
	GetCycleMissAvgTimeConsuming() float64

	GetLastErrorCount() int64
	GetLastHitCount() int64
	GetLastMissCount() int64
	GetLastHitTimeConsuming() int64
	GetLastMissTimeConsuming() int64
	GetLastManualMissCount() int64
	GetLastHitRate() float64
	GetLastErrorRate() float64
	GetLastHitAvgTimeConsuming() float64
	GetLastMissAvgTimeConsuming() float64
}

type SingleFlightCache[R, V any] interface {
	ObservationIndications() ObservationIndications

	InvalidateCache(r R) error

	Do(ctx context.Context, r R) (V, error)
}

```

Simple test!

```go

package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	singleflight_cache "github.com/songzhibin97/singleflight_cache"
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


func main() {
	cache, err := singleflight_cache.NewSingleFlightCache[string, string](context.Background(), time.Second*10, 1, mockCache[string, string]{
		"surprise!",
	})
	if err != nil {
		panic(err)
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
	fmt.Println("ErrorCount:", cache.ObservationIndications().GetLastErrorCount())       // ErrorCount: 2
	fmt.Println("MissCount:", cache.ObservationIndications().GetLastMissCount())         // MissCount: 17
	fmt.Println("HitCount", cache.ObservationIndications().GetLastHitCount())            // HitCount 81
	fmt.Println("HitRate:", cache.ObservationIndications().GetLastHitRate())             // HitRate: 0.81
	fmt.Println("ErrorRate:", cache.ObservationIndications().GetLastErrorRate())         // ErrorRate: 0.02
	fmt.Println("HitAvg", cache.ObservationIndications().GetLastHitAvgTimeConsuming())   // HitAvg 1.0011236630123457e+09
	fmt.Println("MissAvg", cache.ObservationIndications().GetLastMissAvgTimeConsuming()) // MissAvg 4.0023453824117646e+09
}

```