FROM python:3.10

WORKDIR /app

ADD poetry.lock pyproject.toml /app/
RUN pip install poetry && poetry install --no-dev --no-root

ADD home /app/home
ADD templates /app/templates
ADD static /app/static
RUN poetry install --no-dev

CMD poetry run python -m home