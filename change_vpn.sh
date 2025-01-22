#!/bin/bash

PAR_URL="https://objectstorage.sa-vinhedo-1.oraclecloud.com/p/wTVEtH6QCXGWIMKWDWc0SLJ1DKnxn_fxPgrCp1piATStL51LbvC2nt8YY_pWcmpA/n/axn5fmjnwwfz/b/biofabricas-vpn-files/o/"

SCRIPT_DIR="$(dirname "$(realpath "$0")")"

DOWNLOAD_DIR="$SCRIPT_DIR"

mkdir -p "$DOWNLOAD_DIR"

echo "Fetching list of objects from the bucket..."
RESPONSE=$(curl -s "${PAR_URL}")

OBJECTS=$(echo "$RESPONSE" | grep -oP '"name":"\K[^"]+')

if [[ -z "$OBJECTS" ]]; then
    echo "No objects found in the bucket or failed to fetch the list. Please check the PAR URL."
    exit 1
fi

echo "Downloading objects..."
for OBJECT in $OBJECTS; do
    echo "Downloading: $OBJECT"
    OBJECT_DIR=$(dirname "$OBJECT")
    mkdir -p "$DOWNLOAD_DIR/$OBJECT_DIR"
    curl -s -o "$DOWNLOAD_DIR/$OBJECT" "${PAR_URL}${OBJECT}" || echo "Failed to download $OBJECT"
done

echo "All files have been downloaded to $DOWNLOAD_DIR"

FILES=("ca.crt" "bfs.key" "bfs.crt" "ta.key" "client.conf")
DESTINATION="/etc/openvpn"

for file in "${FILES[@]}"; do
  if [ -f "$file" ]; then
    echo "Moving $file to $DESTINATION..."
    sudo mv "$file" "$DESTINATION"
  else
    echo "Warning: $file does not exist and will be skipped."
  fi
done

echo "Restarting OpenVPN client service..."
sudo systemctl restart openvpn@client.service

STATUS=$?
if [ $STATUS -eq 0 ]; then
  echo "OpenVPN client service restarted successfully."
else
  echo "Failed to restart OpenVPN client service. Please check the logs for more details."
fi

grep -q "134.65.246.124 network.hubioagro.com.br" /etc/hosts || echo "134.65.246.124 network.hubioagro.com.br" | sudo tee -a /etc/hosts
