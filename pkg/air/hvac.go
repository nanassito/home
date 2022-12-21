package air

import (
	"math"
	"strconv"
	"time"

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

func hvacTempRefresher(hvac *air_proto.Hvac) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		temp, err := strconv.ParseFloat(string(payload), 64)
		if err != nil {
			logger.Printf("Fail| Can't parse temperature for %s (%v): %v\n", hvac.Name, string(payload), err)
			return
		}
		if hvac.ReportedState.Temperature != temp {
			logger.Printf("Info| %s reports a target of %.2f°C\n", hvac.Name, temp)
		}
		hvac.ReportedState.Temperature = temp
	}
}

func hvacModeRefresher(hvac *air_proto.Hvac) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		value := string(payload)
		mode, ok := Str2Mode[value]
		if !ok {
			logger.Printf("Fail| %s reports unknown mode `%s`\n", hvac.Name, value)
			return
		}
		if hvac.ReportedState.Mode != mode {
			logger.Printf("Info| %s reports mode is %s\n", hvac.Name, mode.String())
		}
		hvac.ReportedState.Mode = mode
	}
}

func hvacFanRefresher(hvac *air_proto.Hvac) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		value := string(payload)
		fan, ok := Str2Fan[value]
		if !ok {
			logger.Printf("Fail| %s reports unknown fan `%s`\n", hvac.Name, value)
			return
		}
		if hvac.ReportedState.Fan != fan {
			logger.Printf("Info| %s reports fan is %s\n", hvac.Name, fan.String())
		}
		hvac.ReportedState.Fan = fan
	}
}

func NewHvac(name string, cfg *air_proto.AirConfig_Hvac, mqttClient mqtt.MqttIface) *air_proto.Hvac {
	hvac := air_proto.Hvac{
		Name:              name,
		Control:           air_proto.Hvac_CONTROL_ROOM,
		ReportedState:     &air_proto.Hvac_State{},
		DesiredState:      &air_proto.Hvac_State{},
		TemperatureOffset: new(float64),
	}

	if *initFromProm {
		if lastControl, ok := LastRunHvacControls[name]; ok {
			hvac.Control = lastControl
		}
		lastOffset, ok1 := LastRunHvacOffsets[name]
		lastTemp, ok2 := LastRunHvacDesiredTemp[name]
		if ok1 && ok2 {
			hvac.TemperatureOffset = &lastOffset
			hvac.DesiredState.Temperature = lastTemp
		}
	}

	err := mqttClient.Subscribe(cfg.ReportTemperatureMqttTopic, hvacTempRefresher(&hvac))
	if err != nil {
		panic(err)
	}

	err = mqttClient.Subscribe(cfg.ReportModeMqttTopic, hvacModeRefresher(&hvac))
	if err != nil {
		panic(err)
	}

	err = mqttClient.Subscribe(cfg.ReportFanMqttTopic, hvacFanRefresher(&hvac))
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			HvacControl(&hvac, cfg, mqttClient)
		}
	}()

	return &hvac
}

func HvacControl(state *air_proto.Hvac, config *air_proto.AirConfig_Hvac, mqttClient mqtt.MqttIface) {
	if state.ReportedState.Mode != state.DesiredState.Mode && state.DesiredState.Mode != air_proto.Hvac_MODE_UNKNOWN {
		mode := state.DesiredState.Mode.String()[5:]
		if !*readonly {
			logger.Printf("Info| Setting %s's mode to %s\n", state.Name, mode)
			mqttClient.PublishString(config.SetModeMqttTopic, mode)
		}
		return
	}
	if state.ReportedState.Mode == air_proto.Hvac_MODE_OFF {
		return
	}
	if state.ReportedState.Fan != state.DesiredState.Fan && state.DesiredState.Fan != air_proto.Hvac_FAN_UNKNOWN {
		fan := state.DesiredState.Fan.String()[4:]
		if !*readonly {
			logger.Printf("Info| Setting %s's fan to %s\n", state.Name, fan)
			mqttClient.PublishString(config.SetFanMqttTopic, fan)
		}
		return
	}
	desiredTemperature := math.Round((state.DesiredState.Temperature+*state.TemperatureOffset)*2) / 2
	if state.ReportedState.Temperature != desiredTemperature && desiredTemperature != 0 {
		temp := strconv.FormatFloat(desiredTemperature, 'f', 2, 64)
		if !*readonly {
			logger.Printf("Info| Setting %s's temperature to %.2f+%.2f=%s\n", state.Name, state.DesiredState.Temperature, *state.TemperatureOffset, temp)
			mqttClient.PublishString(config.SetTemperatureMqttTopic, temp)
		}
		return
	}
}
