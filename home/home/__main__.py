import asyncio
import logging
from pathlib import Path
from time import time_ns

import uvicorn
import yaml
from aioprometheus import Histogram
from fastapi import Request
from fastapi.exceptions import HTTPException
from fastapi.responses import RedirectResponse
from pydantic import BaseModel

import home.air
import home.lawn
import home.mqtt
import home.music
import home.prometheus
import home.valves
import home.weapons
from home.time import now
from home.web import WEB

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)
STARTUP = now()
_PROM_ASYNCIO_LATENCY = Histogram(
    "asyncio_latency_ns",
    "ns level deviance between the asyncio event loop and the wall clock.",
    buckets=[10 ** i for i in range(3, 9)],
)


@WEB.on_event("startup")
def _():
    def shutdown_on_error(loop, context):
        loop.set_debug(True)
        loop.default_exception_handler(context)
        loop.stop()

    async def monitor_event_loop_latency():
        while True:
            before = time_ns()
            await asyncio.sleep(1)
            _PROM_ASYNCIO_LATENCY.observe({}, time_ns() - before - 1_000_000_000)

    asyncio.get_event_loop().set_exception_handler(shutdown_on_error)
    asyncio.create_task(monitor_event_loop_latency())


home.valves.init()
home.weapons.init()
home.lawn.init()
home.prometheus.init()
home.music.init()
home.air.init()
home.mqtt.init()


@WEB.get("/", response_class=RedirectResponse)
async def get_index(request: Request):
    return "/temperature"


class _HttpFeatureFlag(BaseModel):
    enabled: bool
    target: str


@WEB.post("/api/feature_flag")
async def http_post_feature_flag(settings: _HttpFeatureFlag):
    targets = {
        "soaker": home.weapons.Soaker,
        "irrigation": home.lawn.Irrigation,
        "hvac": home.air.HvacController,
    }
    if settings.target.lower() not in targets:
        return HTTPException(400, f"Invalid target: {settings.target}.")
    target = targets[settings.target.lower()]
    if settings.enabled:
        target.FEATURE_FLAG.enable()  # type: ignore
    else:
        target.FEATURE_FLAG.disable()  # type: ignore


uvicorn.run(WEB, host="0.0.0.0", port=8000, log_config=logging_cfg)
