import asyncio
import logging
from pathlib import Path

import uvicorn
import yaml
from fastapi import FastAPI

import home.lawn
import home.prometheus
import home.weapons
from home.lawn import Irrigation
from home.model import Actionable

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)


LOOPERS: list[type[Actionable]] = [
    Irrigation,
]
WEB = FastAPI()


@WEB.on_event("startup")
def _():
    def shutdown_on_error(loop, context):
        loop.default_exception_handler(context)
        loop.stop()

    asyncio.get_event_loop().set_exception_handler(shutdown_on_error)


home.weapons.init(WEB)
home.lawn.init(WEB)
home.prometheus.init(WEB)

# @WEB.get("/users/{user_id}")
# async def get_user(
#     request: Request,
#     user_id: str,
# ):
#     request.app.state.users_events_counter.inc({"path": request.scope["path"]})
#     return Response(f"{user_id}")

uvicorn.run(WEB, port=8000, log_config=logging_cfg)
