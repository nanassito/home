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

func DecideHeatUpFan(nextOffset float64, hvacName string) air_proto.Hvac_Fan {
	if math.Abs(nextOffset) > 2.5 {
		return air_proto.Hvac_FAN_HIGH
	}
	if math.Abs(nextOffset) > 1 {
		return air_proto.Hvac_FAN_MEDIUM
	}
	return air_proto.Hvac_FAN_AUTO
}

func DecideHeatUpTemperature(room *air_proto.Room, hvac *air_proto.Hvac) (temperature float64, offset float64, fastRamp bool) {
	temperature = room.DesiredTemperatureRange.Min
	offset = hvac.TemperatureOffset
	fastRamp = hvac.FastRamp
	step := 0.2 // Note that the hvac can only step by 0.5°C

	if !*readonly && hvac.DesiredState.Temperature != room.DesiredTemperatureRange.Min {
		logger.Printf("Info| %s: Room target temperature changed, resetting the offset.", hvac.Name)
		// The desired temperature changed so we need to reset the offset
		return math.Max(hvacMinimalHeatTemperature, temperature), 0, fastRamp
	}

	if hvac.FastRamp && room.Sensor.Temperature > room.DesiredTemperatureRange.Min {
		logger.Printf("Info| %s: fastRamp is complete, resetting the offset.", hvac.Name)
		offset = math.Min(offset, 0)
		fastRamp = false
	}

	if room.Sensor.Temperature < room.DesiredTemperatureRange.Min {
		offset += step
		logger.Printf("Info| %s: Room is below min, increasing the offset (%.2f).", hvac.Name, offset)
	}
	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+1 {
		offset -= step
		logger.Printf("Info| %s: Room is above min, lowering the offset (%.2f).", hvac.Name, offset)
	}

	// Speed up when we are too far off target (aka. fastRamp)
	if room.Sensor.Temperature < room.DesiredTemperatureRange.Min-3 {
		fastRamp = true
		offset = math.Max(offset, room.DesiredTemperatureRange.Min-room.Sensor.Temperature)
		logger.Printf("Info| %s: (fastRamp) Dramatically increasing the offset to catch up on heating (%.2f).", hvac.Name, offset)
	}

	// We've heat up way too much, cancel any positive offset.
	if room.Sensor.Temperature > room.DesiredTemperatureRange.Min+3 {
		offset = math.Min(offset, 0)
		logger.Printf("Info| %s: We heat up way too much, Cancelling positive the offset (%.2f).", hvac.Name, offset)
	}

	// Prevent trying to apply a temperature that is below the minimal accepted value by the hvac.
	if temperature < hvacMinimalHeatTemperature {
		logger.Printf("Info| %s: Triming the temperature to be above the minimal heating temperature.\n", hvac.Name)
		return hvacMinimalHeatTemperature, math.Max(0, offset), fastRamp
	}
	if temperature+offset < hvacMinimalHeatTemperature {
		logger.Printf("Info| %s: Triming the offset because it would result in a too low heating target.\n", hvac.Name)
		return temperature, math.Max(0, hvacMinimalHeatTemperature-temperature), fastRamp
	}
	return temperature, offset, fastRamp
}

func (s *Server) HeatUp() {
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
		hvac.DesiredState.Temperature, hvac.TemperatureOffset, hvac.FastRamp = DecideHeatUpTemperature(s.State.Rooms[roomName], hvac)
		hvac.DesiredState.Fan = DecideHeatUpFan(hvac.TemperatureOffset, hvac.Name)
	}
}

func (s *Server) DecideCoolDown() {
	logger.Println("Fail| DecideCoolDown is not implemented")
}
