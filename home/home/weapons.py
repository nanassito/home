import asyncio
import json
import logging
from datetime import timedelta

from home import facts
from home.prometheus import COUNTER_NUM_RUNS
from home.time import now
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)

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
