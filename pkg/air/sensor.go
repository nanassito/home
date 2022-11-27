package air

import (
	"encoding/json"
	"math"
	"time"

	"github.com/nanassito/home/pkg/air_proto"
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
			logger.Printf("Info| %s sensor reports %.2f°C\n", sensor.Location, parsed.Temperature)
		}
		sensor.Temperature = parsed.Temperature
		sensor.LastReportedAtTs = time.Now().Unix()
	}
}

func IsSensorAlive(s *air_proto.Sensor) bool {
	last := time.Unix(s.LastReportedAtTs, 0)
	return last.Add(maxSensorStaleness).After(time.Now())
}
