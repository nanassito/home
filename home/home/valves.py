import logging
from dataclasses import dataclass, field
from datetime import timedelta

from aioprometheus import Gauge

from home import facts
from home.model import Actionable
from home.mqtt import mqtt_send
from home.prometheus import prom_query_one

log = logging.getLogger(__name__)


@dataclass(frozen=True)
class ValveSection:
    area: str
    line: int
    base_runtime: timedelta = field(hash=False)


_PROM_VALVE = Gauge("valve_open", "0 the valve is closed, 1 the valve is opened.")


class BackyardValves(Actionable):
    LOG = log.getChild("Backyard")
    WATER_N_DAYS: int = 3
    SECTIONS = [
        ValveSection("side", 1, timedelta(minutes=5)),
        ValveSection("house", 2, timedelta(minutes=15)),
        ValveSection("school", 3, timedelta(minutes=15)),
        ValveSection("deck", 4, timedelta(minutes=15)),
    ]

    @property
    def prom_label(self):
        return "backyard_valves"

    async def get_desired_state(self: "BackyardValves") -> dict[ValveSection, bool]:
        if await facts.is_day_time():
            return {section: False for section in self.SECTIONS}
        weather_multiplier = 1.0
        for section in self.SECTIONS:
            promql = f'sum_over_time(mqtt_state_l{section.line}{{topic="zigbee2mqtt_valve_backyard"}}[{self.WATER_N_DAYS}d])'
            runtime = timedelta(minutes=await prom_query_one(promql))
            if runtime < section.base_runtime * weather_multiplier:
                return {s: (s == section) for s in self.SECTIONS}
        return {section: False for section in self.SECTIONS}

    async def get_current_state(self: "BackyardValves") -> dict[ValveSection, bool]:
        return {
            section: bool(
                await prom_query_one(
                    f'mqtt_state_l{section.line}{{topic="zigbee2mqtt_valve_backyard"}}'
                )
            )
            for section in self.SECTIONS
        }

    async def apply_state(
        self: "BackyardValves", state: dict[ValveSection, bool]
    ) -> None:
        for section, activate in state.items():
            value = "ON" if activate else "OFF"
            await mqtt_send(
                "zigbee2mqtt/valve_backyard/set", {f"state_l{section.line}": value}
            )
            self.LOG.info(f"Switched {section.area} {value.lower()}.")
            _PROM_VALVE.set({"area": section.area, "line": str(section.line)}, int(activate))
