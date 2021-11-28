import asyncio
import logging
from datetime import timedelta
from pathlib import Path

import uvicorn
import yaml
from aioprometheus import MetricsMiddleware
from aioprometheus.asgi.starlette import metrics
from fastapi import FastAPI

from home.model import Actionable
from home.time import now
from home.valves import BackyardValves

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)


LOOPERS: list[Actionable] = [
    BackyardValves(),
]
CYCLE = timedelta(minutes=1)
WEB = FastAPI()


@WEB.on_event("startup")
def init_controller():
    async def controller_main_loop():
        while True:
            before_all = now()
            for looper in LOOPERS:
                before_one = now()
                desired_state = await looper.get_desired_state()
                if desired_state != await looper.get_current_state():
                    await looper.apply_state(desired_state)
                after_one = now()
                duration_one_ms = (after_one - before_one).total_seconds() * 1000
                looper.RUNTIME_MS_GAUGE.set({"looper": looper.prom_label}, duration_one_ms)
            after_all = now()
            duration_all = after_all - before_all
            if duration_all > CYCLE:
                log.warning(f"Full cycle took {duration_all - CYCLE} too long.")
            await asyncio.sleep((CYCLE - duration_all % CYCLE).total_seconds())

    asyncio.create_task(controller_main_loop())


# Any custom application metrics are automatically included in the exposed
# metrics. It is a good idea to attach the metrics to 'app.state' so they
# can easily be accessed in the route handler - as metrics are often
# created in a different module than where they are used.
# WEB.state.users_events_counter = Counter("events", "Number of events.")

WEB.add_middleware(MetricsMiddleware)
WEB.add_route("/metrics", metrics)


# @WEB.get("/users/{user_id}")
# async def get_user(
#     request: Request,
#     user_id: str,
# ):
#     request.app.state.users_events_counter.inc({"path": request.scope["path"]})
#     return Response(f"{user_id}")

uvicorn.run(WEB, port=8000, log_config=logging_cfg)
