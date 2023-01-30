package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/nanassito/home/pkg/mqtt"
	"golang.org/x/time/rate"
)

var (
	quit = make(chan struct{})
)

type MqttBool struct {
	CommandTopic string
	StateTopic   string
	Value        bool
	IsInit       bool
}

type MqttString struct {
	CommandTopic string
	StateTopic   string
	Value        string
	IsInit       bool
}

type MqttFloat64 struct {
	CommandTopic string
	StateTopic   string
	Value        float64
	IsInit       bool
}

type Float64Range struct {
	Min *MqttFloat64
	Max *MqttFloat64
}

type Hvac struct {
	ModeState          *MqttString
	TemperatureState   *MqttFloat64
	TemperatureCommand string
	FanCommand         string
	Limiter            *rate.Limiter
}

type SerialMqtt struct {
	Limiter *rate.Limiter
	Client  mqtt.MqttIface
}

func (m *SerialMqtt) RateLimitedPublishString(topic string, payload string) {
	m.Limiter.Wait(context.Background())
	m.PublishString(topic, payload)

}

func (m *SerialMqtt) PublishString(topic string, payload string) {
	fmt.Printf("INFO | Sending `%s` on mqtt://`%s`\n", payload, topic)
	m.Client.PublishString(topic, payload)
}

var mqttClient = SerialMqtt{
	Limiter: rate.NewLimiter(rate.Every(1*time.Second), 1),
	Client:  mqtt.New("air2"),
}

type Room struct {
	Name                   string
	Enabled                *MqttBool
	Hvacs                  []*Hvac
	AcceptableTemperatures *Float64Range
	Sensors                []*MqttFloat64
	Mutex                  sync.Mutex
}

func (room *Room) IsInit() (bool, string) {
	if !room.Enabled.IsInit {
		return false, room.Enabled.CommandTopic
	}
	if !room.AcceptableTemperatures.Min.IsInit {
		return false, room.AcceptableTemperatures.Min.CommandTopic
	}
	if !room.AcceptableTemperatures.Max.IsInit {
		return false, room.AcceptableTemperatures.Max.CommandTopic
	}
	for _, h := range room.Hvacs {
		if !h.ModeState.IsInit {
			return false, h.ModeState.CommandTopic
		}
		if !h.TemperatureState.IsInit {
			return false, h.TemperatureState.CommandTopic
		}
	}
	for _, s := range room.Sensors {
		if !s.IsInit {
			return false, s.CommandTopic
		}
	}
	return true, ""
}

type temperatureJson struct {
	Temperature float64 `json:"temperature"`
}

func (room *Room) UpdateState(topic string, payload []byte) {
	switch topic {
	case room.Enabled.CommandTopic:
		value, err := strconv.ParseBool(string(payload))
		if err == nil {
			room.Enabled.Value = value
			room.Enabled.IsInit = true
			if room.Enabled.StateTopic != "" {
				mqttClient.PublishString(room.Enabled.StateTopic, strconv.FormatBool(room.Enabled.Value))
			}
		} else {
			fmt.Printf("ERROR | Can't parse message from mqtt://%s: %v\n", topic, err)
		}
	case room.AcceptableTemperatures.Min.CommandTopic:
		value, err := strconv.ParseFloat(string(payload), 64)
		if err == nil {
			room.AcceptableTemperatures.Min.Value = value
			room.AcceptableTemperatures.Min.IsInit = true
			if room.AcceptableTemperatures.Min.StateTopic != "" {
				mqttClient.PublishString(room.AcceptableTemperatures.Min.StateTopic, strconv.FormatFloat(room.AcceptableTemperatures.Min.Value, 'f', 1, 64))
			}
		} else {
			fmt.Printf("ERROR | Can't parse message from mqtt://%s: %v\n", topic, err)
		}
	case room.AcceptableTemperatures.Max.CommandTopic:
		value, err := strconv.ParseFloat(string(payload), 64)
		if err == nil {
			room.AcceptableTemperatures.Max.Value = value
			room.AcceptableTemperatures.Max.IsInit = true
			if room.AcceptableTemperatures.Max.StateTopic != "" {
				mqttClient.PublishString(room.AcceptableTemperatures.Max.StateTopic, strconv.FormatFloat(room.AcceptableTemperatures.Max.Value, 'f', 1, 64))
			}
		} else {
			fmt.Printf("ERROR | Can't parse message from mqtt://%s: %v\n", topic, err)
		}
	}
	for _, h := range room.Hvacs {
		switch topic {
		case h.ModeState.CommandTopic:
			h.ModeState.Value = string(payload)
			h.ModeState.IsInit = true
		case h.TemperatureState.CommandTopic:
			value, err := strconv.ParseFloat(string(payload), 64)
			if err == nil {
				h.TemperatureState.Value = value
				h.TemperatureState.IsInit = true
			} else {
				fmt.Printf("ERROR | Can't parse message from mqtt://%s: %v\n", topic, err)
			}
		}
	}
	for _, s := range room.Sensors {
		if topic == s.CommandTopic {
			var value float64
			parsed := temperatureJson{}
			err := json.Unmarshal(payload, &parsed)
			if err == nil {
				value = parsed.Temperature
			} else {
				value, err = strconv.ParseFloat(string(payload), 64)
			}
			if err == nil {
				s.Value = value
				s.IsInit = true
			} else {
				fmt.Printf("ERROR | Can't parse message from mqtt://%s: %v\n", topic, err)
			}
		}
	}
}

func (room *Room) GetCallback() func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		room.UpdateState(topic, payload)

		if ok, missing := room.IsInit(); !ok {
			fmt.Printf("WARN | Waiting for first message from mqtt://%s\n", missing)
			return
		}
		if !room.Enabled.Value {
			return
		}

		room.Mutex.Lock()
		defer room.Mutex.Unlock()

		for _, h := range room.Hvacs {
			hvac := h
			if hvac.ModeState.Value == "heat" {
				room.TuneHeat()
			}
		}
	}
}

func (room *Room) TuneHeat() {
	targetTemp := room.AcceptableTemperatures.Min.Value
	sensorMax := room.Sensors[0].Value
	sensorMin := room.Sensors[0].Value
	for _, s := range room.Sensors {
		sensorMax = math.Max(sensorMax, s.Value)
		sensorMin = math.Min(sensorMin, s.Value)
	}

	if sensorMax < targetTemp {
		for _, h := range room.Hvacs {
			h.Limiter.Wait(context.Background())
			oldTemp := h.TemperatureState.Value
			newTemp := oldTemp + 0.5
			// Only change when we are increasing the temperature
			if oldTemp < targetTemp+1 && targetTemp+1 <= newTemp {
				mqttClient.RateLimitedPublishString(h.FanCommand, "MEDIUM")
			}
			if oldTemp < targetTemp+2 && targetTemp+2 <= newTemp {
				mqttClient.RateLimitedPublishString(h.FanCommand, "HIGH")
			}
			mqttClient.RateLimitedPublishString(h.TemperatureCommand, strconv.FormatFloat(newTemp, 'f', 1, 64))
		}
	}

	if sensorMin > targetTemp {
		for _, h := range room.Hvacs {
			h.Limiter.Wait(context.Background())
			oldTemp := h.TemperatureState.Value
			newTemp := oldTemp - 0.5
			// Only change when we are increasing the temperature
			if newTemp <= targetTemp+1 && targetTemp+1 < oldTemp {
				mqttClient.PublishString(h.FanCommand, "LOW")
			}
			if newTemp <= targetTemp+2 && targetTemp+2 < oldTemp {
				mqttClient.PublishString(h.FanCommand, "MEDIUM")
			}
			mqttClient.PublishString(h.TemperatureCommand, strconv.FormatFloat(newTemp, 'f', 1, 64))
		}
	}
}

var cfg = []*Room{
	// {
	// 	Name: "Zaya",
	// 	Enabled: &MqttBool{
	// 		CommandTopic: "code/air/rooms/zaya/enabled/command",
	// 		StateTopic:   "code/air/rooms/zaya/enabled/state",
	// 		Value:        false,
	// 	},
	// 	Hvacs: []*Hvac{
	// 		{
	// 			ModeState: &MqttString{
	// 				CommandTopic: "esphome/zaya/mode_state",
	// 				Value:        "",
	// 			},
	// 			TemperatureState: &MqttFloat64{
	// 				CommandTopic: "esphome/zaya/target_temperature_low_state",
	// 				Value:        0,
	// 			},
	// 			TemperatureCommand: "esphome/zaya/target_temperature_command",
	// 			FanCommand:         "esphome/zaya/fan_mode_command",
	// 			Limiter:            rate.NewLimiter(rate.Every(5*time.Minute), 1),
	// 		},
	// 	},
	// 	AcceptableTemperatures: &Float64Range{
	// 		Min: &MqttFloat64{
	// 			CommandTopic: "code/air/rooms/zaya/acceptabletemperatures/min/command",
	// 			StateTopic:   "code/air/rooms/zaya/acceptabletemperatures/min/state",
	// 			Value:        0,
	// 		},
	// 		Max: &MqttFloat64{
	// 			CommandTopic: "code/air/rooms/zaya/acceptabletemperatures/max/command",
	// 			StateTopic:   "code/air/rooms/zaya/acceptabletemperatures/max/state",
	// 			Value:        0,
	// 		},
	// 	},
	// 	Sensors: []*MqttFloat64{
	// 		{
	// 			CommandTopic: "zigbee2mqtt/server/device/zaya/air",
	// 			Value:        0,
	// 		},
	// 		{
	// 			CommandTopic: "zigbee2mqtt/server/device/zaya/followme",
	// 			Value:        0,
	// 		},
	// 	},
	// },
	// {
	// 	Name: "Parent",
	// 	Enabled: &MqttBool{
	// 		CommandTopic: "code/air/rooms/parent/enabled/command",
	// 		StateTopic:   "code/air/rooms/parent/enabled/state",
	// 		Value:        false,
	// 	},
	// 	Hvacs: []*Hvac{
	// 		{
	// 			ModeState: &MqttString{
	// 				CommandTopic: "esphome/parent/mode_state",
	// 				Value:        "",
	// 			},
	// 			TemperatureState: &MqttFloat64{
	// 				CommandTopic: "esphome/parent/target_temperature_low_state",
	// 				Value:        0,
	// 			},
	// 			TemperatureCommand: "esphome/parent/target_temperature_command",
	// 			FanCommand:         "esphome/parent/fan_mode_command",
	// 			Limiter:            rate.NewLimiter(rate.Every(5*time.Minute), 1),
	// 		},
	// 	},
	// 	AcceptableTemperatures: &Float64Range{
	// 		Min: &MqttFloat64{
	// 			CommandTopic: "code/air/rooms/parent/acceptabletemperatures/min/command",
	// 			StateTopic:   "code/air/rooms/parent/acceptabletemperatures/min/state",
	// 			Value:        0,
	// 		},
	// 		Max: &MqttFloat64{
	// 			CommandTopic: "code/air/rooms/parent/acceptabletemperatures/max/command",
	// 			StateTopic:   "code/air/rooms/parent/acceptabletemperatures/max/state",
	// 			Value:        0,
	// 		},
	// 	},
	// 	Sensors: []*MqttFloat64{
	// 		{
	// 			CommandTopic: "zigbee2mqtt/server/device/parent/air",
	// 			Value:        0,
	// 		},
	// 		{
	// 			CommandTopic: "zigbee2mqtt/server/device/parent/followme",
	// 			Value:        0,
	// 		},
	// 	},
	// },
	{
		Name: "Office",
		Enabled: &MqttBool{
			CommandTopic: "code/air/rooms/office/enabled/command",
			StateTopic:   "code/air/rooms/office/enabled/state",
			Value:        false,
		},
		Hvacs: []*Hvac{
			{
				ModeState: &MqttString{
					CommandTopic: "esphome/office/mode_state",
					Value:        "",
				},
				TemperatureState: &MqttFloat64{
					CommandTopic: "esphome/office/target_temperature_low_state",
					Value:        0,
				},
				TemperatureCommand: "esphome/office/target_temperature_command",
				FanCommand:         "esphome/office/fan_mode_command",
				Limiter:            rate.NewLimiter(rate.Every(5*time.Minute), 1),
			},
		},
		AcceptableTemperatures: &Float64Range{
			Min: &MqttFloat64{
				CommandTopic: "code/air/rooms/office/acceptabletemperatures/min/command",
				StateTopic:   "code/air/rooms/office/acceptabletemperatures/min/state",
				Value:        0,
			},
			Max: &MqttFloat64{
				CommandTopic: "code/air/rooms/office/acceptabletemperatures/max/command",
				StateTopic:   "code/air/rooms/office/acceptabletemperatures/max/state",
				Value:        0,
			},
		},
		Sensors: []*MqttFloat64{
			{
				CommandTopic: "zigbee2mqtt/server/device/office/air",
				Value:        0,
			},
			{
				CommandTopic: "zigbee2mqtt/server/device/office/followme",
				Value:        0,
			},
		},
	},
	{
		Name: "Living room",
		Enabled: &MqttBool{
			CommandTopic: "code/air/rooms/livingroom/enabled/command",
			StateTopic:   "code/air/rooms/livingroom/enabled/state",
			Value:        false,
		},
		Hvacs: []*Hvac{
			{
				ModeState: &MqttString{
					CommandTopic: "esphome/living/mode_state",
					Value:        "",
				},
				TemperatureState: &MqttFloat64{
					CommandTopic: "esphome/living/target_temperature_low_state",
					Value:        0,
				},
				TemperatureCommand: "esphome/living/target_temperature_command",
				FanCommand:         "esphome/living/fan_mode_command",
				Limiter:            rate.NewLimiter(rate.Every(5*time.Minute), 1),
			},
			{
				ModeState: &MqttString{
					CommandTopic: "esphome/kitchen/mode_state",
					Value:        "",
				},
				TemperatureState: &MqttFloat64{
					CommandTopic: "esphome/kitchen/target_temperature_low_state",
					Value:        0,
				},
				TemperatureCommand: "esphome/kitchen/target_temperature_command",
				FanCommand:         "esphome/kitchen/fan_mode_command",
				Limiter:            rate.NewLimiter(rate.Every(5*time.Minute), 1),
			},
		},
		AcceptableTemperatures: &Float64Range{
			Min: &MqttFloat64{
				CommandTopic: "code/air/rooms/livingroom/acceptabletemperatures/min/command",
				StateTopic:   "code/air/rooms/livingroom/acceptabletemperatures/min/state",
				Value:        0,
			},
			Max: &MqttFloat64{
				CommandTopic: "code/air/rooms/livingroom/acceptabletemperatures/max/command",
				StateTopic:   "code/air/rooms/livingroom/acceptabletemperatures/max/state",
				Value:        0,
			},
		},
		Sensors: []*MqttFloat64{
			{
				CommandTopic: "zigbee2mqtt/server/device/livingroom/air",
				Value:        0,
			},
			{
				CommandTopic: "zigbee2mqtt/server/device/living/followme",
				Value:        0,
			},
			{
				CommandTopic: "zigbee2mqtt/server/device/kitchen/followme",
				Value:        0,
			},
		},
	},
}

func main() {
	for _, r := range cfg {
		room := r
		mqttClient.Client.Subscribe(room.Enabled.CommandTopic, room.GetCallback())
		mqttClient.Client.Subscribe(room.AcceptableTemperatures.Min.CommandTopic, room.GetCallback())
		mqttClient.Client.Subscribe(room.AcceptableTemperatures.Max.CommandTopic, room.GetCallback())
		for _, h := range room.Hvacs {
			hvac := h
			mqttClient.Client.Subscribe(hvac.ModeState.CommandTopic, room.GetCallback())
			mqttClient.Client.Subscribe(hvac.TemperatureState.CommandTopic, room.GetCallback())
		}
		for _, s := range room.Sensors {
			sensor := s
			mqttClient.Client.Subscribe(sensor.CommandTopic, room.GetCallback())
		}
	}
	<-quit
}
