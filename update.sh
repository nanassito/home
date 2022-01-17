set -eux


cd /github/home/

git_version(){
    git log -q $1 | head -n1 | awk '{print $2}'
}


ALERTMANAGER_GIT_PREV=$(git_version alertmanager)
GRAFANA_GIT_PREV=$(git_version grafana)
HOME_GIT_PREV=$(git_version home)
NGINX_GIT_PREV=$(git_version nginx)
PROMETHEUS_GIT_PREV=$(git_version prometheus)
ZIGBEE2MQTT_GIT_PREV=$(git_version zigbee2mqtt)


git pull -s recursive -X theirs


if [ "${ALERTMANAGER_GIT_PREV}" != "$(git_version alertmanager)" ]; then
    systemctl daemon-reload
    systemctl restart alertmanager
fi

if [ "${GRAFANA_GIT_PREV}" != "$(git_version grafana)" ]; then
    systemctl daemon-reload
    systemctl restart grafana
fi

if [ "${HOME_GIT_PREV}" != "$(git_version home)" ]; then
    OLD_PROC=$(docker ps | grep "home:latest" | awk '{print $1}')
    docker build -t home home/
    docker stop ${OLD_PROC}  # systemd will restart with the new image.
fi

if [ "${NGINX_GIT_PREV}" != "$(git_version nginx)" ]; then
    systemctl daemon-reload
    systemctl restart nginx
fi

if [ "${PROMETHEUS_GIT_PREV}" != "$(git_version prometheus)" ]; then
    systemctl daemon-reload
    systemctl restart prometheus
fi

if [ "${ZIGBEE2MQTT_GIT_PREV}" != "$(git_version zigbee2mqtt)" ]; then
    systemctl daemon-reload
    systemctl restart zigbee2mqtt
fi
