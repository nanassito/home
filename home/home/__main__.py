import asyncio
import logging
from datetime import timedelta
from pathlib import Path

import uvicorn
import yaml
from fastapi import Request
from fastapi.exceptions import HTTPException
from fastapi.responses import HTMLResponse
from pydantic import BaseModel

import home.lawn
import home.prometheus
import home.valves
import home.weapons
import home.music
from home.facts import is_prod
from home.time import TimeZone, now
from home.web import WEB, TEMPLATES

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)
STARTUP = now()


@WEB.on_event("startup")
def _():
    def shutdown_on_error(loop, context):
        loop.default_exception_handler(context)
        loop.stop()

    asyncio.get_event_loop().set_exception_handler(shutdown_on_error)


home.valves.init()
home.weapons.init()
home.lawn.init()
home.prometheus.init()
home.music.init()


class _HttpFeatureFlag(BaseModel):
    enabled: bool
    target: str


@WEB.post("/api/feature_flag")
async def http_post_feature_flag(settings: _HttpFeatureFlag):
    targets = {
        "soaker": home.weapons.Soaker,
        "irrigation": home.lawn.Irrigation,
    }
    if settings.target.lower() not in targets:
        return HTTPException(400, f"Invalid target: {settings.target}.")
    target = targets[settings.target.lower()]
    if settings.enabled:
        target.FEATURE_FLAG.enable()  # type: ignore
    else:
        target.FEATURE_FLAG.disable()  # type: ignore


uvicorn.run(WEB, host="0.0.0.0", port=8000, log_config=logging_cfg)
