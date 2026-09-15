# 一键打包微信云托管代码包（zip 根目录 = Docker 构建上下文）
# 用法：右键本文件 → 使用 PowerShell 运行，或直接执行：
#   powershell -ExecutionPolicy Bypass -File deploy\pack-cloudrun.ps1
# 产物：deploy\cloudrun-pack.zip → 上传到云托管控制台创建版本

$ErrorActionPreference = "Stop"

$root  = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$stage = Join-Path $root "deploy\.cloudrun-pack"
$zip   = Join-Path $root "deploy\cloudrun-pack.zip"

if (Test-Path $stage) { Remove-Item -Recurse -Force $stage }
New-Item -ItemType Directory -Path $stage | Out-Null

# 1. Dockerfile 必须位于代码包根目录（云托管约定）
Copy-Item (Join-Path $root "deploy\Dockerfile.cloudrun") (Join-Path $stage "Dockerfile")

# 2. 构建期 .dockerignore：避免把本地密钥/构建产物送进云端构建
@"
api/.env
api/*.exe
api/*.exe~
api/*.log
api/.idea
api/scripts
"@ | Set-Content -Path (Join-Path $stage ".dockerignore") -Encoding ASCII

# 3. 后端源码（排除本地密钥与构建产物，.dockerignore 会再兜底一次）
robocopy (Join-Path $root "api") (Join-Path $stage "api") /E /XF .env .env.* *.exe *.exe~ *.log /XD .idea scripts | Out-Null
if ($LASTEXITCODE -ge 8) { throw "复制 api 源码失败 (robocopy exit $LASTEXITCODE)" }

# 4. 启动脚本
New-Item -ItemType Directory -Path (Join-Path $stage "deploy") | Out-Null
Copy-Item (Join-Path $root "deploy\cloudrun-start.sh") (Join-Path $stage "deploy")

if (Test-Path $zip) { Remove-Item -Force $zip }
# 必须用 tar.exe（Windows 自带 bsdtar）：Compress-Archive 生成的条目用反斜杠，
# Linux 云端构建会把 api\main.go 当成单个文件名，构建必然失败
tar -a -cf $zip -C $stage .
if ($LASTEXITCODE -ne 0) { throw "打包失败 (tar exit $LASTEXITCODE)" }

Write-Output ""
Write-Output "打包完成: $zip"
Write-Output "下一步：云托管控制台 → 你的服务 → 新建版本 → 上传此 zip"
