set -eux

sudo docker run \
    -it --rm \
    --name certbot \
    -p 80:80 \
    --mount source=letsencrypt_etc,target=/etc/letsencrypt \
    --mount source=letsencrypt_varlib,target=/var/lib/letencrypt \
    certbot/certbot \
    certonly \
    --manual \
    -d "*.epa.jaminais.fr,*.eastpaloalto.jaminais.fr,epa.jaminais.fr,eastpaloalto.jaminais.fr" \
    --preferred-challenges=dns
