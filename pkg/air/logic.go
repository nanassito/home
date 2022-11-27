package air

import (
	"github.com/nanassito/home/pkg/air_proto"
)

const (
	DecisionTemp = 23
)

func (s *Server) InferGeneralMode() air_proto.Hvac_Mode {
	votes := map[air_proto.Hvac_Mode]int{
		air_proto.Hvac_MODE_COOL: 0,
		air_proto.Hvac_MODE_HEAT: 0,
	}
	for _, room := range s.State.Rooms {
		for _, hvac := range room.Hvacs {
			if hvac.Control == air_proto.Hvac_CONTROL_HVAC && hvac.ReportedState.Mode != air_proto.Hvac_MODE_OFF {
				logger.Printf("Warn| Hvac %s is in forced %s\n", hvac.HvacName, hvac.ReportedState.Mode.String())
				return hvac.ReportedState.Mode
			}
		}
		if IsSensorAlive(room.Sensor) {
			if room.Sensor.Temperature > DecisionTemp {
				votes[air_proto.Hvac_MODE_COOL] += 1
			}
			if room.Sensor.Temperature < DecisionTemp {
				votes[air_proto.Hvac_MODE_HEAT] += 1
			}
		} else {
			logger.Printf("Fail| Stalled sensor in %s\n", room.RoomName)
		}
	}
	if votes[air_proto.Hvac_MODE_COOL] > votes[air_proto.Hvac_MODE_HEAT] {
		return air_proto.Hvac_MODE_COOL
	} else {
		return air_proto.Hvac_MODE_HEAT
	}
}

func (s *Server) Decide() {
	generalMode := s.InferGeneralMode()
	logger.Printf("Info| Infered general mode is %s\n", generalMode.String())
	switch generalMode {
	case air_proto.Hvac_MODE_HEAT:
		s.DecideHeatUp()
	case air_proto.Hvac_MODE_COOL:
		s.DecideCoolDown()
	default:
		logger.Printf("Warn| Unsuported general mode `%s`\n", generalMode.String())
	}
}

func (s *Server) DecideHeatUp() {
	last30mHvacsTemp, err := s.GetLast30mHvacsTemperature()
	if err != nil {
		logger.Printf("Fail| Error querying Prometheus (last30mHvacsTemp): %v\n", err)
	}
	tempDeltas, err := s.Get30mRoomTemperatureDeltas()
	if err != nil {
		logger.Printf("Fail| Error querying Prometheus (tempDeltas): %v\n", err)
	}
	for _, room := range s.State.Rooms {
		if !IsSensorAlive(room.Sensor) {
			logger.Printf("Fail| Lost temperature sensor in %s. Giving up control.\n", room.RoomName)
			continue
		}
		roomWillWarmUp := false
		if IsSensorAlive(s.State.OutsideSensor) {
			roomWillWarmUp = s.State.OutsideSensor.Temperature > room.DesiredTemperatureRange.Min
		}
		for _, hvac := range room.Hvacs {
			// First let's figure out if we run at all or not.
			if hvac.Control != air_proto.Hvac_CONTROL_ROOM {
				logger.Printf("Info| Hvac %s is not controlled by the room. Skipping\n", hvac.HvacName)
				continue
			}
			if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+2 {
				logger.Printf("Info| Room %s is more than hot enough, shutting hvac %s down.\n", room.RoomName, hvac.HvacName)
				hvac.DesiredState.Mode = air_proto.Hvac_MODE_OFF
				continue
			}
			if room.Sensor.Temperature > room.DesiredTemperatureRange.Min && roomWillWarmUp {
				logger.Printf("Info| Room %s will continue to warm up, shutting hvac %s down.\n", room.RoomName, hvac.HvacName)
				hvac.DesiredState.Mode = air_proto.Hvac_MODE_OFF
				continue
			}
			hvac.DesiredState.Mode = air_proto.Hvac_MODE_HEAT

			// Then at what speed to we run the fan
			fan := air_proto.Hvac_FAN_AUTO
			if room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
				if last30mTemp, ok := last30mHvacsTemp[hvac.HvacName]; ok {
					if last30mTemp > room.Sensor.Temperature+1.5 {
						fan = air_proto.Hvac_FAN_MEDIUM
					}
					if last30mTemp > room.Sensor.Temperature+3 {
						fan = air_proto.Hvac_FAN_HIGH
					}
				} else {
					logger.Printf("Fail| Got no historical temperature data for hvac %s\n", hvac.HvacName)
				}
			}
			hvac.DesiredState.Fan = fan

			// And finally how much do we need to offset the temperature.
			hvac.DesiredState.Temperature = room.DesiredTemperatureRange.Min
			if delta, ok := tempDeltas[room.RoomName]; ok {
				if delta <= 0 && room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
					hvac.TemperatureOffset += 1
				}
				if delta >= 0 && room.Sensor.Temperature > room.DesiredTemperatureRange.Min+1 {
					hvac.TemperatureOffset -= 1
				}
			} else {
				logger.Printf("Fail| Got no historical temperature data for room %s\n", room.RoomName)
			}
		}
	}
}

func (s *Server) DecideCoolDown() {
	logger.Println("Fail| DecideCoolDown is not implemented")
}
