#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="/opt/rainbow-backend"
BIN_DIR="${APP_DIR}/bin"
UPLOAD_DIR="${APP_DIR}/uploads/prod"
LOG_DIR="${APP_DIR}/logs/prod"
H5_DIR="${APP_DIR}/www/h5"
ADMIN_DIR="${APP_DIR}/www/admin"
LOCAL_BIN="${ROOT_DIR}/bin/rainbow-backend"
SERVICE_NAME="rainbow-backend-prod.service"
SERVICE_TARGET="/etc/systemd/system/${SERVICE_NAME}"
SERVICE_SOURCE="${ROOT_DIR}/deploy/systemd/${SERVICE_NAME}"
ENV_SOURCE="${ROOT_DIR}/deploy/env/prod.env.example"
ENV_TARGET="${APP_DIR}/prod.env"
NGINX_SOURCE="${ROOT_DIR}/deploy/nginx/rainbow-backend-ports.conf"
NGINX_TARGET="/etc/nginx/conf.d/rainbow-backend-ports.conf"
LOCAL_URL="http://127.0.0.1:28081/health"
PUBLIC_URL="http://<public-ip>:28080/health"

cd "${ROOT_DIR}"

echo "Building rainbow-backend binary..."
mkdir -p "${ROOT_DIR}/bin"
go build -o "${LOCAL_BIN}" ./cmd/server

echo "Installing binary to ${BIN_DIR}..."
sudo mkdir -p "${BIN_DIR}"
sudo install -m 0755 "${LOCAL_BIN}" "${BIN_DIR}/rainbow-backend"

echo "Ensuring upload directories exist under ${UPLOAD_DIR}..."
sudo mkdir -p \
  "${UPLOAD_DIR}/images" \
  "${UPLOAD_DIR}/audio" \
  "${UPLOAD_DIR}/default/images" \
  "${UPLOAD_DIR}/default/audio"

echo "Ensuring log directory exists at ${LOG_DIR}..."
sudo mkdir -p "${LOG_DIR}"

echo "Ensuring frontend static roots exist..."
sudo mkdir -p "${H5_DIR}" "${ADMIN_DIR}"

if sudo test -f "${ENV_TARGET}"; then
  echo "Keeping existing env file: ${ENV_TARGET}"
else
  echo "Creating env file from example: ${ENV_TARGET}"
  sudo install -m 0640 "${ENV_SOURCE}" "${ENV_TARGET}"
fi

echo "Installing systemd unit: ${SERVICE_TARGET}"
sudo install -m 0644 "${SERVICE_SOURCE}" "${SERVICE_TARGET}"

echo "Installing Nginx config: ${NGINX_TARGET}"
# sudo install -m 0644 "${NGINX_SOURCE}" "${NGINX_TARGET}"

echo "Reloading systemd and restarting ${SERVICE_NAME}..."
sudo systemctl daemon-reload
sudo systemctl enable "${SERVICE_NAME}"
sudo systemctl restart "${SERVICE_NAME}"

echo "Validating and reloading Nginx..."
sudo nginx -t
sudo systemctl reload nginx

cat <<EOF
WARNING: edit ${ENV_TARGET} and replace the example JWT secret, admin password, database DSN, and ALLOW_ORIGINS before production use.

Verification:
  sudo systemctl status ${SERVICE_NAME} --no-pager
  sudo journalctl -u ${SERVICE_NAME} -n 100 --no-pager
  sudo tail -n 100 ${LOG_DIR}/app.log
  sudo tail -n 100 ${LOG_DIR}/access.log
  curl ${LOCAL_URL}
  curl ${PUBLIC_URL}

URLs:
  Local health check: ${LOCAL_URL}
  Public base URL: http://<public-ip>:28080
  Public health check: ${PUBLIC_URL}
  Static image URL pattern: http://<public-ip>:28080/static/images/<filename>
  Static audio URL pattern: http://<public-ip>:28080/static/audio/<filename>
  Scene image URL pattern: http://<public-host>/static/<scene_code>/images/<filename>
  Scene audio URL pattern: http://<public-host>/static/<scene_code>/audio/<filename>
EOF
