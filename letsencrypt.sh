set -eux

mkdir -p ${ROOT}/letsencrypt
sudo docker run \
    -it --rm \
    --name certbot \
    -p 80:80 \
    --mount type=bind,source=${ROOT}/letsencrypt/etc,target=/etc/letsencrypt \
    --mount type=bind,source=${ROOT}/letsencrypt/varlib,target=/var/lib/letencrypt \
    certbot/certbot \
    certonly \
    --manual \
    -d "*.epa.jaminais.fr,*.eastpaloalto.jaminais.fr,epa.jaminais.fr,eastpaloalto.jaminais.fr" \
    --preferred-challenges=dns