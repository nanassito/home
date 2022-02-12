from pathlib import Path

from fastapi import HTTPException, Request
from fastapi.responses import HTMLResponse

from home.facts import is_prod
from home.web import WEB, TEMPLATES


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
