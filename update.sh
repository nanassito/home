set -eux


cd /github/home/

git_version(){
    git log -q $1 | head -n1 | awk '{print $2}'
}

docker_pid(){
    docker ps --format="{{.ID}} {{.Image}}" | grep "$1" | awk '{print $1}'
}

docker_update(){
    OLD_PROC=$(docker_pid "$1")
    docker pull $(docker ps --format="{{.Image}}" | grep "$1")
    docker stop ${OLD_PROC}
}


ALERTMANAGER_GIT_PREV=$(git_version alertmanager)
HOME_GIT_PREV=$(git_version home)
NGINX_GIT_PREV=$(git_version nginx)
PROMETHEUS_GIT_PREV=$(git_version prometheus)
ZIGBEE2MQTT_GIT_PREV=$(git_version zigbee2mqtt)
SWITCHES_GIT_PREV=$(git_version bin/switches)
NETSCAN_GIT_PREV=$(git_version bin/netscan)


git pull -s recursive -X theirs


if [ "${ALERTMANAGER_GIT_PREV}" != "$(git_version alertmanager)" ]; then
    systemctl daemon-reload
    docker_update "prom/alertmanager"
fi

if [ "${HOME_GIT_PREV}" != "$(git_version home)" ]; then
    OLD_PROC=$(docker_pid "home:latest")
    docker build -t home home/
    docker stop ${OLD_PROC}  # systemd will restart with the new image.
fi

if [ "${NGINX_GIT_PREV}" != "$(git_version nginx)" ]; then
    systemctl daemon-reload
    docker build -t web nginx/
    systemctl restart nginx
fi

if [ "${PROMETHEUS_GIT_PREV}" != "$(git_version prometheus)" ]; then
    systemctl daemon-reload
    docker_update "prom/prometheus"
fi

if [ "${ZIGBEE2MQTT_GIT_PREV}" != "$(git_version zigbee2mqtt)" ]; then
    systemctl daemon-reload
    docker_update "koenkk/zigbee2mqtt"
fi


if [ "${SWITCHES_GIT_PREV}" != "$(git_version bin/switches)" ]; then
    systemctl daemon-reload
    systemctl restart switches
fi


if [ "${NETSCAN_GIT_PREV}" != "$(git_version bin/netscan)" ]; then
    systemctl daemon-reload
    systemctl restart netscan
fi