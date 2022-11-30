package air

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/json_strict"
	"github.com/nanassito/home/pkg/mqtt"
)

var (
	configFile         = flag.String("config", "/github/home/configs/air.json", "Air configuration file.")
	decideLoopInterval = flag.Duration("interval", 5*time.Minute, "Interval between two control loops (default 5m).")
	logger             = log.New(os.Stderr, "", log.Lshortfile)
)

type Server struct {
	air_proto.UnimplementedAirSvcServer

	Mqtt  mqtt.MqttIface
	State *air_proto.ServerState

	GetLast30mHvacsTemperature  func() (map[string]float64, error)
	Get30mRoomTemperatureDeltas func() (map[string]float64, error)
}

func NewServer() *Server {
	var cfg *air_proto.AirConfig
	flag.Parse()
	if err := json_strict.UnmarshalFile(*configFile, &cfg); err != nil {
		logger.Fatalf("Failed to parse config: %v\n", err)
	}

	mqttClient := mqtt.New("hvac")
	rooms := make(map[string]*air_proto.RoomState, len(cfg.Rooms))
	for roomName, roomCfg := range cfg.Rooms {
		rooms[roomName] = NewRoom(roomName, roomCfg, mqttClient)
	}
	outsideSensor := air_proto.Sensor{Location: "outside"}
	err := mqttClient.Subscribe(cfg.OutsideSensor.MqttTopic, sensorRefresher(&outsideSensor))
	if err != nil {
		panic(err)
	}

	server := Server{
		Mqtt: mqttClient,
		State: &air_proto.ServerState{
			Rooms:         rooms,
			OutsideSensor: &outsideSensor,
		},
	}
	RegisterGetLast30mHvacsTemperature(&server, cfg)
	RegisterGet30mRoomTemperatureDeltas(&server, cfg)

	go func() {
		time.Sleep(*decideLoopInterval)
		server.Decide()
	}()

	return &server
}
