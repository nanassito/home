import logging
from datetime import datetime, timedelta

import aiohttp
from aioprometheus import MetricsMiddleware
from aioprometheus.asgi.starlette import metrics
from aioprometheus.collectors import Counter
from pytz import BaseTzInfo
from urllib_ext.parse import urlparse

from home.time import TimeZone, now
from home.utils import n_tries
from home.web import WEB

log = logging.getLogger(__name__)
PROMETHEUS_URL = urlparse("http://192.168.1.1:9090/")
COUNTER_NUM_RUNS = Counter("number_of_runs", "Number of times something is ran.")


@n_tries(3)
async def prom_query_one(query: str) -> float:
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(
                str(PROMETHEUS_URL / "/api/v1/query"),
                params={"query": query},
            ) as response:
                rs = await response.json()
                assert rs["status"] == "success", f"Prometheus API call failed: {rs}"
                assert len(rs["data"]["result"]) == 1
                value = float(rs["data"]["result"][0]["value"][1])
                log.debug(f"prom_query_one: {query} ...{value}")
                return value
    except Exception as err:
        log.debug(f"prom_query_one: {query} ...Failed: {err}")
        raise


@n_tries(3)
async def prom_query_labels(query: str) -> list[dict[str, str]]:
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(
                str(PROMETHEUS_URL / "/api/v1/query"),
                params={"query": query},
            ) as response:
                rs = await response.json()
                assert rs["status"] == "success", f"Prometheus API call failed: {rs}"
                value = [row["metric"] for row in rs["data"]["result"]]
                log.debug(f"prom_query_labels: {query} ...{value}")
                return value
    except Exception as err:
        log.debug(f"prom_query_one: {query} ...Failed: {err}")
        raise


@n_tries(3)
async def prom_query_series(
    query: str,
    duration: timedelta,
    tz: BaseTzInfo | TimeZone = TimeZone.UTC,
    step: timedelta = timedelta(minutes=1),
) -> list[tuple[datetime, float]]:
    tz = tz.value if isinstance(tz, TimeZone) else tz
    end = now()
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(
                str(PROMETHEUS_URL / "/api/v1/query_range"),
                params={
                    "query": query,
                    "start": (end - duration).isoformat(),
                    "end": end.isoformat(),
                    "step": int(step.total_seconds()),
                },
            ) as response:
                rs = await response.json()
                assert rs["status"] == "success", f"Prometheus API call failed: {rs}"
                assert len(rs["data"]["result"]) == 1
                values = [
                    (datetime.fromtimestamp(ts, tz), float(val))
                    for ts, val in rs["data"]["result"][0]["values"]
                ]
                log.debug(f"prom_query_series: {query} ...{len(values)}")
                return values
    except Exception as err:
        log.debug(f"prom_query_series: {query} ...Failed: {err}")
        raise


def init() -> None:
    WEB.add_middleware(MetricsMiddleware)
    WEB.add_route("/metrics", metrics)
