package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nanassito/home/pkg/mqtt"
)

var logger = log.New(os.Stderr, "", log.Lshortfile)

type JsonWithTemperature struct {
	Temperature float64 `json:"temperature"`
}

func main() {
	client := mqtt.New("esp_json_sensor")
	for src, dst := range map[string]string{
		"zigbee2mqtt/server/device/kitchen/followme": "espjsonsensor/kitchen/followme/temperature",
		"zigbee2mqtt/server/device/living/followme":  "espjsonsensor/living/followme/temperature",
		"zigbee2mqtt/server/device/office/followme":  "espjsonsensor/office/followme/temperature",
		"zigbee2mqtt/server/device/parent/followme":  "espjsonsensor/parent/followme/temperature",
		"zigbee2mqtt/server/device/zaya/followme":    "espjsonsensor/zaya/followme/temperature",
	} {
		src := src // Fuck golang for not having value closure on for loops
		dst := dst
		client.Subscribe(src, func(topic string, payload []byte) {
			decoded := &JsonWithTemperature{}
			json.Unmarshal(payload, decoded)
			logger.Printf("%s received %.2f\n", src, decoded.Temperature)
			client.PublishString(dst, strconv.FormatFloat(decoded.Temperature, 'f', 2, 64))
		})
	}
	for {
		time.Sleep(1 * time.Minute)
	}
}
