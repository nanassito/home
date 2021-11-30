import logging
from asyncio import Lock
from datetime import datetime, timedelta

import aiohttp
from aioprometheus import Gauge
from home.prometheus import prom_query_one

from home.time import now
from home.utils import n_tries

log = logging.getLogger(__name__)


_PROM_WATER_USERS_QTY = Gauge(
    "water_user_qty", "Number of active users of the water main line."
)
_WATER_USAGE_LOCK: Lock = Lock()
_WATER_USERS: set[str] = set()


async def start_using_water(user: str) -> None:
    async with _WATER_USAGE_LOCK:
        _WATER_USERS.add(user)
    _PROM_WATER_USERS_QTY.set({}, len(_WATER_USERS))


async def stop_using_water(user: str) -> None:
    async with _WATER_USAGE_LOCK:
        if user in _WATER_USERS:
            _WATER_USERS.remove(user)
    _PROM_WATER_USERS_QTY.set({}, len(_WATER_USERS))


async def has_other_water_users(user: str) -> bool:
    return _WATER_USERS != {user} and _WATER_USERS != set()


_PROM_IS_DAY_TIME = Gauge("is_day_time", "1 if day, 0 if night.")


async def is_day_time() -> bool:
    promql = "max(max_over_time(mqtt_illuminance_lux[1h]))"
    lux = await prom_query_one(promql)
    is_day = lux > 100
    _PROM_IS_DAY_TIME.set({"city": "east_palo_alto"}, is_day)
    return bool(is_day)


_PROM_MOWER_STATUS_CODE = Gauge("mower_status_code", "home=1, mowing=7, others?")


@n_tries(3)
async def is_mower_running() -> bool:
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get("http://172.17.0.1/landroid-s/status") as response:
                rs = await response.json()
                _PROM_MOWER_STATUS_CODE.set({"city": "east_palo_alto"}, rs["statusCode"])
                return rs["statusCode"] == 7
    except Exception as err:
        log.debug(f"is_mower_running Failed: {err}")
        raise
