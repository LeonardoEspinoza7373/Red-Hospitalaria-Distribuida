#!/bin/bash
set -euo pipefail

NODO_SRC="dist/nodo"
PROXY_SRC="dist/proxy"

NODES=("192.168.1.10" "192.168.1.11" "192.168.1.12" "192.168.1.13")
PROXY_IP="192.168.1.14"
REMOTE_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
NODO_SERVICE="deploy/nodo.service"
PROXY_SERVICE="deploy/proxy.service"

if [ ! -f "$NODO_SRC" ] || [ ! -f "$PROXY_SRC" ]; then
    echo "Error: primero ejecuta ./build.sh"
    exit 1
fi

read -rp "Usuario SSH en las laptops: " SSH_USER

copy_and_install() {
    local ip="$1"
    local src="$2"
    local name="$3"
    local service="$4"

    echo "→ $ip: copiando $name ..."
    scp "$src" "${SSH_USER}@${ip}:~/${name}"
    ssh "${SSH_USER}@${ip}" \
        "sudo mv ~/${name} ${REMOTE_DIR}/${name} && sudo chmod +x ${REMOTE_DIR}/${name}"

    if [ -n "$service" ] && [ -f "$service" ]; then
        echo "→ $ip: instalando servicio systemd ..."
        scp "$service" "${SSH_USER}@${ip}:~/${name}.service"
        ssh "${SSH_USER}@${ip}" \
            "sudo mv ~/${name}.service ${SERVICE_DIR}/${name}.service && sudo systemctl daemon-reload"
    fi
}

for ip in "${NODES[@]}"; do
    copy_and_install "$ip" "$NODO_SRC" "nodo" "nodo.service"
done

echo "→ $PROXY_IP: copiando proxy ..."
copy_and_install "$PROXY_IP" "$PROXY_SRC" "proxy" "proxy.service"

echo ""
echo "==> Deploy completado."
echo ""
echo "Para iniciar los servicios manualmente:"
echo "  ssh ${SSH_USER}@<ip> 'sudo systemctl enable --now nodo.service'"
echo "  ssh ${SSH_USER}@${PROXY_IP} 'sudo systemctl enable --now proxy.service'"
echo ""
echo "Para ver logs:"
echo "  ssh ${SSH_USER}@<ip> 'sudo journalctl -fu nodo.service'"
