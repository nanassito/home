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

func (s *Server) SetState(ctx context.Context, req *air_proto.ServerState) (*air_proto.ServerState, error) {
	for roomId, room := range req.Rooms {
		state, ok := s.State.Rooms[roomId]
		if !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No room with ID `%s`", roomId))
		}
		if room.DesiredTemperatureRange != nil {
			logger.Printf("Info| API: changing room %s desired temperature range.\n", state.Name)
			state.DesiredTemperatureRange = room.DesiredTemperatureRange
		}

		if room.Schedule != nil && room.Schedule.IsActive != nil {
			if *room.Schedule.IsActive {
				logger.Printf("Info| API: Enabling schedule for room %s.\n", state.Name)
			} else {
				logger.Printf("Info| API: Disabling schedule for room %s.\n", state.Name)
			}
			state.Schedule.IsActive = room.Schedule.IsActive
		}
	}

	for hvacId, hvac := range req.Hvacs {
		state, ok := s.State.Hvacs[hvacId]
		if !ok {
			return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("No hvac with ID `%s`", hvacId))
		}
		if hvac.Control != air_proto.Hvac_CONTROL_UNKNOWN {
			logger.Printf("Info| API: changing hvac %s control to %v.\n", state.Name, hvac.Control)
			state.Control = hvac.Control
		}
		if hvac.TemperatureOffset != nil {
			logger.Printf("Info| API: changing hvac %s offset to %.2f\n", state.Name, *hvac.TemperatureOffset)
			state.TemperatureOffset = hvac.TemperatureOffset
		}
		if hvac.DesiredState != nil {
			if hvac.DesiredState.Mode != air_proto.Hvac_MODE_UNKNOWN {
				logger.Printf("Info| API: changing hvac %s mode to %v\n", state.Name, hvac.DesiredState.Mode)
				state.DesiredState.Mode = hvac.DesiredState.Mode
			}
		}
	}
	return s.State, nil
}
