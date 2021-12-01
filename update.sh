set -eux


cd /github/home/
git pull

OLD_PROC=$(docker ps | grep "home:latest" | awk '{print $1}')
docker build -t home home/
docker stop ${OLD_PROC}  # systemd will restart with the new image.