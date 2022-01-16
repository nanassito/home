set -eux


cd /github/home/

git_version(){
    git log -q $1 | head -n1 | awk '{print $2}'
}


HOME_GIT_PREV=$(git_version home)
PROM_GIT_PREV=$(git_version prometheus)
NGINX_GIT_PREV=$(git_version nginx)


git pull -s recursive -X theirs


if [ "${HOME_GIT_PREV}" != "$(git_version home)" ]; then
    OLD_PROC=$(docker ps | grep "home:latest" | awk '{print $1}')
    docker build -t home home/
    docker stop ${OLD_PROC}  # systemd will restart with the new image.
fi

if [ "${PROM_GIT_PREV}" != "$(git_version prometheus)" ]; then
    systemctl daemon-reload
    systemctl restart prometheus
fi

if [ "${NGINX_GIT_PREV}" != "$(git_version nginx)" ]; then
    systemctl daemon-reload
    systemctl restart nginx
fi
