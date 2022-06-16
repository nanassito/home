package switches

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	pb "github.com/nanassito/home/proto/go/switches"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/internal/status"
)

type Request struct {
	ClientID string
	From     time.Time
	Until    time.Time
}

func (r *Request) isActive(at time.Time) bool {
	return r.From.After(at) && r.Until.Before(at)
}

type State struct {
	Config   *pb.SwitchConfig
	Requests []Request
}

var state = struct {
	SwitchIDs          []string
	RequestsBySwitchID map[string]State
}{
	SwitchIDs:          []string{},
	RequestsBySwitchID: map[string]State{},
}

func mustLoadSwitches() (switches map[string]*pb.SwitchConfig) {
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
	switches := mustLoadSwitches()
	for switchID, switchCfg := range switches {
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

func (Server) Activate(ctx context.Context, req *pb.ReqActivate) (*pb.RspActivate, error) {
	if req.GetClientID() == "" {
		return nil, status.Err(codes.InvalidArgument, "Must specify clientID")
	}
	if req.GetDurationSeconds() <= 0 {
		return nil, status.Err(codes.InvalidArgument, "Duration must be greater than 0 seconds")
	}
	switchState, ok := state.RequestsBySwitchID[req.GetSwitchID()]
	if !ok {
		return nil, status.Err(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
	}

	now := time.Now()
	switchState.Requests = append(switchState.Requests, Request{
		ClientID: req.GetClientID(),
		From:     now,
		Until:    now.Add(time.Duration(req.GetDurationSeconds()) * time.Second),
	})

	var activeUntil time.Time
	for _, request := range switchState.Requests {
		if request.isActive(now) {
			if request.Until.After(activeUntil) {
				activeUntil = request.Until
			}
		}
	}
	return &pb.RspActivate{
		SwitchID:    req.GetSwitchID(),
		ActiveUntil: int32(activeUntil.Unix()),
	}, nil
}

func (Server) Status(ctx context.Context, req *pb.ReqStatus) (*pb.RspStatus, error) {
	switchState, ok := state.RequestsBySwitchID[req.GetSwitchID()]
	if !ok {
		return nil, status.Err(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
	}

	now := time.Now()
	isActive := false
	var activeUntil time.Time
	outstandingRequests := make([]*pb.RspStatus_Request, 0, len(switchState.Requests))
	for _, request := range switchState.Requests {
		if request.isActive(now) {
			isActive = true
			if request.Until.After(activeUntil) {
				activeUntil = request.Until
			}
			outstandingRequests = append(outstandingRequests, &pb.RspStatus_Request{
				ClientID: request.ClientID,
				From:     int32(request.From.Unix()),
				To:       int32(request.Until.Unix()),
			})
		}
	}
	var activeUntilUnix *int32
	if isActive {
		a := int32(activeUntil.Unix())
		activeUntilUnix = &a
	}
	return &pb.RspStatus{
		SwitchID:    req.GetSwitchID(),
		IsActive:    isActive,
		ActiveUntil: activeUntilUnix,
		Requests:    outstandingRequests,
	}, nil
}
