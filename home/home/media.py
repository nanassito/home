from dataclasses import dataclass
from pathlib import Path
from turtle import color

from fastapi import HTTPException, Request
from fastapi.responses import HTMLResponse

from home.facts import is_prod
from home.web import TEMPLATES, WEB

if is_prod():
    MUSIC = {
        album.name: [
            song.name
            for song in sorted(album.iterdir())
            if song.is_file()
            if song.suffix == ".ogg"
        ]
        for album in sorted(Path("/app/static/music").iterdir())
    }
else:
    MUSIC = {
        "Various-20_chansons_et_berçeuse_du_monde": [
            "09.Italie-Cade_luliva.ogg",
            "10.Grèce-I_trata_mas_i_Kourelou.ogg",
        ]
    }


@dataclass
class Video:
    title: str
    color: str
    id: str
    start: int = 4


VIDEOS = [
    Video("Un éléphant qui se balançait", "#4c85ba", "SQHmmlliuEc"),
    Video("Mon petit lapin a du chagrin", "#9ab360", "jYKPpizdqQo"),
    Video("Petit escargot", "#b3b92b", "7R7Aoc7sC0A"),
    Video("L'araignée Gipsy", "#0091bf", "H5BcoZgpaI4"),
    Video("Il pleut, il mouille", "#e87256", "XRCQnGaNHy0"),
    Video("Mousse mousse dans le bain", "#b27cad", "deTyMcthKg0"),
    Video("Doucement s'en va le jour", "#999184", "W5uLg-Fd4jQ"),
    Video("Y avait des gros crocodiles", "#19713e", "5oLDMJeEQ1w"),
    Video("La citrouille", "#bf6a7f", "CTAIt3ymli8"),
    Video("Une fourmi m'a piqué la main", "#817dc9", "VvMk0zGq1ew"),
    Video("Le Clown", "#3494e1", "poMHPAM_Sbk"),
]


def init() -> None:
    @WEB.get("/music", response_class=HTMLResponse)
    async def get_music(request: Request, album: str = "", song: str = ""):
        if album or song:
            if song not in MUSIC.get(album, {}):
                raise HTTPException(404)
        return TEMPLATES.TemplateResponse(
            "music.html.jinja",
            {
                "request": request,
                "page": "Music",
                "albums": MUSIC,
                "listen": {
                    "album": album,
                    "song": song,
                },
            },
        )

    @WEB.get("/video", response_class=HTMLResponse)
    async def get_video(request: Request):
        return TEMPLATES.TemplateResponse(
            "video.html.jinja",
            {
                "request": request,
                "page": "Video",
                "videos": VIDEOS,
            },
        )
