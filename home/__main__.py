import asyncio
from argparse import ArgumentParser

from argparse_logging import add_logging_arguments

import home


parser = ArgumentParser()
add_logging_arguments(parser)
parser.parse_args()

asyncio.run(home.run())
