package air

import (
	"context"

	"github.com/nanassito/home/pkg/air_proto"
)

func (s *Server) GetAllStates(ctx context.Context, req *air_proto.ReqGetAllStates) (*air_proto.ServerState, error) {
	return s.State, nil
}

func (s *Server) ConfigureRoom(ctx context.Context, req *air_proto.ReqConfigureRoom) (*air_proto.ServerState, error) {
	// room, ok := s.State.Rooms[req.GetRoom()]
	// if !ok {
	// 	return s.State, status.Error(codes.InvalidArgument, fmt.Sprintf("Unknown room `%s`", req.GetRoom()))
	// }
	// TODO: Make more flexible by mang every field optional
	// if req.DesiredTemperatureRange == nil || req.DesiredTemperatureRange.Min == nil || req.DesiredTemperatureRange.Max == nil {
	// 	return s.State, status.Error(codes.InvalidArgument, "DesiredTemperatureRange is nil")
	// }
	// if req.DesiredTemperatureRange.Min >= req.DesiredTemperatureRange.Max {
	// 	return s.State, status.Error(codes.InvalidArgument, "Min temperature must be less than max temperature.")
	// }
	// room.DesiredTemperatureRange = req.DesiredTemperatureRange
	go s.Control()
	return s.State, nil
}
