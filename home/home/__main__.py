import asyncio
import logging
from pathlib import Path
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
import uvicorn
import yaml

import home.lawn
import home.prometheus
import home.weapons
from home.web import API


with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)


@API.on_event("startup")
def _():
    def shutdown_on_error(loop, context):
        loop.default_exception_handler(context)
        loop.stop()

    asyncio.get_event_loop().set_exception_handler(shutdown_on_error)


home.weapons.init()
home.lawn.init()
home.prometheus.init()

web = FastAPI()
web.mount("/api", API, name="api")
web.mount(
    "/",
    StaticFiles(directory=str(Path("__file__").parent / "web"), html=True),
    name="static",
)
uvicorn.run(web, port=8000, log_config=logging_cfg)
