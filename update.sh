set -eux

cd /github/home/
git pull

systemctl stop home
docker build -t home home/
systemctl start home