from datetime import timedelta

from home.valves import Valve


async def run():
    valve_schedules = {
        "backyard_side": timedelta(minutes=10),
        "backyard_house": timedelta(minutes=15),
        "backyard_school": timedelta(minutes=15),
        "backyard_deck": timedelta(minutes=15),
        # "frontyard": timedelta(minutes=0),
    }
    for name, duration in valve_schedules.items():
        valve = Valve(name)
        await valve.run_for(duration)
