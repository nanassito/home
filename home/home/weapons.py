import asyncio
import json
import logging
from datetime import timedelta

from home import facts
from home.mqtt import watch_mqtt_topic
from home.prometheus import COUNTER_NUM_RUNS
from home.time import now
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)
from home.web import API
from pydantic import BaseModel

log = logging.getLogger(__name__)


class Soaker:
    ENABLED = True
    DURATION = timedelta(seconds=10)
    ANTI_REBOUND = timedelta(minutes=2)

    def __init__(self: "Soaker", valve: Valve) -> None:
        self.log = log.getChild("Soaker").getChild(valve.area)
        self.valve = valve
        self.last_activation = now() - Soaker.ANTI_REBOUND

    async def soak(self: "Soaker", message: str) -> None:
        if not Soaker.ENABLED:
            return
        data = json.loads(message)
        if data["occupancy"]:
            if await facts.is_mower_running():
                return

            if self.last_activation + Soaker.ANTI_REBOUND > now():
                self.log.info("Anti-rebound is swallowing an activation.")
                return
            self.last_activation = now()

            COUNTER_NUM_RUNS.inc({"item": "Soaker"})
            self.log.info("Soaking!")
            should_shutoff = not self.valve.should_be_running
            await self.valve.switch_on()
            await asyncio.sleep(Soaker.DURATION.total_seconds())
            if should_shutoff:
                await self.valve.switch_off()


SOAKER_SIDE = Soaker(VALVE_BACKYARD_SIDE)
SOAKER_SCHOOL = Soaker(VALVE_BACKYARD_SCHOOL)
SOAKER_DECK = Soaker(VALVE_BACKYARD_DECK)


class _HttpSoakerSettings(BaseModel):
    enabled: bool

@API.post("/weapons/soaker")
async def http_post_soaker(settings: _HttpSoakerSettings):
    Soaker.ENABLED = settings.enabled


@API.get("/weapons/soaker")
async def http_get_soaker() -> _HttpSoakerSettings:
    return _HttpSoakerSettings(enabled=Soaker.ENABLED)


def init():
    @API.on_event("startup")
    def _():
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_side", SOAKER_SIDE.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_back", SOAKER_SCHOOL.soak)
        )
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/motion_back", SOAKER_DECK.soak)
        )
