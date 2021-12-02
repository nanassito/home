import logging
from dataclasses import dataclass, field
from datetime import timedelta

from aioprometheus import Gauge

from home import facts
from home.model import Actionable
from home.mqtt import mqtt_send
from home.prometheus import prom_query_one

log = logging.getLogger(__name__)

_PROM_VALVE = Gauge(
    "valve_should_be_running",
    "0 the valve should be closed, 1 the valve should be opened.",
)


class Valve:
    def __init__(self: "Valve", area: str, line: int) -> None:
        self.area = area
        self.line = line
        self.log = log.getChild("Valve").getChild(area)
        self.should_be_running = False

    def __repr__(self: "Valve") -> str:
        return (
            f"<{type(self).__module__}.{type(self).__name__} {self.area}|{self.line}>"
        )

    def __eq__(self, __o: object) -> bool:
        return isinstance(__o, Valve) and (self.area, self.line) == (__o.area, __o.line)

    async def is_running(self: "Valve") -> bool:
        return bool(
            await prom_query_one(
                f'mqtt_state_l{self.line}{{topic="zigbee2mqtt_valve_backyard"}}'
            )
        )

    async def switch_on(self: "Valve") -> None:
        self.should_be_running = True
        _PROM_VALVE.set(
            {"area": self.area, "line": str(self.line)}, self.should_be_running
        )
        await mqtt_send("zigbee2mqtt/valve_backyard/set", {f"state_l{self.line}": "ON"})
        self.log.info(f"Switched on.")

    async def switch_off(self: "Valve") -> None:
        self.should_be_running = False
        _PROM_VALVE.set(
            {"area": self.area, "line": str(self.line)}, self.should_be_running
        )
        await mqtt_send("zigbee2mqtt/valve_backyard/set", {f"state_l{self.line}": "ON"})
        self.log.info(f"Switched on.")


VALVE_BACKYARD_SIDE = Valve("side", 1)
VALVE_BACKYARD_HOUSE = Valve("house", 2)
VALVE_BACKYARD_SCHOOL = Valve("school", 3)
VALVE_BACKYARD_DECK = Valve("deck", 4)
