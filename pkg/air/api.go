package air

import (
	"context"
	"fmt"
	"regexp"

	"github.com/nanassito/home/pkg/air_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var timeRx = regexp.MustCompile("^[0-9]{2}:[0-9]{2}$")

func (s *Server) GetAllStates(ctx context.Context, req *air_proto.ReqGetAllStates) (*air_proto.ServerState, error) {
	return s.State, nil
}

func (s *Server) ConfigureRoom(ctx context.Context, req *air_proto.ReqConfigureRoom) (*air_proto.ServerState, error) {
	room, ok := s.State.Rooms[req.GetRoom()]
	if !ok {
		return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("Unknown room `%s`", req.GetRoom()))
	}
	if req.DesiredTemperatureRange == nil {
		return s.State, status.Error(codes.InvalidArgument, "Got an empty request.")
	}

	setMin := room.DesiredTemperatureRange.Min
	setMax := room.DesiredTemperatureRange.Max
	if req.DesiredTemperatureRange.Min != 0 {
		setMin = req.DesiredTemperatureRange.Min
	}
	if req.DesiredTemperatureRange.Max != 0 {
		setMax = req.DesiredTemperatureRange.Max
	}

	if setMin >= setMax {
		return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("Min (%2f) temperature must be less than max (%2f) temperature.", setMin, setMax))
	}
	room.DesiredTemperatureRange.Min = setMin
	room.DesiredTemperatureRange.Max = setMax

	go s.Control()
	return s.State, nil
}

func (s *Server) ConfigureSchedule(ctx context.Context, req *air_proto.ReqConfigureSchedule) (*air_proto.ServerState, error) {
	if req.GetId() == "" {
		return s.State, status.Error(codes.InvalidArgument, "schedule ID is required.")
	}

	if req.Schedule == nil {
		if _, ok := s.State.Schedules[req.GetId()]; !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No schedule with id `%s`", req.GetId()))
		}
		logger.Printf("Deleting schedule %s", req.Id)
		delete(s.State.Schedules, req.GetId())

	} else {
		roomName := req.Schedule.GetRoomName()
		if _, ok := s.State.Rooms[roomName]; !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No room with id `%s`", roomName))
		}
		for _, spec := range req.Schedule.Weekday {
			if !timeRx.MatchString(spec.GetStart()) {
				return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid start time `%s`", spec.GetStart()))
			}
			if !timeRx.MatchString(spec.GetEnd()) {
				return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid end time `%s`", spec.GetEnd()))
			}
			if spec.Settings == nil {
				return s.State, status.Error(codes.InvalidArgument, "Must specify a temperature range")
			}
			// TODO enforce absolute min/max
		}
		s.State.Schedules[req.GetId()] = req.Schedule
	}
	return s.State, nil
}
