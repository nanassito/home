set -eux

git pull
poetry install --no-dev
poetry run python -m home