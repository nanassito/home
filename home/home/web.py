from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles


WEB = FastAPI()
WEB.mount(
    "/static",
    StaticFiles(directory=str(Path("__file__").parent / "web"), html=True),
    name="static",
)
