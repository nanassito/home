package air

import (
	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/mqtt"
)

func NewRoom(name string, cfg *air_proto.AirConfig_Sensor, mqttClient mqtt.MqttIface) *air_proto.Room {
	desiredMin := float64(19)
	if last, ok := LastRunDesiredMinimalRoomTemperatures[name]; ok && *initFromProm {
		desiredMin = last
	}
	desiredMax := float64(33)
	if last, ok := LastRunDesiredMaximalRoomTemperatures[name]; ok && *initFromProm {
		desiredMax = last
	}

	room := air_proto.Room{
		Name:   name,
		Sensor: NewSensor(name, cfg, mqttClient),
		DesiredTemperatureRange: &air_proto.TemperatureRange{
			Min: desiredMin,
			Max: desiredMax,
		},
	}

	return &room
}
