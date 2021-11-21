from aioprometheus import Gauge
from asyncio import Lock


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

