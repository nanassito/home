import asyncio
import json
import logging
from dataclasses import dataclass
from datetime import datetime, timedelta

import aiohttp
from aioprometheus import Gauge
from fastapi import HTTPException
from pydantic import BaseModel

from home.facts import is_prod
from home.mqtt import MQTTMessage, mqtt_send, watch_mqtt_topic
from home.prometheus import prom_query_one
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
    switch_id: str

    def __post_init__(self: "Valve") -> None:
        self.log = log.getChild("Valve").getChild(self.area)
        self.water_until_requests: list[datetime] = []
        self.request_lock = asyncio.Lock()

    @property
    def prom_query(self: "Valve") -> str:
        return f'mqtt_state_l{self.line}{{location="{self.section}", type="valve"}}'

    async def water_for(self: "Valve", duration: timedelta) -> None:
        if is_prod():
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    "http://192.168.1.1:7003/activate",
                    data={
                        "SwitchID": self.switch_id,
                        "DurationSeconds": int(duration.total_seconds()),
                        "ClientID": "home",
                    },
                ) as resp:
                    self.log.info(f"{resp.status} - {await resp.text()}")

    async def is_really_running(self: "Valve") -> bool:
        return bool(await prom_query_one(self.prom_query))


VALVE_BACKYARD_SIDE = Valve("backyard", "side", 1, "valve_backyard_side")
VALVE_BACKYARD_HOUSE = Valve("backyard", "house", 2, "valve_backyard_house")
VALVE_BACKYARD_SCHOOL = Valve("backyard", "school", 3, "valve_backyard_school")
VALVE_BACKYARD_DECK = Valve("backyard", "deck", 4, "valve_backyard_deck")
VALVE_FRONTYARD_STREET = Valve("frontyard", "street", 1, "valve_frontyard_street")
VALVE_FRONTYARD_DRIVEWAY = Valve("frontyard", "driveway", 2, "valve_frontyard_driveway")
VALVE_FRONTYARD_NEIGHBOR = Valve("frontyard", "neighbor", 3, "valve_frontyard_neighbor")
VALVE_FRONTYARD_PLANTER = Valve("frontyard", "planter", 4, "valve_frontyard_planter")


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
    @WEB.post("/api/valve/activate")
    async def http_valve_activate(rq: _HttpValveRequest):
        if rq.area not in _HTTP_VALVE_MAPPING:
            return HTTPException(400, f"No known valve covering the {rq.area} area.")
        if not 0 <= rq.duration_sec <= 15 * 60:
            return HTTPException(400, "Duration must be between 0s and 15m.")
        await _HTTP_VALVE_MAPPING[rq.area].water_for(timedelta(seconds=rq.duration_sec))
