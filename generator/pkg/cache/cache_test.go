package cache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLimitedCache(t *testing.T) {

	c := NewLimitedCache[int](3)

	require.ElementsMatch(t, c.Values(), []int{})

	c.Add(1)
	require.ElementsMatch(t, c.Values(), []int{1})

	c.Add(2)
	require.ElementsMatch(t, c.Values(), []int{1, 2})

	c.Add(3)
	require.ElementsMatch(t, c.Values(), []int{1, 2, 3})

	c.Add(4)
	require.ElementsMatch(t, c.Values(), []int{2, 3, 4})

	c.Add(5)
	require.ElementsMatch(t, c.Values(), []int{3, 4, 5})

	c.Add(6)
	require.ElementsMatch(t, c.Values(), []int{4, 5, 6})

	c.Add(7)
	require.ElementsMatch(t, c.Values(), []int{5, 6, 7})
}
