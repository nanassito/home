import asyncio
from argparse import ArgumentParser

from argparse_logging import add_logging_arguments

from home.valves import ALL_VALVES


async def run():
    for valve in ALL_VALVES:
        await valve.run()


parser = ArgumentParser()
add_logging_arguments(parser)
parser.parse_args()

asyncio.run(run())
