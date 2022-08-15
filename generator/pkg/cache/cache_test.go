package cache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLimitedCache1(t *testing.T) {

	c := NewLimitedCache[int](3)
	out := make([]int, 0, 5)
	c.Fill(out)
	require.ElementsMatch(t, c.Fill(out), []int{})

	c.Put(1)
	require.ElementsMatch(t, c.Fill(out), []int{1})

	c.Put(2)
	require.ElementsMatch(t, c.Fill(out), []int{1, 2})

	c.Put(3)
	require.ElementsMatch(t, c.Fill(out), []int{1, 2, 3})

	c.Put(4)
	require.ElementsMatch(t, c.Fill(out), []int{2, 3, 4})

	c.Put(5)
	require.ElementsMatch(t, c.Fill(out), []int{3, 4, 5})

	c.Put(6)
	require.ElementsMatch(t, c.Fill(out), []int{4, 5, 6})

	c.Put(7)
	require.ElementsMatch(t, c.Fill(out), []int{5, 6, 7})
}

func TestLimitedCache2(t *testing.T) {

	c := NewLimitedCache[int](3)
	out := make([]int, 0, 2)
	out = c.Fill(out)
	require.ElementsMatch(t, out, []int{})

	c.Put(1)
	out = c.Fill(out)
	require.ElementsMatch(t, out, []int{1})

	c.Put(2)
	out = c.Fill(out)
	require.ElementsMatch(t, out, []int{1, 2})

	c.Put(3)
	out = c.Fill(out)
	require.ElementsMatch(t, out, []int{1, 2})
}
