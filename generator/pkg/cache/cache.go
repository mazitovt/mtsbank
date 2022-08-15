package cache

import "sync"

type Cache[T any] interface {
	Put(v T)
	Fill([]T) []T
}

var _ Cache[any] = (*LimitedCache[any])(nil)

type LimitedCache[T any] struct {
	mu    sync.RWMutex
	index int
	s     []T
}

func NewLimitedCache[T any](limit uint64) *LimitedCache[T] {
	return &LimitedCache[T]{
		index: 0,
		s:     make([]T, 0, limit),
	}
}

func (l *LimitedCache[T]) Put(v T) {
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

func (l *LimitedCache[T]) Fill(out []T) []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if cap(out) < len(l.s) {
		return out
	}

	out = out[:len(l.s)]
	copy(out[:len(out)-l.index], l.s[l.index:])
	copy(out[len(out)-l.index:], l.s[:l.index])

	return out
}
