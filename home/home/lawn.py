import asyncio
from dataclasses import dataclass
import logging
from datetime import timedelta
from typing import Optional

from fastapi import HTTPException
from pydantic import BaseModel

from home import facts
from home.model import Actionable
from home.prometheus import prom_query_one
from home.time import now
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    Valve,
)
from home.web import API, WEB
from starlette.responses import RedirectResponse

log = logging.getLogger(__name__)


@dataclass
class Schedule:
    water_time: timedelta
    over: timedelta


class Irrigation(Actionable):
    ENABLED = True
    LOG = log.getChild("Irrigation")
    SCHEDULE = {
        VALVE_BACKYARD_SIDE: Schedule(timedelta(minutes=5), timedelta(days=7)),
        VALVE_BACKYARD_SCHOOL: Schedule(timedelta(minutes=10), timedelta(days=5)),
        VALVE_BACKYARD_HOUSE: Schedule(timedelta(minutes=10), timedelta(days=5)),
        VALVE_BACKYARD_DECK: Schedule(timedelta(minutes=10), timedelta(days=5)),
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
            if runtime < schedule.water_time:
                return {v: (v == valve) for v in cls.SCHEDULE}
        return {v: False for v in cls.SCHEDULE}

    @classmethod
    async def get_current_state(cls: type["Irrigation"]) -> dict[Valve, bool]:
        return {valve: await valve.is_running() for valve in cls.SCHEDULE}

    @classmethod
    async def apply_state(cls: type["Irrigation"], state: dict[Valve, bool]) -> None:
        if not cls.ENABLED:
            cls.LOG.warning("Irrigation is disabled.")
            return
        cls.LOG.info("Applying changes on the backyard valves.")
        for valve, should_run in state.items():
            if should_run:
                await valve.switch_on()
            else:
                await valve.switch_off()


class _HttpSchedule(BaseModel):
    water_time_minutes: int
    over_days: int


class _HttpIrrigation(BaseModel):
    enabled: Optional[bool]
    valves: Optional[dict[str, _HttpSchedule]]


@API.get("/lawn/irrigation", response_model=_HttpIrrigation)
async def http_get_irrigation() -> _HttpIrrigation:
    return _HttpIrrigation(
        enabled=Irrigation.ENABLED,
        valves={
            valve.area: _HttpSchedule(
                water_time_minutes=int(schedule.water_time.total_seconds() / 60),
                over_days=schedule.over.days,
            )
            for valve, schedule in Irrigation.SCHEDULE.items()
        },
    )


@API.post("/lawn/irrigation", response_model=_HttpIrrigation)
async def http_post_irrigation(config: _HttpIrrigation) -> RedirectResponse:
    if config.enabled is not None:
        Irrigation.ENABLED = config.enabled
    area2valve = {valve.area: valve for valve in Irrigation.SCHEDULE}
    for area, schedule in (config.valves or {}).items():
        if area not in area2valve:
            raise HTTPException(
                status_code=422,
                detail=f"{area} isn't a valve that is part of the schedule.",
            )
        valve = area2valve[area]
        Irrigation.SCHEDULE[valve] = Schedule(
            water_time=timedelta(minutes=schedule.water_time_minutes),
            over=timedelta(days=schedule.over_days),
        )
    return RedirectResponse("/")


def init() -> None:
    @WEB.on_event("startup")
    def _():
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
