@echo off
title 面试模拟系统
color 0A

echo ================================
echo    面试模拟系统 启动中...
echo ================================

REM 启动后端服务
echo [1/3] 启动后端 API 服务...
cd /d e:\interview\api
start "API服务" cmd /k "go run main.go"

REM 启动前端服务
echo [2/3] 启动前端 Vue 服务...
cd /d e:\interview\ui
start "前端服务" cmd /k "npm run serve"

REM 等待前端启动
echo 等待服务启动...
timeout /t 15 /nobreak >nul

REM 启动 cpolar 内网穿透
echo [3/3] 启动内网穿透服务...
start "内网穿透" cmd /k "e:\cpolar-portable\cpolar.exe http 3000"

echo.
echo ================================
echo    所有服务已启动！
echo    前端: http://localhost:3000
echo ================================
pause
