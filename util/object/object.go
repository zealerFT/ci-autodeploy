package object

import (
	"autodeploy/util/prometheussvc"

	"github.com/gin-gonic/gin"
)

var (
	Metrics = K[*prometheussvc.MetricsList]("metrics")
)

type K[T any] string

func (k K[T]) Get(c *gin.Context) T {
	if v, exist := c.Get(string(k)); !exist {
		return *new(T)
	} else if typed, ok := v.(T); !ok {
		return *new(T)
	} else {
		return typed
	}
}
func (k K[T]) Set(value T, c *gin.Context) {
	c.Set(string(k), value)
}
func P[T1, T2 any](f func(a T1, b T2), arg1 T1) func(T2) {
	return func(t2 T2) {
		f(arg1, t2)
	}
}
