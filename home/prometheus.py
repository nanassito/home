import aiohttp
from urllib_ext.parse import urlparse

PROMETHEUS_URL = urlparse("http://192.168.1.1:9090/")


async def prom_query_one(query: str) -> float:
    async with aiohttp.ClientSession() as session:
        async with session.get(
            str(PROMETHEUS_URL / "/api/v1/query"),
            params={"query": query},
        ) as response:
            rs = await response.json()
            assert rs["status"] == "success", f"Prometheus API call failed: {rs}"
            assert len(rs["data"]["result"]) == 1
            return float(rs["data"]["result"][0]["value"][1])
