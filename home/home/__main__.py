import asyncio
import logging
from datetime import timedelta
from pathlib import Path

import uvicorn
import yaml
from fastapi import Request
from fastapi.exceptions import HTTPException
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel

import home.lawn
import home.prometheus
import home.weapons
from home.time import TimeZone, now
from home.web import WEB

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


home.weapons.init()
home.lawn.init()
home.prometheus.init()


TEMPLATES = Jinja2Templates(directory=str(Path("__file__").parent / "templates"))


@WEB.get("/", response_class=HTMLResponse)
async def get_index(request: Request):
    return TEMPLATES.TemplateResponse(
        "index.html",
        {
            "request": request,
            "app": {
                "uptime": str(
                    (now() - STARTUP) // timedelta(seconds=1) * timedelta(seconds=1)
                ),
            },
            "soaker": {
                "enabled": home.weapons.Soaker.FEATURE_FLAG.enabled,
                "last_runs": [
                    (ts.astimezone(tz=TimeZone.PT.value).isoformat()[:16], area)
                    for ts, area in home.weapons.Soaker.LAST_RUNS
                ],
            },
            "irrigation": {
                "enabled": home.lawn.Irrigation.FEATURE_FLAG.enabled,
            },
        },
    )


class _HttpFeatureFlag(BaseModel):
    enabled: bool
    target: str


@WEB.post("/api/feature_flag")
async def http_post_soaker(settings: _HttpFeatureFlag):
    targets = {
        "soaker": home.weapons.Soaker,
        "irrigation": home.lawn.Irrigation,
    }
    if settings.target.lower() not in targets:
        return HTTPException(400, detail=f"Invalid target: {settings.target}.")
    target = targets[settings.target.lower()]
    if settings.enabled:
        target.FEATURE_FLAG.enable()  # type: ignore
    else:
        target.FEATURE_FLAG.disable()  # type: ignore


uvicorn.run(WEB, host="0.0.0.0", port=8000, log_config=logging_cfg)
