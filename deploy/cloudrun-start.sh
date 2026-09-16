#!/bin/bash
# 云托管 all-in-one 启动脚本：Redis -> MySQL -> Go 后端（后端保持前台运行）
# 关键设计：
#   - MySQL 数据目录由我们自己 initialize-insecure 初始化（root 空密码 + caching_sha2），
#     不依赖 Ubuntu 发行版的 auth_socket/root 随机密码，容器内 100% 可复现
#   - 所有管理操作走 TCP 127.0.0.1 验证真实认证（mysqladmin ping 密码错了也返回 alive，不可靠）
#   - 后端启动时自动建库 + 执行 db/migrations/*.sql，无需手动导表
set -e

APP_USER="${MYSQL_USER:-interview}"
APP_PASS="${MYSQL_PASSWORD:-interview_dev_pass}"
APP_DB="${MYSQL_DB:-interview_sim}"
REDIS_PORT="${REDIS_PORT:-6379}"
DATA_DIR=/data/mysql
REDIS_DATA_DIR=/data/redis

echo "==> [1/3] 启动 Redis (port ${REDIS_PORT})"
# 显式指定数据目录：不写 --dir 时 Redis 把 AOF/RDB 落在进程工作目录（/app），位置飘忽难以备份
mkdir -p "${REDIS_DATA_DIR}"
chmod 755 /data "${REDIS_DATA_DIR}" 2>/dev/null || true
redis-server --bind 127.0.0.1 --port "${REDIS_PORT}" --daemonize yes --appendonly yes --dir "${REDIS_DATA_DIR}" \
  ${REDIS_PASSWORD:+--requirepass "${REDIS_PASSWORD}"}

echo "==> [2/3] 初始化并启动 MySQL"
mkdir -p /var/run/mysqld "${DATA_DIR}"
chmod 755 /data 2>/dev/null || true
chown mysql:mysql /var/run/mysqld
if [ ! -d "${DATA_DIR}/mysql" ]; then
  echo "    首次启动，初始化数据目录（root 空密码）..."
  mysqld --initialize-insecure --user=mysql --datadir="${DATA_DIR}"
  chown -R mysql:mysql "${DATA_DIR}"
fi

# 直接启动 mysqld（不用 service/init.d，容器内更稳），日志走 stdout 便于云平台排查
mysqld --user=mysql --datadir="${DATA_DIR}" \
       --bind-address=127.0.0.1 --port=3306 \
       --mysqlx=OFF &

# 等待 MySQL 真实可登录（最多 120 秒）
READY=""
for i in $(seq 1 120); do
  if mysql -h 127.0.0.1 -P 3306 -uroot --password= -e "SELECT 1" >/dev/null 2>&1; then
    READY="yes"
    echo "    MySQL 已就绪"
    break
  fi
  sleep 1
done
if [ -z "${READY}" ]; then
  echo "    MySQL 启动超时，退出"
  exit 1
fi

# 创建后端专用账号与业务库（initialize-insecure 的 root 走 TCP 空密码认证）
mysql -h 127.0.0.1 -P 3306 -uroot --password= <<SQL
CREATE USER IF NOT EXISTS '${APP_USER}'@'%' IDENTIFIED BY '${APP_PASS}';
ALTER USER '${APP_USER}'@'%' IDENTIFIED BY '${APP_PASS}';
CREATE DATABASE IF NOT EXISTS \`${APP_DB}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON *.* TO '${APP_USER}'@'%';
FLUSH PRIVILEGES;
SQL
echo "    MySQL 账号就绪: ${APP_USER}"

echo "==> [3/3] 启动后端服务 (port ${SERVER_PORT:-8080})"
exec /app/api/server
