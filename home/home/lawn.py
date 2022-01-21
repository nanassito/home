import asyncio
import logging
from dataclasses import dataclass
from datetime import timedelta

from fastapi import HTTPException
from pydantic import BaseModel

from home import facts
from home.prometheus import prom_query_one
from home.time import now
from home.utils import FeatureFlag
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)
from home.web import WEB

log = logging.getLogger(__name__)


@dataclass
class Schedule:
    water_time: timedelta
    over: timedelta


class Irrigation:
    FEATURE_FLAG = FeatureFlag("Irrigation")
    LOG = log.getChild("Irrigation")
    SCHEDULE = {
        VALVE_BACKYARD_SIDE: Schedule(timedelta(minutes=5), timedelta(days=7)),
        VALVE_BACKYARD_SCHOOL: Schedule(timedelta(minutes=5), timedelta(days=3)),
        VALVE_BACKYARD_HOUSE: Schedule(timedelta(minutes=5), timedelta(days=3)),
        VALVE_BACKYARD_DECK: Schedule(timedelta(minutes=5), timedelta(days=3)),
    }

    async def run_forever(self: "Irrigation") -> None:
        while True:
            if await facts.is_night_time():
                for valve, schedule in Irrigation.SCHEDULE.items():
                    promql = f'sum without(instance) (sum_over_time(mqtt_state_l{valve.line}{{topic="zigbee2mqtt_valve_backyard"}}[{schedule.over.days}d]))'
                    runtime = timedelta(minutes=await prom_query_one(promql))
                    self.LOG.debug(f"{valve} has had {runtime} of water out of {schedule.water_time}")
                    if runtime < schedule.water_time / 2:
                        duration = schedule.water_time - runtime
                        self.LOG.info(
                            f"Irrigation requesting {duration} of water on {valve}"
                        )
                        valve.water_for(duration)
                        break
            await asyncio.sleep(60)


class _HttpIrrigationValveSettings(BaseModel):
    water_time_minutes: int
    over_days: int


class _HttpIrrigationSettings(BaseModel):
    valves: dict[str, _HttpIrrigationValveSettings]


@WEB.post("/api/lawn/irrigation")
async def http_update_irrigation_settings(settings: _HttpIrrigationSettings):
    schedule = {valve.area: setting for valve, setting in Irrigation.SCHEDULE.items()}
    unknown_valves = settings.valves.keys() - schedule.keys()
    if unknown_valves:
        raise HTTPException(400, f"Unknown valves {unknown_valves}")
    for area, setting in settings.valves.items():
        schedule[area].water_time = timedelta(minutes=setting.water_time_minutes)
        schedule[area].over = timedelta(days=setting.over_days)
        # TODO: Find a way to expose this to Prometheus


def init() -> None:
    @WEB.on_event("startup")
    def _():
        Irrigation.FEATURE_FLAG.enable()
        asyncio.create_task(Irrigation().run_forever())
