from datetime import datetime
from enum import Enum

from pytz import BaseTzInfo, timezone


class TimeZone(Enum):
    PT = timezone("US/Pacific")
    UTC = timezone("UTC")


def today_at(hours: int, minutes: int, tz: BaseTzInfo | TimeZone) -> datetime:
    tz = tz.value if isinstance(tz, TimeZone) else tz
    return datetime(
        *datetime.now().timetuple()[:3], hour=hours, minute=minutes, tzinfo=tz
    )


def now(
    tz: BaseTzInfo | TimeZone = TimeZone.UTC,
) -> datetime:
    tz = tz.value if isinstance(tz, TimeZone) else tz
    return datetime.now(tz)
