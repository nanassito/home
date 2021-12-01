import logging
from dataclasses import dataclass
from datetime import timedelta

from home import facts
from home.model import Actionable
from home.prometheus import prom_query_one
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)

log = logging.getLogger(__name__)


@dataclass
class Schedule:
    water_time: timedelta
    over: timedelta


class BackyardIrrigation(Actionable):
    ENABLED = True
    LOG = log.getChild("BackyardIrrigation")
    SCHEDULE = {
        VALVE_BACKYARD_SIDE: Schedule(timedelta(minutes=5), timedelta(days=7)),
        VALVE_BACKYARD_SCHOOL: Schedule(timedelta(minutes=15), timedelta(days=3)),
        VALVE_BACKYARD_HOUSE: Schedule(timedelta(minutes=15), timedelta(days=3)),
        VALVE_BACKYARD_DECK: Schedule(timedelta(minutes=15), timedelta(days=3)),
    }

    @property
    def prom_label(self: "BackyardIrrigation") -> str:
        return "BackyardIrrigation"

    @classmethod
    async def get_desired_state(cls: type["BackyardIrrigation"]) -> dict[Valve, bool]:
        if any([await facts.is_day_time(), await facts.is_mower_running()]):
            return {section: False for section in cls.SCHEDULE}
        for valve, schedule in cls.SCHEDULE.items():
            promql = f'sum_over_time(mqtt_state_l{valve.line}{{topic="zigbee2mqtt_valve_backyard"}}[{schedule.over.days}d])'
            runtime = timedelta(minutes=await prom_query_one(promql))
            if runtime < schedule.water_time:
                return {v: (v == valve) for v in cls.SCHEDULE}
        return {v: False for v in cls.SCHEDULE}

    @classmethod
    async def get_current_state(cls: type["BackyardIrrigation"]) -> dict[Valve, bool]:
        return {valve: await valve.is_running() for valve in cls.SCHEDULE}

    @classmethod
    async def apply_state(
        cls: type["BackyardIrrigation"], state: dict[Valve, bool]
    ) -> None:
        if not cls.ENABLED:
            cls.LOG.warning("BackyardIrrigation is disabled.")
            return
        for valve, should_run in state.items():
            if should_run:
                await valve.switch_on()
            else:
                await valve.switch_off()