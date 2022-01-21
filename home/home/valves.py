import asyncio
from datetime import datetime, timedelta
import logging
from dataclasses import dataclass

from aioprometheus import Gauge

from home.mqtt import mqtt_send
from home.prometheus import prom_query_one
from home.time import now
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
        self.water_until_requests: list[datetime] = []

    def water_for(self: "Valve", duration: timedelta) -> None:
        self.water_until_requests.append(now() + duration)

    async def is_really_running(self: "Valve") -> bool:
        return bool(
            await prom_query_one(
                f'mqtt_state_l{self.line}{{topic="zigbee2mqtt_valve_backyard"}}'
            )
        )

    async def switch(self: "Valve", should_be_running: bool) -> None:
        value = "ON" if should_be_running else "OFF"
        _PROM_VALVE.set({"area": self.area, "line": str(self.line)}, should_be_running)
        await mqtt_send(
            "zigbee2mqtt/valve_backyard/set", {f"state_l{self.line}": value}
        )
        self.log.info(f"Switched {value}.")

    async def run_forever(self: "Valve") -> None:
        await self.switch(False)
        is_running = False
        while True:
            for _ in range(60):  # So that we pull external data only once a minute
                self.water_until_requests = [
                    until for until in self.water_until_requests if until > now()
                ]
                should_run = bool(self.water_until_requests)
                if should_run != is_running:
                    await self.switch(should_run)
                    is_running = should_run
                await asyncio.sleep(1)
            is_running = await self.is_really_running()


VALVE_BACKYARD_SIDE = Valve("side", 1)
VALVE_BACKYARD_HOUSE = Valve("house", 2)
VALVE_BACKYARD_SCHOOL = Valve("school", 3)
VALVE_BACKYARD_DECK = Valve("deck", 4)


def init():
    @WEB.on_event("startup")
    def start_valves_controllers():
        all_valves = (
            VALVE_BACKYARD_DECK,
            VALVE_BACKYARD_HOUSE,
            VALVE_BACKYARD_HOUSE,
            VALVE_BACKYARD_SIDE,
        )
        for valve in all_valves:
            asyncio.create_task(valve.run_forever())
