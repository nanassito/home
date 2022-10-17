package switches

import (
	"context"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nanassito/home/pkg/json_strict"
	"github.com/nanassito/home/pkg/mqtt"
	"github.com/nanassito/home/pkg/prom"
	"github.com/nanassito/home/pkg/switches_proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var configFile = flag.String("config", "/github/home/configs/switches.json", "Switch configuration file.")

var (
	PromDesiredState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "switch",
			Name:      "desired_state",
			Help:      "Desired state of a switch.",
		},
		[]string{"switchID"},
	)
	PromReportedState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "switch",
			Name:      "reported_state",
			Help:      "Reported state of a switch. For convenience only since this is already in prom anyway.",
		},
		[]string{"switchID"},
	)
)

type Request struct {
	ClientID string
	From     time.Time
	Until    time.Time
}

func (r *Request) isActive(at time.Time) bool {
	return (r.From.Equal(at) || r.From.Before(at)) && r.Until.After(at)
}

type State struct {
	SwitchID        string
	Config          *switches_proto.SwitchConfig
	Requests        []Request
	ReportedActive  bool
	lastMqttMessage time.Time
}

func boolAsFloat(v bool) float64 {
	if v {
		return 1
	} else {
		return 0
	}
}

func (s *State) monitor() {
	labels := strings.Builder{}
	for key, value := range s.Config.Prometheus.Labels {
		labels.WriteString(key + "=\"" + value + "\", ")
	}
	promql := fmt.Sprintf("%s{%s}", s.Config.Prometheus.Metric, labels.String())
	for {
		stateAsNum, err := prom.QueryOne(promql, "state_"+s.SwitchID)
		fmt.Printf("monitorSwitchState | %s | %s = %f\n", s.SwitchID, promql, stateAsNum)
		if err == nil {
			switch int32(stateAsNum) {
			case s.Config.Prometheus.ValueActive:
				s.ReportedActive = true
			case s.Config.Prometheus.ValueRest:
				s.ReportedActive = false
			default:
				fmt.Printf("monitorSwitchState | %s | Unknown state %f\n", s.SwitchID, stateAsNum)
				// TODO: Instrument the invalid result
			}
			PromReportedState.WithLabelValues(s.SwitchID).Set(boolAsFloat(s.ReportedActive))
		} else {
			fmt.Printf("monitorSwitchState | %s | %v\n", s.SwitchID, err)
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

var bool2Verb = map[bool]string{
	true:  "active",
	false: "rest",
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
		PromDesiredState.WithLabelValues(s.SwitchID).Set(boolAsFloat(shouldBeActive))
		if s.ReportedActive != shouldBeActive {
			msg := s.Config.Mqtt.MsgRest
			if shouldBeActive {
				msg = s.Config.Mqtt.MsgActive
			}
			fmt.Printf("ControlLoop | %s | Changing state from %v to %v\n", s.SwitchID, bool2Verb[s.ReportedActive], bool2Verb[shouldBeActive])
			err := mqtt.PublishString(s.Config.Mqtt.SetTopic, msg)
			if err == nil {
				s.ReportedActive = shouldBeActive
			} else {
				fmt.Printf("Mqtt failure: %v\n", err)
				mqtt.Reset()
			}
		}
		prom.LoopRunsCounter.With(prometheus.Labels{
			"domain":   "Switches",
			"type":     "controlSwitchState",
			"instance": s.SwitchID,
		}).Inc()
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *State) updateOnMqtt(topic string, payload []byte) {
	s.lastMqttMessage = time.Now()
	rx := regexp.MustCompile(s.Config.Mqtt.GetRegex)
	groups := rx.FindStringSubmatch(string(payload))
	isActive := groups[rx.SubexpIndex("Active")] != ""
	isAtRest := groups[rx.SubexpIndex("AtRest")] != ""
	if isActive == isAtRest {
		fmt.Printf("updateOnMqtt | %s | Invalid value `%v` in mqtt payload `%v`\n", s.SwitchID, groups[rx.SubexpIndex("State")], payload)
		return
	}
	s.ReportedActive = isActive
}

type ServerState struct {
	SwitchIDs  []string
	BySwitchID map[string]*State
}

type Server struct {
	switches_proto.UnimplementedSwitchSvcServer

	Mqtt  mqtt.MqttIface
	State *ServerState
}

func New() *Server {
	server := Server{
		Mqtt: mqtt.New("switches"),
		State: &ServerState{
			SwitchIDs:  []string{},
			BySwitchID: map[string]*State{},
		},
	}

	var switches map[string]*switches_proto.SwitchConfig
	flag.Parse()
	if err := json_strict.UnmarshalFile(*configFile, &switches); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	for switchID, switchCfg := range switches {
		server.State.SwitchIDs = append(server.State.SwitchIDs, switchID)
		server.State.BySwitchID[switchID] = &State{
			SwitchID: switchID,
			Config:   switchCfg,
			Requests: make([]Request, 0),
		}
	}
	go server.ControlLoop()
	return &server
}

func (s *Server) List(ctx context.Context, req *switches_proto.ReqList) (*switches_proto.RspList, error) {
	return &switches_proto.RspList{SwitchIDs: s.State.SwitchIDs}, nil
}

func (s *Server) Activate(ctx context.Context, req *switches_proto.ReqActivate) (*switches_proto.RspActivate, error) {
	if req.GetClientID() == "" {
		err := status.Error(codes.InvalidArgument, "Must specify clientID")
		fmt.Printf("Activate | %v", err)
		return nil, err
	}
	if req.GetDurationSeconds() <= 0 {
		err := status.Error(codes.InvalidArgument, "Duration must be greater than 0 seconds")
		fmt.Printf("Activate | %v", err)
		return nil, err
	}
	switchState, ok := s.State.BySwitchID[req.GetSwitchID()]
	if !ok {
		err := status.Error(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
		fmt.Printf("Activate | %v", err)
		return nil, err
	}

	fmt.Printf("Activate | %v", req)
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
	return &switches_proto.RspActivate{
		SwitchID:    req.GetSwitchID(),
		ActiveUntil: activeUntil.Unix(),
	}, nil
}

func (s *Server) Status(ctx context.Context, req *switches_proto.ReqStatus) (*switches_proto.RspStatus, error) {
	switchState, ok := s.State.BySwitchID[req.GetSwitchID()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Not switch with ID "+req.GetSwitchID())
	}

	now := time.Now()
	isActive := false
	var activeUntil time.Time
	outstandingRequests := make([]*switches_proto.RspStatus_Request, 0, len(switchState.Requests))
	for _, request := range switchState.Requests {
		if request.isActive(now) {
			isActive = true
			if request.Until.After(activeUntil) {
				activeUntil = request.Until
			}
			outstandingRequests = append(outstandingRequests, &switches_proto.RspStatus_Request{
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
	return &switches_proto.RspStatus{
		SwitchID:    req.GetSwitchID(),
		IsActive:    isActive,
		ActiveUntil: activeUntilUnix,
		Requests:    outstandingRequests,
	}, nil
}

func (s *Server) ControlLoop() {
	for _, switchID := range s.State.SwitchIDs {
		fmt.Printf("Starting monitoring and control loops for %s\n", switchID)
		switchState := s.State.BySwitchID[switchID]
		s.Mqtt.Subscribe(switchState.Config.Mqtt.GetTopic, switchState.updateOnMqtt)
		go switchState.monitor()
		go switchState.control(s.Mqtt)
	}
}
