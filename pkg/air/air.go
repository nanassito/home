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
	configFile          = flag.String("config", "/github/home/configs/air.json", "Air configuration file.")
	controlLoopInterval = flag.Duration("interval", 5*time.Minute, "Interval between two control loops (default 5m).")
	initFromProm        = flag.Bool("init-from-prom", true, "Whether to initializing from Prometheus data.")
	readonly            = flag.Bool("readonly", false, "Don't apply decisions.")
	logger              = log.New(os.Stderr, "", log.Lshortfile)
)

type Server struct {
	air_proto.UnimplementedAirSvcServer

	Mqtt   mqtt.MqttIface
	State  *air_proto.ServerState
	Config *air_proto.AirConfig

	// GetLast30mHvacsTemperature  func() (map[string]float64, error)
	// Get30mRoomTemperatureDeltas func() (map[string]float64, error)
}

func NewServer() *Server {
	var cfg *air_proto.AirConfig
	flag.Parse()
	if err := json_strict.UnmarshalFile(*configFile, &cfg); err != nil {
		logger.Fatalf("Failed to parse config: %v\n", err)
	}

	server := Server{
		Mqtt:   mqtt.New("air"),
		Config: cfg,
	}
	server.initState()
	server.initHvacMatchers()
	server.initRoomMatchers()

	go func() {
		for {
			time.Sleep(*controlLoopInterval)
			server.Control()
		}
	}()
	return &server
}

func (s *Server) initState() {
	rooms := map[string]*air_proto.Room{}
	for roomName, roomCfg := range s.Config.Sensors {
		rooms[roomName] = NewRoom(roomName, roomCfg, s.Mqtt)
	}
	hvacs := map[string]*air_proto.Hvac{}
	for hvacName, hvacCfg := range s.Config.Hvacs {
		hvacs[hvacName] = NewHvac(hvacName, hvacCfg, s.Mqtt)
	}
	s.State = &air_proto.ServerState{
		Outside: NewSensor("outside", s.Config.Outside, s.Mqtt),
		Rooms:   rooms,
		Hvacs:   hvacs,
	}
}
