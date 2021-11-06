import asyncio
import logging
from argparse import ArgumentParser
from datetime import timedelta

from argparse_logging import add_logging_arguments

from home.time import now
from home.valves import ALL_VALVES

log = logging.getLogger(__name__)


ACTIONABLES = ALL_VALVES + []
CYCLE = timedelta(minutes=1)


async def run():
    while True:
        before = now()
        for actionable in ACTIONABLES:
            state = await actionable.get_current_state()
            if state != await actionable.get_desired_state():
                await actionable.apply_state(state)
        after = now()
        duration = after - before
        if duration > CYCLE:
            log.warn(f"Full cycle took {duration - CYCLE} too long")
        await asyncio.sleep((CYCLE - duration % CYCLE).total_seconds())


parser = ArgumentParser()
add_logging_arguments(parser)
parser.parse_args()

asyncio.run(run())
