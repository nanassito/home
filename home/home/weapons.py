import asyncio
import json
import logging
from datetime import datetime, timedelta
from typing import Deque

from fastapi import Request
from fastapi.responses import HTMLResponse

from home.mqtt import MQTTMessage, watch_mqtt_topic
from home.prometheus import COUNTER_NUM_RUNS, prom_query_one
from home.time import TimeZone, now
from home.utils import FeatureFlag
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)
from home.web import TEMPLATES, WEB

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

    async def soak(self: "Soaker", msg: MQTTMessage) -> None:
        if Soaker.FEATURE_FLAG.disabled:
            self.log.warning("Disabled, ignoring the trigger.")
            return
        if not json.loads(msg.payload)["occupancy"]:
            return
        if Soaker.SNOOZE_UNTIL >= now():
            self.log.warning(
                f"Snoozed until {Soaker.SNOOZE_UNTIL}, ignoring the trigger."
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


async def snooze_on_door_opening(msg: MQTTMessage) -> None:
    if not json.loads(msg.payload)["contact"]:
        await snooze(timedelta(minutes=5))


SOAKER_SIDE = Soaker(VALVE_BACKYARD_SIDE)
SOAKER_SCHOOL = Soaker(VALVE_BACKYARD_SCHOOL)
SOAKER_DECK = Soaker(VALVE_BACKYARD_DECK)


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
            watch_mqtt_topic("zigbee2mqtt/contact_livingroom", snooze_on_door_opening)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_bedroom", snooze_on_door_opening)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_garage", snooze_on_door_opening)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/contact_mower_back", snooze_on_door_opening)
        )

    @WEB.get("/soaker", response_class=HTMLResponse)
    async def get_soaker(request: Request):
        return TEMPLATES.TemplateResponse(
            "soaker.html.jinja",
            {
                "request": request,
                "page": "Soaker",
                "enabled": Soaker.FEATURE_FLAG.enabled,
                "snoozed_until": Soaker.SNOOZE_UNTIL.astimezone(
                    tz=TimeZone.PT.value
                ).isoformat()[:16],
                "last_runs": [
                    (ts.astimezone(tz=TimeZone.PT.value).isoformat()[:16], area)
                    for ts, area in Soaker.LAST_RUNS
                ],
            },
        )
