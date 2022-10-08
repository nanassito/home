package prom

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"
)

var prom v1.API

var LoopRunsCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "home",
		Name:      "loop_runs_total",
		Help:      "Number of iteration of a loop",
	},
	[]string{"domain", "type", "instance"},
)

func init() {
	client, err := api.NewClient(api.Config{
		Address: "http://192.168.1.1:9090",
	})
	if err != nil {
		panic(fmt.Sprintf("Error creating client: %v\n", err))
	}
	prom = v1.NewAPI(client)
}

func Query(promql string, logID string) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := prom.Query(ctx, promql, time.Now())
	if err != nil {
		fmt.Printf("Error | %s | %v\n", logID, err.Error())
		return nil, err
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings| %s | %v\n", logID, warnings)
	}
	return result, nil
}

func QueryOne(promql string, logID string) (float64, error) {
	result, err := Query(promql, logID)
	if err != nil {
		return 0, err
	}
	switch {
	case result.Type() == model.ValScalar:
		return float64(result.(*model.Scalar).Value), nil
	case result.Type() == model.ValVector:
		rs := result.(model.Vector)
		if len(rs) != 1 {
			return 0, fmt.Errorf("%s returned %d results instead of 1", logID, len(rs))
		}
		return float64(rs[0].Value), nil
	default:
		return 0, fmt.Errorf("%s returned a %v instead of %v", logID, result.Type(), model.ValScalar)
	}
}
