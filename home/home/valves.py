import asyncio
import logging
from dataclasses import dataclass

from aioprometheus import Gauge

from home.mqtt import mqtt_send
from home.prometheus import prom_query_one
from home.web import WEB

log = logging.getLogger(__name__)

_PROM_VALVE = Gauge(
    "valve_should_be_running",
    "0 the valve should be closed, 1 the valve should be opened.",
)


@dataclass(unsafe_hash=True)
class Valve:
    area: str
    line: int

    def __post_init__(self: "Valve") -> None:
        self.log = log.getChild("Valve").getChild(self.area)
        self.should_be_running = False

    async def is_running(self: "Valve") -> bool:
        return bool(
            await prom_query_one(
                f'mqtt_state_l{self.line}{{topic="zigbee2mqtt_valve_backyard"}}'
            )
        )

    async def _switch(self: "Valve", should_be_running: bool) -> None:
        self.should_be_running = should_be_running
        value = "ON" if should_be_running else "OFF"
        _PROM_VALVE.set(
            {"area": self.area, "line": str(self.line)}, self.should_be_running
        )
        await mqtt_send(
            "zigbee2mqtt/valve_backyard/set", {f"state_l{self.line}": value}
        )
        self.log.info(f"Switched {value}.")

    async def switch_on(self: "Valve") -> None:
        await self._switch(True)

    async def switch_off(self: "Valve") -> None:
        await self._switch(False)


VALVE_BACKYARD_SIDE = Valve("side", 1)
VALVE_BACKYARD_HOUSE = Valve("house", 2)
VALVE_BACKYARD_SCHOOL = Valve("school", 3)
VALVE_BACKYARD_DECK = Valve("deck", 4)


def init():
    @WEB.on_event("startup")
    async def ensure_valves_off():
        await asyncio.gather(
            *[
                valve.switch_off()
                for valve in (
                    VALVE_BACKYARD_DECK,
                    VALVE_BACKYARD_HOUSE,
                    VALVE_BACKYARD_HOUSE,
                    VALVE_BACKYARD_SIDE,
                )
            ]
        )
