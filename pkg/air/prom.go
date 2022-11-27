package air

import (
	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/prom"
	"github.com/prometheus/common/model"
)

func RegisterGetLast30mHvacsTemperature(s *Server, cfg *air_proto.AirConfig) {
	matchers := make(map[string]func(model.LabelSet) bool, len(cfg.Rooms)+1)
	for _, roomCfg := range cfg.Rooms {
		for hvacName, hvacCfg := range roomCfg.Hvacs {
			matchers[hvacName] = func(labels model.LabelSet) bool {
				for k, v := range hvacCfg.PrometheusLabels {
					if labels[model.LabelName(k)] != model.LabelValue(v) {
						return false
					}
				}
				return true
			}
		}
	}
	s.GetLast30mHvacsTemperature = func() (map[string]float64, error) {
		resp := make(map[string]float64, len(s.State.Rooms)+1)

		results, err := prom.Query("avg_over_time(mqtt_target_temperature_low_state[30m])", "last30mHvacsTemps")
		if err != nil {
			return resp, err
		}

		rows := results.(model.Vector)
		for _, row := range rows {
			for hvacName, matcher := range matchers {
				if matcher(model.LabelSet(row.Metric)) {
					resp[hvacName] = float64(row.Value)
				}
			}
		}
		return resp, nil
	}
}

func RegisterGet30mRoomTemperatureDeltas(s *Server, cfg *air_proto.AirConfig) {
	matchers := make(map[string]func(model.LabelSet) bool, len(cfg.Rooms))
	for roomName, roomCfg := range cfg.Rooms {
		matchers[roomName] = func(labels model.LabelSet) bool {
			for k, v := range roomCfg.Sensor.PrometheusLabels {
				if labels[model.LabelName(k)] != model.LabelValue(v) {
					return false
				}
			}
			return true
		}
	}
	s.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
		resp := make(map[string]float64, len(s.State.Rooms)+1)

		results, err := prom.Query("delta(mett_temperature{type=\"air\"}[30m])", "30mTempDeltas")
		if err != nil {
			return resp, err
		}

		rows := results.(model.Vector)
		for _, row := range rows {
			for hvacName, matcher := range matchers {
				if matcher(model.LabelSet(row.Metric)) {
					resp[hvacName] = float64(row.Value)
				}
			}
		}
		return resp, nil
	}
}
