package cache

import "sync"

type Cache[T any] interface {
	Add(v T)
	Values() []T
}

var _ Cache[any] = (*LimitedCache[any])(nil)

type LimitedCache[T any] struct {
	mu    sync.RWMutex
	index int
	s     []T
}

func NewLimitedCache[T any](size uint64) *LimitedCache[T] {
	return &LimitedCache[T]{
		index: 0,
		s:     make([]T, 0, size),
	}
}

func (l *LimitedCache[T]) Add(v T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = append(l.s[:l.index], v)
	if len(l.s) != cap(l.s) {
		l.s = l.s[:l.index+1]
	}
	l.index++
	if l.index == cap(l.s) {
		l.index = 0
	}
}

func (l *LimitedCache[T]) Values() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	r := make([]T, len(l.s))
	copy(r[:len(r)-l.index], l.s[l.index:])
	copy(r[len(r)-l.index:], l.s[:l.index])

	return r
}
