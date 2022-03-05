import time

import requests
from prometheus_client import start_http_server
from prometheus_client.core import REGISTRY, CounterMetricFamily

from rain.config import SYNOPTIC_TOKEN


class RainCollector(object):
    def __init__(self):
        self.start_time = time.time()
        self.stations = {
            "F4582",
            "PAAC1",
            "F4751",
            "E0597",
            "C7418",
            "RWCC1",
            "AR321",
            "AV510",
            "E7138",
            "COOPRWCC1",
        }
        self.last_values = {s: 0 for s in self.stations}

    def get_minutes_since_start(self) -> int:
        return int((time.time() - self.start_time) / 60)
    
    def get_counter_family(self):
        return CounterMetricFamily(
            "rainfall", "Amount of rain fall in millimeters", labels=["station"]
        )

    def describe(self):
        c = self.get_counter_family()
        for station in self.stations:
            c.add_metric([station], 0)
        yield c

    def collect(self):
        c = self.get_counter_family()
        recent = self.get_minutes_since_start()
        print(f"Fech {recent} minutes of rainfall data")
        data = requests.get(
            "https://api.synopticdata.com/v2/stations/precip",
            params={
                "token": SYNOPTIC_TOKEN,
                "recent": recent,
                "pmode": "totals",
                "units": "metric",
                "stid": ",".join(self.stations),
            },
        ).json()
        match data["SUMMARY"]["RESPONSE_CODE"]:
            case 1:  # success
                for station in data["STATION"]:
                    for obs in station["OBSERVATIONS"]["precipitation"]:
                        self.last_values[station["NAME"]] = obs["total"]
                        c.add_metric([station["NAME"]], obs["total"])
            case 2:  # No data in the requested range
                for station in self.stations:
                    c.add_metric([station], 0)
            case _:
                raise RuntimeError(f"Can't process response form Synoptic: {data}")

        yield c


if __name__ == "__main__":
    start_http_server(8001)
    REGISTRY.register(RainCollector())
    while True:
        time.sleep(1)
