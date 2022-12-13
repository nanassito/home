package air

import (
	"fmt"
	"time"
)

func (s *Server) ApplySchedules() {
	now := fmt.Sprintf("%2d:%2d", time.Now().Hour(), time.Now().Minute())
	for name, weekSchedule := range s.State.Schedules {
		schedule := weekSchedule.Weekday
		if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
			schedule = weekSchedule.Weekend
		}
		for _, window := range schedule {
			if window.Start >= now && now < window.End {
				room := s.State.Rooms[weekSchedule.RoomName]
				if room.DesiredTemperatureRange != window.Settings {
					logger.Printf("Schedule %s changed Room %s to %v\n", name, room.Name, window.Settings)
					room.DesiredTemperatureRange = window.Settings
				}
			}
		}
	}
}
