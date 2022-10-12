package air

import (
	"flag"
	"log"

	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/json_strict"
	"github.com/nanassito/home/pkg/mqtt"
)

var configFile = flag.String("config", "/github/home/configs/air.json", "Air configuration file.")

type Server struct {
	air_proto.UnimplementedAirSvcServer

	Mqtt mqtt.MqttIface
	// State *ServerState
}

func New() *Server {
	var cfg *air_proto.AirConfig
	flag.Parse()
	if err := json_strict.UnmarshalFile(*configFile, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return nil
}
