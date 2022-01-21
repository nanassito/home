import asyncio

from home.valves import (
    VALVE_BACKYARD_DECK,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_SIDE,
)


async def stop_all():
    for valve in [
        VALVE_BACKYARD_DECK,
        VALVE_BACKYARD_SCHOOL,
        VALVE_BACKYARD_SIDE,
        VALVE_BACKYARD_HOUSE,
    ]:
        await valve.switch(False)


asyncio.run(stop_all())
