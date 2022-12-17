package air

import (
	"math"

	"github.com/nanassito/home/pkg/air_proto"
)

const hvacMinimalHeatTemperature = 17

func (s *Server) InferGeneralHvacMode() air_proto.Hvac_Mode {
	// If an HVAC is under direct control (control = hvac or none) and is turned on (mode != off)
	// Then that mode should be used for all Hvacs.
	for _, hvac := range s.State.Hvacs {
		if hvac.Control == air_proto.Hvac_CONTROL_HVAC || hvac.Control == air_proto.Hvac_CONTROL_NONE {
			if hvac.ReportedState.Mode != air_proto.Hvac_MODE_OFF {
				return hvac.ReportedState.Mode
			}
		}
	}

	// All the rooms vote for the mode they need.
	// We add the distance between the limit temperature and the reported temperature.
	// Whichever mode has the lowest score is the one we need to use.
	votes := map[air_proto.Hvac_Mode]float64{
		air_proto.Hvac_MODE_COOL: 0,
		air_proto.Hvac_MODE_HEAT: 0,
	}
	for _, roomState := range s.State.Rooms {
		hasControlledHvac := false
		for hvacName, hvacCfg := range s.Config.Hvacs {
			if hvacCfg.Room == roomState.Name {
				if s.State.Hvacs[hvacName].Control == air_proto.Hvac_CONTROL_ROOM {
					hasControlledHvac = true
					break
				}
			}
		}
		if !hasControlledHvac {
			logger.Printf("Info| Room %s doesn't have any hvacs under control.\n", roomState.Name)
			continue
		}
		if !IsSensorAlive(roomState.Sensor) {
			logger.Printf("Fail| Stalled sensor in %s\n", roomState.Name)
			continue
		}
		votes[air_proto.Hvac_MODE_HEAT] += roomState.Sensor.Temperature - roomState.DesiredTemperatureRange.Min
		votes[air_proto.Hvac_MODE_COOL] += roomState.DesiredTemperatureRange.Max - roomState.Sensor.Temperature
	}
	if votes[air_proto.Hvac_MODE_COOL] < votes[air_proto.Hvac_MODE_HEAT] {
		return air_proto.Hvac_MODE_COOL
	} else {
		return air_proto.Hvac_MODE_HEAT
	}
}

func (s *Server) Control() {
	generalMode := s.InferGeneralHvacMode()
	logger.Printf("Info| Inferred general mode is %s\n", generalMode.String())
	switch generalMode {
	case air_proto.Hvac_MODE_HEAT:
		s.HeatUp()
	case air_proto.Hvac_MODE_COOL:
		s.DecideCoolDown()
	default:
		logger.Printf("Warn| Unsuported general mode `%s`\n", generalMode.String())
	}
}

func DecideHeatUpMode(room *air_proto.Room, outside *air_proto.Sensor, hvacName string) air_proto.Hvac_Mode {
	roomWillWarmUp := false
	if IsSensorAlive(outside) {
		roomWillWarmUp = outside.Temperature > room.DesiredTemperatureRange.Min
	}

	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+2 {
		logger.Printf("Info| Room %s is more than hot enough, shutting hvac %s down.\n", room.Name, hvacName)
		return air_proto.Hvac_MODE_OFF
	}
	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min && roomWillWarmUp {
		logger.Printf("Info| Room %s will continue to warm up, shutting hvac %s down.\n", room.Name, hvacName)
		return air_proto.Hvac_MODE_OFF
	}
	return air_proto.Hvac_MODE_HEAT
}

func DecideHeatUpFan(room *air_proto.Room, last30mHvacsTemp map[string]float64, hvacName string) air_proto.Hvac_Fan {
	if room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
		if last30mTemp, ok := last30mHvacsTemp[hvacName]; ok {
			if last30mTemp > room.Sensor.Temperature+3 {
				return air_proto.Hvac_FAN_HIGH
			}
			if last30mTemp > room.Sensor.Temperature+1.5 {
				return air_proto.Hvac_FAN_MEDIUM
			}
		} else {
			logger.Printf("Fail| Got no historical temperature data for hvac %s\n", hvacName)
		}
	}
	return air_proto.Hvac_FAN_AUTO
}

func DecideHeatUpTemperature(room *air_proto.Room, hvac *air_proto.Hvac, tempDeltas map[string]float64) (temperature float64, offset float64) {
	temperature = math.Max(hvacMinimalHeatTemperature, room.DesiredTemperatureRange.Min)
	step := 0.2 // Note that the hvac can only step by 0.5°C

	if hvac.DesiredState.Temperature != room.DesiredTemperatureRange.Min {
		// The desired temperature changed so we need to reset the offset
		return temperature, 0
	}

	if delta, ok := tempDeltas[room.Name]; ok {
		if delta <= 0 && room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
			offset += step
		}
		if delta >= 0 && room.Sensor.Temperature > room.DesiredTemperatureRange.Min {
			offset -= step
		}
	} else {
		logger.Printf("Fail| Got no historical temperature data for room %s\n", room.Name)
	}

	if room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
		offset += step
	}
	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+1 {
		offset -= step
	}

	// Speed up when we are too far off target
	if room.Sensor.Temperature < room.DesiredTemperatureRange.Min-3 {
		offset = math.Max(offset, room.DesiredTemperatureRange.Min-room.Sensor.Temperature)
	}
	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+3 {
		offset = math.Min(offset, 0)
	}

	return temperature, offset
}

func (s *Server) HeatUp() {
	last30mHvacsTemp, err := s.GetLast30mHvacTemperatures()
	if err != nil {
		logger.Printf("Fail| Error querying Prometheus (last30mHvacsTemp): %v\n", err)
	}
	tempDeltas, err := s.Get30mRoomTemperatureDeltas()
	if err != nil {
		logger.Printf("Fail| Error querying Prometheus (tempDeltas): %v\n", err)
	}

	for _, hvac := range s.State.Hvacs {
		roomName := s.Config.Hvacs[hvac.Name].Room

		if hvac.Control != air_proto.Hvac_CONTROL_ROOM {
			logger.Printf("Info| Hvac %s is not controlled by the room. Skipping\n", hvac.Name)
			continue
		}

		if !IsSensorAlive(s.State.Rooms[roomName].Sensor) {
			logger.Printf("Fail| Lost temperature sensor in %s. Giving up control.\n", roomName)
			continue
		}

		hvac.DesiredState.Mode = DecideHeatUpMode(s.State.Rooms[roomName], s.State.Outside, hvac.Name)
		hvac.DesiredState.Fan = DecideHeatUpFan(s.State.Rooms[roomName], last30mHvacsTemp, hvac.Name)
		hvac.DesiredState.Temperature, hvac.TemperatureOffset = DecideHeatUpTemperature(s.State.Rooms[roomName], hvac, tempDeltas)
	}
}

func (s *Server) DecideCoolDown() {
	logger.Println("Fail| DecideCoolDown is not implemented")
}
