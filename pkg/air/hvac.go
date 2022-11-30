package air

import (
	"strconv"
	"sync"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/mqtt"
)

var (
	Str2Mode = map[string]air_proto.Hvac_Mode{
		"off":       air_proto.Hvac_MODE_OFF,
		"heat_cool": air_proto.Hvac_MODE_AUTO,
		"cool":      air_proto.Hvac_MODE_COOL,
		"heat":      air_proto.Hvac_MODE_HEAT,
		"fan_only":  air_proto.Hvac_MODE_FAN_ONLY,
		"dry":       air_proto.Hvac_MODE_DRY,
	}
	Mode2Str = map[air_proto.Hvac_Mode]string{
		air_proto.Hvac_MODE_OFF:      "off",
		air_proto.Hvac_MODE_AUTO:     "heat_cool",
		air_proto.Hvac_MODE_COOL:     "cool",
		air_proto.Hvac_MODE_HEAT:     "heat",
		air_proto.Hvac_MODE_FAN_ONLY: "fan_only",
		air_proto.Hvac_MODE_DRY:      "dry",
	}

	Str2Fan = map[string]air_proto.Hvac_Fan{
		"auto":   air_proto.Hvac_FAN_AUTO,
		"low":    air_proto.Hvac_FAN_LOW,
		"medium": air_proto.Hvac_FAN_MEDIUM,
		"high":   air_proto.Hvac_FAN_HIGH,
	}
	Fan2Str = map[air_proto.Hvac_Fan]string{
		air_proto.Hvac_FAN_AUTO:   "auto",
		air_proto.Hvac_FAN_LOW:    "low",
		air_proto.Hvac_FAN_MEDIUM: "medium",
		air_proto.Hvac_FAN_HIGH:   "high",
	}
)

func NewHvac(name string, cfg *air_proto.AirConfig_Room_Hvac, mqttClient *mqtt.Mqtt) *air_proto.Hvac {
	hvac := air_proto.Hvac{
		HvacName: name,
		// TODO: Needs to initialized to last reported metric.
		Control:       air_proto.Hvac_CONTROL_ROOM,
		ReportedState: &air_proto.Hvac_State{},
		DesiredState:  &air_proto.Hvac_State{},
	}

	initTemperature := sync.Once{}
	err := mqttClient.Subscribe(cfg.ReportTemperatureMqttTopic, func(topic string, payload []byte) {
		temp, err := strconv.ParseFloat(string(payload), 64)
		if err != nil {
			logger.Printf("Fail| Can't parse temperature for %s (%v): %v\n", name, string(payload), err)
			return
		}
		if hvac.ReportedState.Temperature != temp {
			logger.Printf("Info| %s reports %.2f°C\n", name, temp)
		}
		hvac.ReportedState.Temperature = temp
		initTemperature.Do(func() { hvac.DesiredState.Temperature = temp })
	})
	if err != nil {
		panic(err)
	}

	initMode := sync.Once{}
	err = mqttClient.Subscribe(cfg.ReportModeMqttTopic, func(topic string, payload []byte) {
		value := string(payload)
		mode, ok := Str2Mode[value]
		if !ok {
			logger.Printf("Fail| %s reports unknown mode `%s`\n", name, value)
			return
		}
		if hvac.ReportedState.Mode != mode {
			logger.Printf("Info| %s reports mode is %s\n", name, mode.String())
		}
		hvac.ReportedState.Mode = mode
		initMode.Do(func() { hvac.DesiredState.Mode = mode })
	})
	if err != nil {
		panic(err)
	}

	initFan := sync.Once{}
	err = mqttClient.Subscribe(cfg.ReportFanMqttTopic, func(topic string, payload []byte) {
		value := string(payload)
		fan, ok := Str2Fan[value]
		if !ok {
			logger.Printf("Fail| %s reports unknown fan `%s`\n", name, value)
			return
		}
		if hvac.ReportedState.Fan != fan {
			logger.Printf("Info| %s reports fan is %s\n", name, fan.String())
		}
		hvac.ReportedState.Fan = fan
		initFan.Do(func() { hvac.DesiredState.Fan = fan })
	})
	if err != nil {
		panic(err)
	}

	return &hvac
}
