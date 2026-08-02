#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPO="t3rcio/tk-clipboard-go-service"
BIN_DIR="$HOME/.local/bin"
SYSTEMD_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SYSTEMD_DIR/clipsync.service"

echo -e "${BLUE}=================================================="${NC}
echo -e "${BLUE}        ClipSync Daemon - Instalação Rápida        "${NC}
echo -e "${BLUE}=================================================="${NC}

DEFAULT_SERVER="https://wordpress.vps-kinghost.net"
SERVER_URL="${CLIPSYNC_SERVER:-$DEFAULT_SERVER}"

if [ -z "$CLIPSYNC_SERVER" ]; then
    read -p "Digite a URL do servidor ClipSync [padrão: $DEFAULT_SERVER]: " INPUT_SERVER
    if [ -n "$INPUT_SERVER" ]; then
        SERVER_URL="$INPUT_SERVER"
    fi
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ASSET_NAME="clipsync-daemon-linux-amd64"
        ;;
    aarch64|arm64)
        ASSET_NAME="clipsync-daemon-linux-arm64"
        ;;
    *)
        echo -e "${RED}[ERRO] Arquitetura não suportada: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "\n${YELLOW}-> Arquitetura detectada:${NC} $ARCH ($ASSET_NAME)"
echo -e "${YELLOW}-> URL do servidor:${NC} $SERVER_URL"

DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$ASSET_NAME"

mkdir -p "$BIN_DIR"
mkdir -p "$SYSTEMD_DIR"

echo -e "\n${YELLOW}-> Baixando o executável compilado do GitHub...${NC}"
if command -v curl &> /dev/null; then
    curl -sSL "$DOWNLOAD_URL" -o "$BIN_DIR/clipsync-daemon"
elif command -v wget &> /dev/null; then
    wget -qO "$BIN_DIR/clipsync-daemon" "$DOWNLOAD_URL"
else
    echo -e "${RED}[ERRO] cURL ou Wget não encontrados. Instale um para continuar.${NC}"
    exit 1
fi

chmod +x "$BIN_DIR/clipsync-daemon"
echo -e "${GREEN}✓ Executável instalado em:${NC} $BIN_DIR/clipsync-daemon"

echo -e "\n${YELLOW}-> Criando serviço Systemd do usuário...${NC}"

cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=ClipSync Go Daemon
After=network.target

[Service]
ExecStart=$BIN_DIR/clipsync-daemon -server $SERVER_URL
Restart=always
RestartSec=5s

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable clipsync.service
systemctl --user restart clipsync.service

echo -e "\n${GREEN}=================================================="${NC}
echo -e "${GREEN} ClipSync instalado e em execução! 🚀"${NC}
echo -e "${GREEN}=================================================="${NC}
echo -e "\nPara visualizar o PIN de pareamento:"
echo -e "  ${YELLOW}journalctl --user -u clipsync -f${NC}\n"