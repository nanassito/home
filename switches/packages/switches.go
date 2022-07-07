package switches

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/nanassito/home/lib/mqtt"
	prom "github.com/nanassito/home/lib/prometheus"
	pb "github.com/nanassito/home/proto/switches"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var configFile = flag.String("config", "/github/home/switches/switches.json", "Switch configuration file.")

type Request struct {
	ClientID string
	From     time.Time
	Until    time.Time
}

func (r *Request) isActive(at time.Time) bool {
	return (r.From.Equal(at) || r.From.Before(at)) && r.Until.After(at)
}

type State struct {
	SwitchID       string
	Config         *pb.SwitchConfig
	Requests       []Request
	ReportedActive bool
}

func (s *State) monitor() {
	labels := strings.Builder{}
	for key, value := range s.Config.Prometheus.Labels {
		labels.WriteString(key + "=\"" + value + "\", ")
	}
	promql := fmt.Sprintf("%s{%s}", s.Config.Prometheus.Metric, labels.String())
	for {
		stateAsNum, err := prom.QueryOne(promql, "state_"+s.SwitchID)
		if err == nil {
			switch stateAsNum {
			case 0:
				s.ReportedActive = false
			case 1:
				s.ReportedActive = true
			default:
				fmt.Printf("monitorSwitchState | %s | Unknown state %f\n", s.SwitchID, stateAsNum)
				// TODO: Instrument the invalid result
			}
		} else {
			fmt.Printf("monitorSwitchState | %s | %v", s.SwitchID, err)
			// TODO: Instrument the failure
		}
		prom.LoopRunsCounter.With(prometheus.Labels{
			"domain":   "Switches",
			"type":     "monitorSwitchState",
			"instance": s.SwitchID,
		}).Inc()
		time.Sleep(1 * time.Minute)
	}
}

func (s *State) control(mqtt mqtt.MqttIface) {
	for {
		shouldBeActive := false
		for _, request := range s.Requests {
			if request.isActive(time.Now()) {
				shouldBeActive = true
				break
			}
		}
		if s.ReportedActive != shouldBeActive {
			fmt.Printf("ControlLoop | %s | Changing state from %v to %v\n", s.SwitchID, s.ReportedActive, shouldBeActive)
			msg := s.Config.Mqtt.MsgOff
			if shouldBeActive {
				msg = s.Config.Mqtt.MsgOn
			}
			err := mqtt.PublishString(s.Config.Mqtt.SetTopic, msg)
			if err == nil {
				s.ReportedActive = shouldBeActive
			} else {
				fmt.Printf("Mqtt failure: %v\n", err)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type ServerState struct {
	SwitchIDs  []string
	BySwitchID map[string]State
}

type Server struct {
	pb.UnimplementedSwitchSvcServer

	Mqtt  mqtt.MqttIface
	State *ServerState
}

func New() *Server {
	server := Server{
		Mqtt: mqtt.New("switches"),
		State: &ServerState{
			SwitchIDs:  []string{},
			BySwitchID: map[string]State{},
		},
	}

	var switches map[string]*pb.SwitchConfig
	flag.Parse()
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	if err := json.Unmarshal(data, &switches); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	for switchID, switchCfg := range switches {
		server.State.SwitchIDs = append(server.State.SwitchIDs, switchID)
		server.State.BySwitchID[switchID] = State{
			SwitchID: switchID,
			Config:   switchCfg,
			Requests: make([]Request, 0),
		}
	}
	go server.ControlLoop()
	return &server
}

func (s *Server) List(ctx context.Context, req *pb.ReqList) (*pb.RspList, error) {
	return &pb.RspList{SwitchIDs: s.State.SwitchIDs}, nil
}

func (s *Server) Activate(ctx context.Context, req *pb.ReqActivate) (*pb.RspActivate, error) {
	if req.GetClientID() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must specify clientID")
	}
	if req.GetDurationSeconds() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Duration must be greater than 0 seconds")
	}
	switchState, ok := s.State.BySwitchID[req.GetSwitchID()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
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
		ActiveUntil: activeUntil.Unix(),
	}, nil
}

func (s *Server) Status(ctx context.Context, req *pb.ReqStatus) (*pb.RspStatus, error) {
	switchState, ok := s.State.BySwitchID[req.GetSwitchID()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
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
				From:     request.From.Unix(),
				To:       request.Until.Unix(),
			})
		}
	}
	var activeUntilUnix *int64
	if isActive {
		a := activeUntil.Unix()
		activeUntilUnix = &a
	}
	return &pb.RspStatus{
		SwitchID:    req.GetSwitchID(),
		IsActive:    isActive,
		ActiveUntil: activeUntilUnix,
		Requests:    outstandingRequests,
	}, nil
}

func (s *Server) ControlLoop() {
	// Monitor the state of the switch
	for _, switchState := range s.State.BySwitchID {
		go switchState.monitor()
		go switchState.control(s.Mqtt)
	}
}
