package cache

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLimitedCache(t *testing.T) {

	c := NewLimitedCache[int](3)

	c.Add(1)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{1})

	c.Add(2)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{1, 2})

	c.Add(3)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{1, 2, 3})

	c.Add(4)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{4, 2, 3})

	c.Add(5)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{4, 5, 3})

	c.Add(6)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{4, 5, 6})

	c.Add(7)
	fmt.Println(c.Values())
	require.ElementsMatch(t, c.Values(), []int{7, 5, 6})
}
