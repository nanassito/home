func install() {
    set -eux
    sudo apt install mosquitto
    set +eux
}

install()