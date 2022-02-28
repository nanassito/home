from fastapi import Request
from fastapi.responses import HTMLResponse

from home.prometheus import prom_query_one
from home.web import TEMPLATES, WEB


def init():
    @WEB.get("/temperature", response_class=HTMLResponse)
    async def get_soaker(request: Request):
        return TEMPLATES.TemplateResponse(
            "temperature.html.jinja",
            {
                "request": request,
                "page": "Temperature",
                "rooms": [
                    {
                        "name": room,
                        "current": round(
                            await prom_query_one(
                                f'mqtt_temperature{{topic="{topic}"}}'
                            ),
                            1,
                        ),
                        "min_1d": round(
                            await prom_query_one(
                                f'min_over_time(mqtt_temperature{{topic="{topic}"}}[1d])'
                            ),
                            1,
                        ),
                        "max_1d": round(
                            await prom_query_one(
                                f'max_over_time(mqtt_temperature{{topic="{topic}"}}[1d])'
                            ),
                            1,
                        ),
                        "link": f'https://prometheus.epa.jaminais.fr/graph?g0.expr=mqtt_temperature{{topic%3D"{topic}"}}&g0.tab=0&g0.range_input=1d',
                    }
                    for room, topic in {
                        "Zaya": "zigbee2mqtt_air_zaya",
                        "Parent": "zigbee2mqtt_air_parent",
                        "Salon": "zigbee2mqtt_air_livingroom",
                        "Office": "zigbee2mqtt_air_office",
                        "Outside": "zigbee2mqtt_air_outside",
                    }.items()
                ],
            },
        )
