import asyncio
import logging
from pathlib import Path

import uvicorn
import yaml
from fastapi import Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates

import home.lawn
import home.prometheus
import home.weapons
from home.web import WEB

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)


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
        {"request": request, "soaker_enabled": home.weapons.Soaker.ENABLED},
    )


uvicorn.run(WEB, host="0.0.0.0", port=8000, log_config=logging_cfg)
