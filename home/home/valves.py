import asyncio
import logging
from dataclasses import dataclass
from datetime import datetime, timedelta

from aioprometheus import Gauge
from fastapi import HTTPException
from pydantic import BaseModel

from home.facts import is_prod
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
    section: str
    area: str
    line: int

    def __post_init__(self: "Valve") -> None:
        self.log = log.getChild("Valve").getChild(self.area)
        self.water_until_requests: list[datetime] = []
        self.request_lock = asyncio.Lock()
        self.is_running = False

    @property
    def prom_query(self: "Valve") -> str:
        return f'mqtt_state_l{self.line}{{topic="zigbee2mqtt_valve_{self.section}"}}'

    async def water_for(self: "Valve", duration: timedelta) -> None:
        async with self.request_lock:
            self.water_until_requests.append(now() + duration)

    async def is_really_running(self: "Valve") -> bool:
        return bool(await prom_query_one(self.prom_query))

    async def switch(self: "Valve", should_be_running: bool) -> None:
        value = "ON" if should_be_running else "OFF"
        _PROM_VALVE.set(
            {"area": self.area, "line": str(self.line)}, int(should_be_running)
        )
        if is_prod():
            await mqtt_send(
                f"zigbee2mqtt/valve_{self.section}/set", {f"state_l{self.line}": value}
            )
        else:
            self.log.info(
                f"Fake mqtt send zigbee2mqtt/valve_{self.section}/set %s",
                {f"state_l{self.line}": value},
            )
        self.log.info(f"Switched {value}.")

    async def run_forever(self: "Valve") -> None:
        await self.switch(False)
        self.is_running = False
        while True:
            for _ in range(60):  # So that we pull external data only once a minute
                async with self.request_lock:
                    self.water_until_requests = [
                        until for until in self.water_until_requests if until > now()
                    ]
                should_run = bool(self.water_until_requests)
                if should_run != self.is_running:
                    await self.switch(should_run)
                    self.is_running = should_run
                await asyncio.sleep(1)
            self.is_running = await self.is_really_running()
            outstanding = [rq - now() for rq in self.water_until_requests]
            self.log.debug(f"Outstanding watering requests: {outstanding}")


VALVE_BACKYARD_SIDE = Valve("backyard", "side", 1)
VALVE_BACKYARD_HOUSE = Valve("backyard", "house", 2)
VALVE_BACKYARD_SCHOOL = Valve("backyard", "school", 3)
VALVE_BACKYARD_DECK = Valve("backyard", "deck", 4)
VALVE_FRONTYARD_STREET = Valve("frontyard", "street", 1)
VALVE_FRONTYARD_DRIVEWAY = Valve("frontyard", "driveway", 2)
VALVE_FRONTYARD_NEIGHBOR = Valve("frontyard", "neighbor", 3)
VALVE_FRONTYARD_PLANTER = Valve("frontyard", "planter", 4)


class _HttpValveRequest(BaseModel):
    area: str
    duration_sec: int


_HTTP_VALVE_MAPPING = {
    "side": VALVE_BACKYARD_SIDE,
    "house": VALVE_BACKYARD_HOUSE,
    "school": VALVE_BACKYARD_SCHOOL,
    "deck": VALVE_BACKYARD_DECK,
    "street": VALVE_FRONTYARD_STREET,
    "driveway": VALVE_FRONTYARD_DRIVEWAY,
    "neighbor": VALVE_FRONTYARD_NEIGHBOR,
    "planter": VALVE_FRONTYARD_PLANTER,
}


def init():
    @WEB.on_event("startup")
    def start_valves_controllers():
        all_valves = (
            VALVE_BACKYARD_DECK,
            VALVE_BACKYARD_HOUSE,
            VALVE_BACKYARD_SCHOOL,
            VALVE_BACKYARD_SIDE,
            VALVE_FRONTYARD_STREET,
            VALVE_FRONTYARD_DRIVEWAY,
            VALVE_FRONTYARD_NEIGHBOR,
            VALVE_FRONTYARD_PLANTER,
        )
        for valve in all_valves:
            asyncio.create_task(valve.run_forever())

    @WEB.post("/api/valve/activate")
    async def http_valve_activate(rq: _HttpValveRequest):
        if rq.area not in _HTTP_VALVE_MAPPING:
            return HTTPException(400, f"No known valve covering the {rq.area} area.")
        if not 0 <= rq.duration_sec <= 15 * 60:
            return HTTPException(400, "Duration must be between 0s and 15m.")
        await _HTTP_VALVE_MAPPING[rq.area].water_for(timedelta(seconds=rq.duration_sec))
