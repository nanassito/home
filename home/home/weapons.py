import asyncio
import json
import logging
from datetime import timedelta

from home import facts
from home.prometheus import COUNTER_NUM_RUNS
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
)

log = logging.getLogger(__name__)


class Soaker:
    LOG = log.getChild("Soaker")
    ENABLED = True
    VALVES = (
        VALVE_BACKYARD_DECK,
        VALVE_BACKYARD_HOUSE,
        VALVE_BACKYARD_SCHOOL,
        VALVE_BACKYARD_SIDE,
    )
    DURATION = timedelta(seconds=10)

    @classmethod
    async def soak(cls: type["Soaker"], message: str) -> None:
        if not cls.ENABLED:
            return
        data = json.loads(message)
        if data["occupancy"]:
            if await facts.is_mower_running():
                return
            cls.LOG.info("Soaker turning on backyard sprinkler 1")
            COUNTER_NUM_RUNS.inc({"item": "Soaker"})
            valves_to_shutoff = [
                valve for valve in cls.VALVES if not valve.should_be_running
            ]
            for valve in cls.VALVES:
                await valve.switch_on()
            await asyncio.sleep(cls.DURATION.total_seconds())
            for valve in valves_to_shutoff:
                await valve.switch_off()
