import asyncio
import logging
from datetime import timedelta
from pathlib import Path

import uvicorn
import yaml
from aioprometheus import Counter, MetricsMiddleware
from aioprometheus.asgi.starlette import metrics
from fastapi import FastAPI, Request, Response

from home.time import now
from home.valves import ALL_VALVES

with (Path(__file__).parent / "logging.yaml").open() as fd:
    logging_cfg = yaml.load(fd.read(), yaml.Loader)

log = logging.getLogger(__name__)


ACTIONABLES = ALL_VALVES + []
CYCLE = timedelta(minutes=1)
WEB = FastAPI()


@WEB.on_event("startup")
def controller():
    async def run_in_background():
        while True:
            before = now()
            for actionable in ACTIONABLES:
                desired_state = await actionable.get_desired_state()
                if desired_state != await actionable.get_current_state():
                    await actionable.apply_state(desired_state)
            after = now()
            duration = after - before
            if duration > CYCLE:
                log.warning(f"Full cycle took {duration - CYCLE} too long")
            await asyncio.sleep((CYCLE - duration % CYCLE).total_seconds())

    asyncio.create_task(run_in_background())


# Any custom application metrics are automatically included in the exposed
# metrics. It is a good idea to attach the metrics to 'app.state' so they
# can easily be accessed in the route handler - as metrics are often
# created in a different module than where they are used.
# WEB.state.users_events_counter = Counter("events", "Number of events.")

WEB.add_middleware(MetricsMiddleware)
WEB.add_route("/metrics", metrics)


@WEB.get("/")
async def root(request: Request):
    return Response("FastAPI Middleware Example")


# @WEB.get("/users/{user_id}")
# async def get_user(
#     request: Request,
#     user_id: str,
# ):
#     request.app.state.users_events_counter.inc({"path": request.scope["path"]})
#     return Response(f"{user_id}")

uvicorn.run(WEB, port=8000, log_config=logging_cfg)
