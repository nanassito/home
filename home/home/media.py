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


VIDEOS = [
    Video("Un éléphant qui se balançait", "#4c85ba"),
    Video("Mon petit lapin a du chagrin", "#9ab360"),
    Video("Petit escargot", "#b3b92b"),
    Video("L'araignée Gipsy", "#0091bf"),
    Video("Il pleut, il mouille", "#e87256"),
    Video("Mousse mousse dans le bain", "#b27cad"),
    Video("Doucement s'en va le jour", "#999184"),
    Video("Y avait des gros crocodiles", "#19713e"),
    Video("La citrouille", "#bf6a7f"),
    Video("Une fourmi m'a piqué la main", "#817dc9"),
    Video("Le clown", "#3494e1"),
    Video("J'fais pipi sur l'gazon", "#eba134"),
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
