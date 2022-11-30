package air

import (
	"fmt"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

type PromCollector struct {
	Server *Server
}

var (
	metricRoomDesiredMin = "air_room_desired_min_temperature"
	promRoomDesiredMin   = prometheus.NewDesc(
		metricRoomDesiredMin,
		"Minimum of the desired temperature range.",
		[]string{"room"},
		nil,
	)
	metricRoomDesiredMax = "air_room_desired_max_temperature"
	promRoomDesiredMax   = prometheus.NewDesc(
		metricRoomDesiredMax,
		"Maximum of the desired temperature range.",
		[]string{"room"},
		nil,
	)
	promHvacDesiredTemp = prometheus.NewDesc(
		"air_hvac_desired_temperature",
		"Desired target temperature for the Hvac unit.",
		[]string{"room", "hvac"},
		nil,
	)
	promHvacReportedTemp = prometheus.NewDesc(
		"air_hvac_reported_temperature",
		"Reported target temperature for the Hvac unit.",
		[]string{"room", "hvac"},
		nil,
	)
	promHvacDesiredMode = prometheus.NewDesc(
		"air_hvac_desired_mode",
		"Desired mode for the Hvac unit.",
		[]string{"room", "hvac", "mode"},
		nil,
	)
	promHvacReportedMode = prometheus.NewDesc(
		"air_hvac_reported_mode",
		"Reported mode for the Hvac unit.",
		[]string{"room", "hvac", "mode"},
		nil,
	)
	promHvacDesiredFan = prometheus.NewDesc(
		"air_hvac_desired_fan",
		"Desired fan for the Hvac unit.",
		[]string{"room", "hvac", "fan"},
		nil,
	)
	promHvacReportedFan = prometheus.NewDesc(
		"air_hvac_reported_fan",
		"Reported fan for the Hvac unit.",
		[]string{"room", "hvac", "fan"},
		nil,
	)
)

func (p *PromCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- promRoomDesiredMin
	ch <- promRoomDesiredMax
	ch <- promHvacDesiredTemp
	ch <- promHvacReportedTemp
	ch <- promHvacDesiredMode
	ch <- promHvacReportedMode
	ch <- promHvacDesiredFan
	ch <- promHvacReportedFan
}
func (p *PromCollector) Collect(ch chan<- prometheus.Metric) {
	b2f := func(b bool) float64 {
		if b {
			return 1
		} else {
			return 0
		}
	}
	for _, room := range p.Server.State.Rooms {
		ch <- prometheus.MustNewConstMetric(promRoomDesiredMin, prometheus.GaugeValue, room.DesiredTemperatureRange.Min, room.RoomName)
		ch <- prometheus.MustNewConstMetric(promRoomDesiredMax, prometheus.GaugeValue, room.DesiredTemperatureRange.Max, room.RoomName)
		for _, hvac := range room.Hvacs {
			ch <- prometheus.MustNewConstMetric(promHvacDesiredTemp, prometheus.GaugeValue, hvac.DesiredState.Temperature, room.RoomName, hvac.HvacName)
			ch <- prometheus.MustNewConstMetric(promHvacReportedTemp, prometheus.GaugeValue, hvac.ReportedState.Temperature, room.RoomName, hvac.HvacName)
			for _, mode := range air_proto.Hvac_Mode_name {
				ch <- prometheus.MustNewConstMetric(promHvacDesiredMode, prometheus.GaugeValue, b2f(mode == hvac.DesiredState.Mode.String()), room.RoomName, hvac.HvacName, mode[5:])
				ch <- prometheus.MustNewConstMetric(promHvacReportedMode, prometheus.GaugeValue, b2f(mode == hvac.ReportedState.Mode.String()), room.RoomName, hvac.HvacName, mode[5:])
			}
			for _, fan := range air_proto.Hvac_Fan_name {
				ch <- prometheus.MustNewConstMetric(promHvacDesiredFan, prometheus.GaugeValue, b2f(fan == hvac.DesiredState.Fan.String()), room.RoomName, hvac.HvacName, fan[4:])
				ch <- prometheus.MustNewConstMetric(promHvacReportedFan, prometheus.GaugeValue, b2f(fan == hvac.ReportedState.Fan.String()), room.RoomName, hvac.HvacName, fan[4:])
			}
		}
	}
}

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

func getLastDesiredRoomTemperatures(metric string) map[string]float64 {
	resp := make(map[string]float64, 4)

	results, err := prom.Query(fmt.Sprintf("last_over_time(%s[1w])", metric), "initRoomDesiredTemp")
	if err != nil {
		logger.Printf("Fail| Could not fetch the last minimal desired room temperatures: %v", err)
		return resp
	}

	rows := results.(model.Vector)
	for _, row := range rows {
		resp[string(row.Metric["room"])] = float64(row.Value)
	}
	return resp
}

var (
	LastRunDesiredMinimalRoomTemperatures = map[string]float64{}
	LastRunDesiredMaximalRoomTemperatures = map[string]float64{}
)

func init() {
	LastRunDesiredMaximalRoomTemperatures = getLastDesiredRoomTemperatures(metricRoomDesiredMax)
	LastRunDesiredMinimalRoomTemperatures = getLastDesiredRoomTemperatures(metricRoomDesiredMin)
}
