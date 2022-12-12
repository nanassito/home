package air

import (
	"context"
	"fmt"

	"github.com/nanassito/home/pkg/air_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
