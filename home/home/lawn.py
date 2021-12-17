import asyncio
import logging
from dataclasses import dataclass
from datetime import timedelta

from home import facts
from home.model import Actionable
from home.prometheus import prom_query_one
from home.time import now
from home.utils import FeatureFlag
from pydantic import BaseModel
from fastapi import HTTPException
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


class Irrigation(Actionable):
    FEATURE_FLAG = FeatureFlag("Irrigation")
    LOG = log.getChild("Irrigation")
    SCHEDULE = {
        VALVE_BACKYARD_SIDE: Schedule(timedelta(minutes=0), timedelta(days=30)),
        VALVE_BACKYARD_SCHOOL: Schedule(timedelta(minutes=0), timedelta(days=30)),
        VALVE_BACKYARD_HOUSE: Schedule(timedelta(minutes=0), timedelta(days=30)),
        VALVE_BACKYARD_DECK: Schedule(timedelta(minutes=0), timedelta(days=30)),
    }

    @property
    def prom_label(self: "Irrigation") -> str:
        return "BackyardIrrigation"

    @classmethod
    async def get_desired_state(cls: type["Irrigation"]) -> dict[Valve, bool]:
        if any([await facts.is_day_time(), await facts.is_mower_running()]):
            return {section: False for section in cls.SCHEDULE}
        for valve, schedule in cls.SCHEDULE.items():
            promql = f'sum without(instance) (sum_over_time(mqtt_state_l{valve.line}{{topic="zigbee2mqtt_valve_backyard"}}[{schedule.over.days}d]))'
            runtime = timedelta(minutes=await prom_query_one(promql))
            if runtime < schedule.water_time and valve.should_be_running:
                # Valve is already running so let's keep it going.
                return {v: (v == valve) for v in cls.SCHEDULE}
            if runtime < schedule.water_time / 2:
                # Valve isn't running so we'll wait a bit more to avoid watering a single minute.
                return {v: (v == valve) for v in cls.SCHEDULE}
        return {v: False for v in cls.SCHEDULE}

    @classmethod
    async def get_current_state(cls: type["Irrigation"]) -> dict[Valve, bool]:
        return {valve: await valve.is_running() for valve in cls.SCHEDULE}

    @classmethod
    async def apply_state(cls: type["Irrigation"], state: dict[Valve, bool]) -> None:
        cls.LOG.debug(f"apply_state({state})")
        if cls.FEATURE_FLAG.disabled:
            cls.LOG.warning("Irrigation is disabled.")
            return
        cls.LOG.info("Applying changes on the backyard valves.")
        for valve, should_run in state.items():
            if should_run:
                await valve.switch_on()
            else:
                await valve.switch_off()


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
        cycle = timedelta(minutes=1)

        async def controller_main_loop():
            while True:
                before = now()

                desired_state = await Irrigation.get_desired_state()
                if desired_state != await Irrigation.get_current_state():
                    await Irrigation.apply_state(desired_state)

                after = now()
                duration = after - before
                Irrigation.RUNTIME_MS_GAUGE.set(
                    {"looper": "Irrigation"}, duration.total_seconds() * 1000
                )
                if duration > cycle:
                    log.warning(f"Full cycle took {duration - cycle} too long.")
                await asyncio.sleep((cycle - duration % cycle).total_seconds())

        asyncio.create_task(controller_main_loop())
