package air

import (
	"context"
	"fmt"

	"github.com/nanassito/home/pkg/air_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// var timeRx = regexp.MustCompile("^[0-9]{2}:[0-9]{2}$")

func (s *Server) GetState(ctx context.Context, req *air_proto.ReqGetState) (*air_proto.ServerState, error) {
	return s.State, nil
}

func (s *Server) SetAllStates(ctx context.Context, req *air_proto.ServerState) (*air_proto.ServerState, error) {
	for roomId, room := range req.Rooms {
		state, ok := s.State.Rooms[roomId]
		if !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No room with ID `%s`", roomId))
		}
		if room.DesiredTemperatureRange != nil {
			logger.Printf("Info| API: changing room %s desired temperature range.\n", state.Name)
			state.DesiredTemperatureRange = room.DesiredTemperatureRange
		}
	}

	for scheduleId, schedule := range req.Schedules {
		state, ok := s.State.Schedules[scheduleId]
		if !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No schedule with ID `%s`", scheduleId))
		}
		if schedule.IsActive != nil {
			if *schedule.IsActive {
				logger.Printf("Info| API: Enabling schedule %s.\n", scheduleId)
			} else {
				logger.Printf("Info| API: Disabling schedule %s.\n", scheduleId)
			}
			state.IsActive = schedule.IsActive
		}
	}

	for hvacId, hvac := range req.Hvacs {
		state, ok := s.State.Hvacs[hvacId]
		if !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No hvac with ID `%s`", hvacId))
		}
		if hvac.Control != air_proto.Hvac_CONTROL_UNKNOWN {
			logger.Printf("Info| API: changing hvac control............... %s desired temperature range.\n", state.Name)
			state.Control = hvac.Control
		}
	}
	return s.State, nil
}
