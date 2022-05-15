import asyncio
import logging
from base64 import b64encode
from collections import deque
from dataclasses import dataclass, field
from datetime import timedelta
from enum import Enum

import plotly.express as px
from fastapi import Request
from fastapi.responses import HTMLResponse
from pandas import DataFrame

from home import facts
from home.prometheus import prom_query_one, prom_query_series
from home.utils import FeatureFlag
from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
    VALVE_FRONTYARD_DRIVEWAY,
    VALVE_FRONTYARD_NEIGHBOR,
    VALVE_FRONTYARD_PLANTER,
    VALVE_FRONTYARD_STREET,
    Valve,
)
from home.web import TEMPLATES, WEB

log = logging.getLogger(__name__)


class ScheduleModifiers(Enum):
    HUMIDITY = "humidity"


@dataclass
class Schedule:
    water_time: timedelta
    over: timedelta
    modifiers: set[ScheduleModifiers] = field(
        default_factory=lambda: {
            ScheduleModifiers.HUMIDITY,
        }
    )


class Irrigation:
    FEATURE_FLAG = FeatureFlag("Irrigation")
    LOG = log.getChild("Irrigation")
    SCHEDULE = {
        VALVE_BACKYARD_SIDE: Schedule(timedelta(minutes=5), timedelta(days=2)),
        VALVE_BACKYARD_SCHOOL: Schedule(timedelta(minutes=4), timedelta(days=2)),
        VALVE_BACKYARD_HOUSE: Schedule(timedelta(minutes=8), timedelta(days=2)),
        VALVE_BACKYARD_DECK: Schedule(timedelta(minutes=7), timedelta(days=2)),
        VALVE_FRONTYARD_STREET: Schedule(timedelta(minutes=8), timedelta(days=2)),
        VALVE_FRONTYARD_DRIVEWAY: Schedule(timedelta(minutes=5), timedelta(days=2)),
        VALVE_FRONTYARD_NEIGHBOR: Schedule(timedelta(minutes=8), timedelta(days=2)),
        VALVE_FRONTYARD_PLANTER: Schedule(timedelta(minutes=15), timedelta(days=2)),
    }

    async def run_forever(self: "Irrigation") -> None:
        while True:
            if await facts.is_night_time():
                # Sigmoid tuned to that we water every day during drought and water half as often when super humid
                promql = '2 / (1 + 2.71828 ^ -( avg_over_time(mqtt_humidity{topic="zigbee2mqtt_air_outside"}[7d]) * 0.05 -2.5) )'
                dryness_factor = await prom_query_one(promql)
                self.LOG.info(f"Humidity factor is {round(dryness_factor, 3)}")
                for valve, schedule in Irrigation.SCHEDULE.items():
                    hours = round(schedule.over.days * 24)
                    promql = f"sum without(instance) (sum_over_time({valve.prom_query}[{hours}h]))"
                    runtime = timedelta(minutes=await prom_query_one(promql))
                    self.LOG.debug(
                        f"{valve} has had {runtime} of water out of {schedule.water_time}"
                    )
                    if runtime < schedule.water_time / 2:
                        self.LOG.info(
                            f"Requesting {schedule.water_time} of water from {valve}"
                        )
                        await valve.water_for(schedule.water_time)
                        break
                    if valve.is_running:
                        break  # If the valve is running we don't want to start another one.
            await asyncio.sleep(60)


def init() -> None:
    @WEB.on_event("startup")
    def _():
        Irrigation.FEATURE_FLAG.enable()
        asyncio.create_task(Irrigation().run_forever())

    @WEB.get("/irrigation", response_class=HTMLResponse)
    async def get_irrigation(request: Request):
        async def get_valve_history(valve: Valve) -> DataFrame:
            df = DataFrame(
                await prom_query_series(valve.prom_query, timedelta(days=7)),
                columns=("ts", valve.area),
            ).set_index("ts")
            df["date"] = df.index.date
            return df.groupby("date").sum()

        valves = deque(Irrigation.SCHEDULE)
        history = await get_valve_history(valves.pop())
        while valves:
            history = history.merge(await get_valve_history(valves.pop()), on="date")
        return TEMPLATES.TemplateResponse(
            "irrigation.html.jinja",
            {
                "request": request,
                "page": "Irrigation",
                "enabled": Irrigation.FEATURE_FLAG.enabled,
                "history": b64encode(
                    px.imshow(
                        history, color_continuous_scale="BuGn", zmin=0, zmax=10
                    ).to_image("webp")
                ).decode(),
            },
        )
