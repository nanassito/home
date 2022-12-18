package air

import (
	"bytes"
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
	promHvacOffsetTemp = prometheus.NewDesc(
		"air_hvac_offset_temperature",
		"Offset to apply to target temperature for the Hvac unit.",
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
	metricHvacControl = "air_hvac_control"
	promHvacControl   = prometheus.NewDesc(
		metricHvacControl,
		"What controls each Hvac.",
		[]string{"room", "hvac", "control"},
		nil,
	)
)

func (p *PromCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- promRoomDesiredMin
	ch <- promRoomDesiredMax
	ch <- promHvacDesiredTemp
	ch <- promHvacOffsetTemp
	ch <- promHvacReportedTemp
	ch <- promHvacDesiredMode
	ch <- promHvacReportedMode
	ch <- promHvacDesiredFan
	ch <- promHvacReportedFan
	ch <- promHvacControl
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
		ch <- prometheus.MustNewConstMetric(promRoomDesiredMin, prometheus.GaugeValue, room.DesiredTemperatureRange.Min, room.Name)
		ch <- prometheus.MustNewConstMetric(promRoomDesiredMax, prometheus.GaugeValue, room.DesiredTemperatureRange.Max, room.Name)

	}
	for _, hvac := range p.Server.State.Hvacs {
		hvacCfg, ok := p.Server.Config.Hvacs[hvac.Name]
		if !ok {
			panic(fmt.Sprintf("Can't find config for hvac %v", hvac.Name))
		}
		roomName := hvacCfg.Room

		ch <- prometheus.MustNewConstMetric(promHvacDesiredTemp, prometheus.GaugeValue, hvac.DesiredState.Temperature, roomName, hvac.Name)
		ch <- prometheus.MustNewConstMetric(promHvacOffsetTemp, prometheus.GaugeValue, hvac.TemperatureOffset, roomName, hvac.Name)
		ch <- prometheus.MustNewConstMetric(promHvacReportedTemp, prometheus.GaugeValue, hvac.ReportedState.Temperature, roomName, hvac.Name)
		for _, mode := range air_proto.Hvac_Mode_name {
			ch <- prometheus.MustNewConstMetric(promHvacDesiredMode, prometheus.GaugeValue, b2f(mode == hvac.DesiredState.Mode.String()), roomName, hvac.Name, mode[5:])
			ch <- prometheus.MustNewConstMetric(promHvacReportedMode, prometheus.GaugeValue, b2f(mode == hvac.ReportedState.Mode.String()), roomName, hvac.Name, mode[5:])
		}
		for _, fan := range air_proto.Hvac_Fan_name {
			ch <- prometheus.MustNewConstMetric(promHvacDesiredFan, prometheus.GaugeValue, b2f(fan == hvac.DesiredState.Fan.String()), roomName, hvac.Name, fan[4:])
			ch <- prometheus.MustNewConstMetric(promHvacReportedFan, prometheus.GaugeValue, b2f(fan == hvac.ReportedState.Fan.String()), roomName, hvac.Name, fan[4:])
		}
		for _, control := range air_proto.Hvac_Control_name {
			ch <- prometheus.MustNewConstMetric(promHvacControl, prometheus.GaugeValue, b2f(control == hvac.Control.String()), roomName, hvac.Name, control[8:])
		}
	}
}

func getLastDesiredRoomTemperatures(metric string) map[string]float64 {
	resp := map[string]float64{}

	results, err := prom.Query(fmt.Sprintf("last_over_time(%s[1w])", metric), "initRoomDesiredTemp")
	if err != nil {
		logger.Printf("Fail| Failed to initialize desired room temperatures: %v", err)
		return resp
	}

	rows := results.(model.Vector)
	for _, row := range rows {
		resp[string(row.Metric["room"])] = float64(row.Value)
	}
	return resp
}

func getLastHvacControls() map[string]air_proto.Hvac_Control {
	resp := map[string]air_proto.Hvac_Control{}

	results, err := prom.Query(fmt.Sprintf("last_over_time(%s[1w]) > 0", metricHvacControl), "initHvacControl")
	if err != nil {
		logger.Printf("Fail| Failed to initialize Hvac controls: %v", err)
		return resp
	}

	rows := results.(model.Vector)
	for _, row := range rows {
		resp[string(row.Metric["hvac"])] = air_proto.Hvac_Control(air_proto.Hvac_Control_value["CONTROL_"+string(row.Metric["control"])])
	}
	return resp
}

var (
	LastRunDesiredMinimalRoomTemperatures = map[string]float64{}
	LastRunDesiredMaximalRoomTemperatures = map[string]float64{}
	LastRunHvacControls                   = map[string]air_proto.Hvac_Control{}
)

func init() {
	LastRunDesiredMaximalRoomTemperatures = getLastDesiredRoomTemperatures(metricRoomDesiredMax)
	LastRunDesiredMinimalRoomTemperatures = getLastDesiredRoomTemperatures(metricRoomDesiredMin)
	LastRunHvacControls = getLastHvacControls()
}

func promLabelsAsFilter(labels map[string]string) string {
	promFilter := new(bytes.Buffer)
	for key, value := range labels {
		fmt.Fprintf(promFilter, "%s=\"%s\", ", key, value)
	}
	return promFilter.String()
}
