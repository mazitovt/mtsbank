package cache

type Cache[T any] interface {
	Add(at T)
	Values() []T
}

var _ Cache[any] = (*LimitedCache[any])(nil)

type LimitedCache[T any] struct {
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
	return l.s[:]
}
