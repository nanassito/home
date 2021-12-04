from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles


API = FastAPI()
WEB = FastAPI()
WEB.mount("/api", API, name="api")
WEB.mount(
    "/",
    StaticFiles(directory=str(Path("__file__").parent / "web"), html=True),
    name="static",
)
