package switches

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	pb "github.com/nanassito/home/proto/go/switches"
)

type Request struct {
	ClientID string
	From     time.Time
	Until    time.Time
}

type State struct {
	Config   pb.SwitchConfig
	Requests []Request
}

var state = struct {
	SwitchIDs          []string
	RequestsBySwitchID map[string]State
}{
	SwitchIDs:          []string{},
	RequestsBySwitchID: map[string]State{},
}

func mustLoadSwitches() (switches map[string]pb.SwitchConfig) {
	data, err := ioutil.ReadFile("./switches.json")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	if err := json.Unmarshal(data, &switches); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return switches
}

func init() {
	for switchID, switchCfg := range mustLoadSwitches() {
		state.SwitchIDs = append(state.SwitchIDs, switchID)
		state.RequestsBySwitchID[switchID] = State{
			Config:   switchCfg,
			Requests: make([]Request, 0),
		}
	}
}

type Server struct {
	pb.UnimplementedSwitchSvcServer
}

func (Server) List(ctx context.Context, req *pb.ReqList) (*pb.RspList, error) {
	return &pb.RspList{SwitchIDs: state.SwitchIDs}, nil
}

// func (Server) Activate(ctx context.Context, req *pb.ReqActivate) (*pb.RspActivate, error) {
// 	req.
// }

// func (Server) Status(context.Context, *pb.ReqStatus) (*pb.RspStatus, error) {
// 	return nil, nil
// }
