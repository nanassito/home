package air

import (
	"fmt"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/mqtt"
)

// TODO: Init room Sensor data from Prometheus

func NewRoom(name string, cfg *air_proto.AirConfig_Room, mqttClient *mqtt.Mqtt) *air_proto.RoomState {
	sensor := air_proto.Sensor{Location: fmt.Sprintf("room-%s", name)}
	err := mqttClient.Subscribe(cfg.Sensor.MqttTopic, sensorRefresher(&sensor))
	if err != nil {
		panic(err)
	}
	hvacs := make(map[string]*air_proto.Hvac, len(cfg.Hvacs))
	for hvacName, hvacCfg := range cfg.Hvacs {
		hvacs[hvacName] = NewHvac(hvacName, hvacCfg, mqttClient)
	}
	desiredMin := float64(19)
	if last, ok := LastRunDesiredMinimalRoomTemperatures[name]; ok {
		desiredMin = last
	}
	desiredMax := float64(33)
	if last, ok := LastRunDesiredMinimalRoomTemperatures[name]; ok {
		desiredMax = last
	}
	return &air_proto.RoomState{
		RoomName: name,
		DesiredTemperatureRange: &air_proto.TemperatureRange{
			Min: desiredMin,
			Max: desiredMax,
		},
		Sensor: &sensor,
		Hvacs:  hvacs,
	}
}
