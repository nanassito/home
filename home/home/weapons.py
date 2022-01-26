import asyncio
import json
import logging
from datetime import datetime, timedelta
from http.client import HTTPException
from typing import Deque

from pydantic import BaseModel

from home.mqtt import watch_mqtt_topic
from home.prometheus import COUNTER_NUM_RUNS, prom_query_one
from home.time import now
from home.utils import FeatureFlag
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)
from home.web import WEB

log = logging.getLogger(__name__)


class Soaker:
    FEATURE_FLAG = FeatureFlag("Soaker")
    DURATION = timedelta(seconds=10)
    ANTI_REBOUND = timedelta(minutes=2)
    SNOOZE_UNTIL = now() - timedelta(minutes=1)
    LAST_RUNS: Deque[tuple[datetime, str]] = Deque([], maxlen=10)

    def __init__(self: "Soaker", valve: Valve) -> None:
        self.log = log.getChild("Soaker").getChild(valve.area)
        self.valve = valve
        self.last_activation = now() - Soaker.ANTI_REBOUND

    async def soak(self: "Soaker", message: str) -> None:
        if Soaker.FEATURE_FLAG.disabled:
            self.log.warning("Disabled, ignoring the trigger.")
            return
        if not json.loads(message)["occupancy"]:
            return
        if Soaker.SNOOZE_UNTIL >= now():
            self.log.warning(
                "Snoozed until {Soaker.SNOOZE_UNTIL}, ignoring the trigger."
            )
            return
        if self.last_activation + Soaker.ANTI_REBOUND > now():
            self.log.info("Anti-rebound, ignoring the trigger.")
            return
        if await prom_query_one("min(mqtt_contact)") == 0:
            self.log.info("A door is opened, ignoring the trigger.")
            return
        self.last_activation = now()

        COUNTER_NUM_RUNS.inc({"item": "Soaker", "soaker": self.valve.area})
        Soaker.LAST_RUNS.append((now(), self.valve.area))
        self.log.info("Soaking!")
        await self.valve.water_for(timedelta(seconds=10))


async def snooze(ttl: timedelta) -> None:
    Soaker.SNOOZE_UNTIL = now() + ttl
    log.info(f"Snoozing the soakers for {ttl}.")


async def snooze_on_door_opening(message: str) -> None:
    if not json.loads(message)["contact"]:
        await snooze(timedelta(minutes=5))


SOAKER_SIDE = Soaker(VALVE_BACKYARD_SIDE)
SOAKER_SCHOOL = Soaker(VALVE_BACKYARD_SCHOOL)
SOAKER_DECK = Soaker(VALVE_BACKYARD_DECK)


class _HttpSoakerSnooze(BaseModel):
    ttl_minutes: int


def init():
    @WEB.on_event("startup")
    def _():
        Soaker.FEATURE_FLAG.enable()
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_side", SOAKER_SIDE.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_garage", SOAKER_SIDE.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_back", SOAKER_SCHOOL.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_pergola", SOAKER_DECK.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_livingroom", snooze_on_door_opening)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_bedroom", snooze_on_door_opening)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_garage", snooze_on_door_opening)
        )

    @WEB.post("/api/soaker/snooze")
    async def http_post_soaker_snooze(settings: _HttpSoakerSnooze):
        if settings.ttl_minutes <= 0:
            return HTTPException(400, "Snooze must be a positive number.")
        await snooze(timedelta(minutes=settings.ttl_minutes))
