package singleflight_cache

import (
	"context"
	"time"

	"github.com/songzhibin97/go-baseutils/sys/nanotime"
	"golang.org/x/sync/singleflight"
)

type SingleFlightCache[R, V any] interface {
	ObservationIndications() ObservationIndications

	InvalidateCache(r R) error

	Do(ctx context.Context, r R) (V, error)
}

type SingleFlightCacheInterface[R, V any] interface {
	InvalidateCache(r R) error

	LoadCacheData(ctx context.Context, r R) (V, error)

	LoadSourceData(ctx context.Context, r R) (V, error)

	LoadUniqueIdentity(r R) string
}

type singleFlightCache[R, V any] struct {
	group                      singleflight.Group
	observationIndications     ObservationIndications
	singleFlightCacheInterface SingleFlightCacheInterface[R, V]
}

func (s *singleFlightCache[R, V]) ObservationIndications() ObservationIndications {
	return s.observationIndications
}

func (s *singleFlightCache[R, V]) InvalidateCache(r R) error {

	err := s.singleFlightCacheInterface.InvalidateCache(r)
	if err != nil {
		return err
	}
	s.observationIndications.AddManualMissCount()
	return nil
}

func (s *singleFlightCache[R, V]) Do(ctx context.Context, r R) (V, error) {
	cost := nanotime.RuntimeNanotime()
	vs, err, _ := s.group.Do(s.singleFlightCacheInterface.LoadUniqueIdentity(r), func() (interface{}, error) {
		nexCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		vs, err := s.singleFlightCacheInterface.LoadCacheData(nexCtx, r)
		if err == nil {
			s.observationIndications.AddHitCount()
			s.observationIndications.AddHitTimeConsuming(nanotime.RuntimeNanotime() - cost)
			return vs, nil
		}
		nexCtx, cancel = context.WithCancel(ctx)
		defer cancel()
		vs, err = s.singleFlightCacheInterface.LoadSourceData(nexCtx, r)
		if err == nil {
			s.observationIndications.AddMissCount()
			s.observationIndications.AddMissTimeConsuming(nanotime.RuntimeNanotime() - cost)
			return vs, err
		}
		return vs, err
	})
	if err != nil {
		s.observationIndications.AddErrorCount()
	}
	return vs.(V), err
}

func NewSingleFlightCache[R, V any](ctx context.Context, interval time.Duration, maxLen int64, singleFlightCacheInterface SingleFlightCacheInterface[R, V]) (SingleFlightCache[R, V], error) {
	observationIndications, err := NewObservationIndications(ctx, interval, maxLen)
	if err != nil {
		return nil, err
	}
	return &singleFlightCache[R, V]{
		observationIndications:     observationIndications,
		singleFlightCacheInterface: singleFlightCacheInterface,
	}, nil
}
