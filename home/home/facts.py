import logging

from aioprometheus import Gauge

from home.prometheus import prom_query_one

log = logging.getLogger(__name__)


_PROM_IS_DAY_TIME = Gauge("is_day_time", "1 if day, 0 if night.")


async def is_day_time() -> bool:
    promql = "max(max_over_time(mqtt_illuminance_lux[1h]))"
    lux = await prom_query_one(promql)
    is_day = lux > 100
    _PROM_IS_DAY_TIME.set({"city": "east_palo_alto"}, int(is_day))
    return is_day


async def is_night_time() -> bool:
    return not await is_day_time()