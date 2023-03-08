package singleflight_cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

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

// NewObservationIndications interval * maxLen  is the total time of the ring
func NewObservationIndications(ctx context.Context, interval time.Duration, maxLen int64) (ObservationIndications, error) {
	if interval == 0 || maxLen == 0 {
		return nil, errors.New("interval or maxLen is zero")
	}

	r := &ring{
		data:     make([]*indications, maxLen),
		maxLen:   maxLen,
		summary:  &indications{},
		interval: interval,
		idx:      0,
	}
	for idx := range r.data {
		r.data[idx] = &indications{}
	}
	r.ctx, r.cancel = context.WithCancel(ctx)
	go r.run()
	return r, nil
}

type ring struct {
	data   []*indications
	idx    int64
	maxLen int64

	summary  *indications
	interval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	close  atomic.Bool
}

func (r *ring) Close() {
	if r.close.Load() {
		return
	}
	r.close.Store(true)
	r.cancel()
}

func (r *ring) GetCycleErrorCount() int64 {
	return r.summary.getErrorCount()
}

func (r *ring) GetCycleHitCount() int64 {
	return r.summary.getHitCount()
}

func (r *ring) GetCycleMissCount() int64 {
	return r.summary.getMissCount()
}

func (r *ring) GetCycleHitTimeConsuming() int64 {
	return r.summary.getHitTimeConsuming()
}

func (r *ring) GetCycleMissTimeConsuming() int64 {
	return r.summary.getMissTimeConsuming()
}

func (r *ring) GetCycleManualMissCount() int64 {
	return r.summary.getManualMissCount()
}

func (r *ring) GetCycleHitRate() float64 {
	return r.summary.getHitRate()
}

func (r *ring) GetCycleErrorRate() float64 {
	return r.summary.getErrorRate()
}

func (r *ring) GetCycleHitAvgTimeConsuming() float64 {
	return r.summary.getHitAvgTimeConsuming()
}

func (r *ring) GetCycleMissAvgTimeConsuming() float64 {
	return r.summary.getMissAvgTimeConsuming()
}

func (r *ring) GetLastErrorCount() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getErrorCount()
}

func (r *ring) GetLastHitCount() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getHitCount()
}

func (r *ring) GetLastMissCount() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getMissCount()
}

func (r *ring) GetLastHitTimeConsuming() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getHitTimeConsuming()
}

func (r *ring) GetLastMissTimeConsuming() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getMissTimeConsuming()
}

func (r *ring) GetLastManualMissCount() int64 {
	return r.data[atomic.LoadInt64(&r.idx)].getManualMissCount()
}

func (r *ring) GetLastHitRate() float64 {
	return r.data[atomic.LoadInt64(&r.idx)].getHitRate()
}

func (r *ring) GetLastErrorRate() float64 {
	return r.data[atomic.LoadInt64(&r.idx)].getErrorRate()
}

func (r *ring) GetLastHitAvgTimeConsuming() float64 {
	return r.data[atomic.LoadInt64(&r.idx)].getHitAvgTimeConsuming()
}

func (r *ring) GetLastMissAvgTimeConsuming() float64 {
	return r.data[atomic.LoadInt64(&r.idx)].getMissAvgTimeConsuming()
}

func (r *ring) AddErrorCount() {
	r.data[atomic.LoadInt64(&r.idx)].addErrorCount()
}

func (r *ring) AddHitCount() {
	r.data[atomic.LoadInt64(&r.idx)].addHitCount()
}

func (r *ring) AddMissCount() {
	r.data[atomic.LoadInt64(&r.idx)].addMissCount()
}

func (r *ring) AddManualMissCount() {
	r.data[atomic.LoadInt64(&r.idx)].addManualMissCount()
}

func (r *ring) AddHitTimeConsuming(timeConsuming int64) {
	r.data[atomic.LoadInt64(&r.idx)].addHitTimeConsuming(timeConsuming)
}

func (r *ring) AddMissTimeConsuming(timeConsuming int64) {
	r.data[atomic.LoadInt64(&r.idx)].addMissTimeConsuming(timeConsuming)
}

func (r *ring) run() {
	timer := time.NewTicker(r.interval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			idx := atomic.LoadInt64(&r.idx)
			pre := (idx - 1 + r.maxLen) % r.maxLen
			r.summary.subtractionFiltering(r.data[pre])
			r.summary.addFiltering(r.data[idx])

			next := (idx + 1) % r.maxLen
			r.data[next].reset()

			atomic.StoreInt64(&r.idx, next)

		case <-r.ctx.Done():
			return
		}
	}
}

type indications struct {
	errorCount       int64
	hitCount         int64
	hitTimeConsuming int64

	missCount         int64
	missTimeConsuming int64

	manualMissCount int64
}

func (i *indications) addErrorCount() {
	atomic.AddInt64(&i.errorCount, 1)
}

func (i *indications) addHitCount() {
	atomic.AddInt64(&i.hitCount, 1)
}

func (i *indications) addMissCount() {
	atomic.AddInt64(&i.missCount, 1)
}

func (i *indications) addManualMissCount() {
	atomic.AddInt64(&i.manualMissCount, 1)
}

func (i *indications) addHitTimeConsuming(hitTimeConsuming int64) {
	atomic.AddInt64(&i.hitTimeConsuming, hitTimeConsuming)
}

func (i *indications) addMissTimeConsuming(missTimeConsuming int64) {
	atomic.AddInt64(&i.missTimeConsuming, missTimeConsuming)
}

func (i *indications) getErrorCount() int64 {
	return atomic.LoadInt64(&i.errorCount)
}

func (i *indications) getHitCount() int64 {
	return atomic.LoadInt64(&i.hitCount)
}

func (i *indications) getMissCount() int64 {
	return atomic.LoadInt64(&i.missCount)
}

func (i *indications) getHitTimeConsuming() int64 {
	return atomic.LoadInt64(&i.hitTimeConsuming)
}

func (i *indications) getMissTimeConsuming() int64 {
	return atomic.LoadInt64(&i.missTimeConsuming)
}

func (i *indications) getManualMissCount() int64 {
	return atomic.LoadInt64(&i.manualMissCount)
}

func (i *indications) getHitRate() float64 {
	total := i.getHitCount() + i.getMissCount() + i.getErrorCount() - i.getManualMissCount()
	if total == 0 {
		return 0
	}
	return float64(i.getHitCount()) / float64(total)
}

func (i *indications) getErrorRate() float64 {
	total := i.getHitCount() + i.getMissCount() + i.getErrorCount() - i.getManualMissCount()
	if total == 0 {
		return 0
	}
	return float64(i.getErrorCount()) / float64(total)
}

func (i *indications) getHitAvgTimeConsuming() float64 {
	count := i.getHitCount()
	if count == 0 {
		return 0
	}
	return float64(i.getHitTimeConsuming()) / float64(count)
}

func (i *indications) getMissAvgTimeConsuming() float64 {
	count := i.getMissCount()
	if count == 0 {
		return 0
	}
	return float64(i.getMissTimeConsuming()) / float64(count)
}

func (i *indications) subtractionFiltering(old *indications) {
	atomic.AddInt64(&i.errorCount, -old.getErrorCount())
	atomic.AddInt64(&i.hitCount, -old.getHitCount())
	atomic.AddInt64(&i.missCount, -old.getMissCount())
	atomic.AddInt64(&i.hitTimeConsuming, -old.hitTimeConsuming)
	atomic.AddInt64(&i.missTimeConsuming, -old.missTimeConsuming)
	atomic.AddInt64(&i.manualMissCount, -old.manualMissCount)
}

func (i *indications) addFiltering(new *indications) {
	atomic.AddInt64(&i.errorCount, new.getErrorCount())
	atomic.AddInt64(&i.hitCount, new.getHitCount())
	atomic.AddInt64(&i.missCount, new.getMissCount())
	atomic.AddInt64(&i.hitTimeConsuming, new.hitTimeConsuming)
	atomic.AddInt64(&i.missTimeConsuming, new.missTimeConsuming)
	atomic.AddInt64(&i.manualMissCount, new.manualMissCount)
}

func (i *indications) reset() {
	atomic.StoreInt64(&i.errorCount, 0)
	atomic.StoreInt64(&i.hitCount, 0)
	atomic.StoreInt64(&i.missCount, 0)
	atomic.StoreInt64(&i.hitTimeConsuming, 0)
	atomic.StoreInt64(&i.missTimeConsuming, 0)
	atomic.StoreInt64(&i.manualMissCount, 0)
}
