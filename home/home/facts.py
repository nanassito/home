import logging
from pathlib import Path

from aioprometheus import Gauge

from home.prometheus import prom_query_one

log = logging.getLogger(__name__)


_PROM_IS_DAY_TIME = Gauge("is_day_time", "1 if day, 0 if night.")


async def is_day_time() -> bool:
    # Temporarily excluding the site motion sensor due to reliability issues.
    promql = (
        'quantile(0.5, max_over_time(mqtt_illuminance_lux{topic!="zigbee2mqtt_motion_side"}[1h]))'
    )
    lux = await prom_query_one(promql)
    is_day = lux > 100
    _PROM_IS_DAY_TIME.set({"city": "east_palo_alto"}, int(is_day))
    return is_day


async def is_night_time() -> bool:
    return not await is_day_time()


def is_prod() -> bool:
    if Path("/.dockerenv").exists():
        return True
    cgroup = Path("/proc/self/cgroup")
    if cgroup.exists():
        with cgroup.open() as fd:
            while line := fd.readline():
                if "docker" in line:
                    return True
    return False


async def get_outside_temp() -> float:
    promql = 'max(mqtt_temperature{topic="zigbee2mqtt_air_outside"})'
    return await prom_query_one(promql)