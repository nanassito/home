from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

WEB = FastAPI()


TEMPLATES = Jinja2Templates(directory=str(Path("__file__").parent / "templates"))
