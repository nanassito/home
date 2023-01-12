package air

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/mqtt"
	"github.com/nanassito/home/pkg/prom"
)

const (
	maxSensorStaleness = 1 * time.Hour
)

type sensorMqttPayload struct {
	Temperature float64 `json:"temperature"`
}

func sensorRefresher(sensor *air_proto.Sensor) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		parsed := sensorMqttPayload{}
		err := json.Unmarshal(payload, &parsed)
		if err != nil {
			logger.Printf("Fail| Failed to parse mqtt message from %s (%s): %v\n", topic, string(payload), err)
			return
		}
		if math.Abs(sensor.Temperature-parsed.Temperature) > 0.25 {
			logger.Printf("Info| %s sensor reports %.2f°C\n", sensor.Name, parsed.Temperature)
		}
		sensor.Temperature = parsed.Temperature
		sensor.LastReportedAtTs = time.Now().Unix()
	}
}

func IsSensorAlive(s *air_proto.Sensor) bool {
	last := time.Unix(s.LastReportedAtTs, 0)
	return last.Add(maxSensorStaleness).After(time.Now())
}

func NewSensor(name string, cfg *air_proto.AirConfig_Sensor, mqttClient mqtt.MqttIface) *air_proto.Sensor {
	var err error
	sensor := air_proto.Sensor{Name: name}
	sensor.Temperature, err = prom.QueryOne(
		fmt.Sprintf("last_over_time(mqtt_temperature{%s}[1h])", promLabelsAsFilter(cfg.PrometheusLabels)),
		"init-sensor-temp-"+name,
	)
	if err != nil {
		logger.Printf("Fail| Can't init sensor(temp) %v from Prometheus: %v\n", name, err)
	}
	lastReportedAtTs, err := prom.QueryOne(
		fmt.Sprintf("max_over_time((group(increase(mqtt_temperature{%s}[5m]) > 0) * time())[1w:5m])", promLabelsAsFilter(cfg.PrometheusLabels)),
		"init-sensor-time-"+name,
	)
	if err == nil {
		sensor.LastReportedAtTs = int64(lastReportedAtTs)
	} else {
		logger.Printf("Fail| Can't init sensor(time) %v from Prometheus: %v\n", name, err)
	}
	err = mqttClient.Subscribe(cfg.MqttTopic, sensorRefresher(&sensor))
	if err != nil {
		panic(err)
	}
	return &sensor
}
