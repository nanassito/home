import asyncio
import json
import logging
from datetime import timedelta

from home import facts
from home.prometheus import COUNTER_NUM_RUNS
from home.valves import VALVE_BACKYARD_SCHOOL, VALVE_BACKYARD_SIDE, Valve

log = logging.getLogger(__name__)


class Soaker:
    def __init__(self: "Soaker", valve: Valve) -> None:
        self.log = log.getChild("Soaker")
        self.enabled = True
        self.valve = valve
        self.duration = timedelta(seconds=10)

    async def soak(self: "Soaker", message: str) -> None:
        if not self.enabled:
            return
        data = json.loads(message)
        if data["occupancy"]:
            if await facts.is_mower_running():
                return
            COUNTER_NUM_RUNS.inc({"item": "Soaker"})
            self.log.info("Soaking!")
            should_shutoff = not self.valve.should_be_running
            await self.valve.switch_on()
            await asyncio.sleep(self.duration.total_seconds())
            if should_shutoff:
                await self.valve.switch_off()


SOAKER_SIDE = Soaker(VALVE_BACKYARD_SIDE)
SOAKER_BACK = Soaker(VALVE_BACKYARD_SCHOOL)