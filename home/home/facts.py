import logging
from asyncio import Lock
from datetime import datetime, timedelta

import aiohttp
from aioprometheus import Gauge

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


@n_tries(3)
async def is_day_time() -> bool:
    async def _get_sun_times() -> tuple[datetime, datetime]:
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(
                    "https://api.sunrise-sunset.org/json",
                    params={
                        "lat": 37.4845223,
                        "lng": -122.1405406,
                        "formatted": 0,
                    },
                ) as response:
                    rs = await response.json()
                    assert rs["status"] == "OK", f"Sunrise/Sunset API failed: {rs}"
                    sunrise = datetime.fromisoformat(rs["results"]["sunrise"])
                    sunset = datetime.fromisoformat(rs["results"]["sunset"])
                    return sunrise, sunset
        except Exception as err:
            log.debug(f"is_night_time Failed: {err}")
            raise

    sunrise, sunset = await _get_sun_times()
    for offset in (
        timedelta(days=-1),
        timedelta(days=0),
        timedelta(days=1),
    ):
        if sunrise < now() + offset < sunset:
            _PROM_IS_DAY_TIME.set({"city": "east_palo_alto"}, 1)
            return True
    _PROM_IS_DAY_TIME.set({"city": "east_palo_alto"}, 0)
    return False


async def is_night_time() -> bool:
    return not await is_day_time()
