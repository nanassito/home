package prom_test

import (
	"testing"
	"time"

	"github.com/matryer/is"
	prom "github.com/nanassito/home/pkg/prom"
)

func TestQueryOne(t *testing.T) {
	is := is.New(t)
	result, err := prom.QueryOne("time()", "time")
	is.NoErr(err)
	is.True(result < float64(time.Now().Unix())+10)
	is.True(float64(time.Now().Unix())-10 < result)
}
