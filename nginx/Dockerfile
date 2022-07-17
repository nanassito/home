FROM debian:latest

RUN apt-get update && \
    apt-get install -y --no-install-recommends nginx-extras lua-cjson git ca-certificates && \
    rm -rf /var/lib/apt/lists/*
RUN git clone -c transfer.fsckobjects=true https://github.com/pintsized/lua-resty-http.git /tmp/lua-resty-http && \
    cd /tmp/lua-resty-http && \
    mkdir -p /etc/nginx/lua && \
    cp -aR /tmp/lua-resty-http/lib/resty /etc/nginx/lua/resty && \
    rm -rf /tmp/lua-resty-http && \
    mkdir /etc/nginx/http.conf.d && \
    sed 's%http {%include /etc/nginx/http.conf.d/*.conf;\n\nhttp {%' -i /etc/nginx/nginx.conf

COPY ./access.lua /etc/nginx/lua/nginx-google-oauth/access.lua
COPY ./docker/etc-nginx /etc/nginx

ENTRYPOINT ["/etc/nginx/run.sh"]
