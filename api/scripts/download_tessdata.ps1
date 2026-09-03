# Download Tesseract language data
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$dest = "C:\tessdata"
if (-not (Test-Path $dest)) { New-Item -ItemType Directory -Path $dest -Force | Out-Null }

$files = @(
    "https://github.com/tesseract-ocr/tessdata_best/raw/main/chi_sim.traineddata",
    "https://github.com/tesseract-ocr/tessdata_best/raw/main/eng.traineddata"
)

foreach ($url in $files) {
    $name = Split-Path $url -Leaf
    $path = Join-Path $dest $name
    if (-not (Test-Path $path)) {
        Write-Host "Downloading $name..."
        Invoke-WebRequest -Uri $url -OutFile $path -UseBasicParsing -TimeoutSec 60
    }
}

Write-Host "Done. Files in tessdata:"
Get-ChildItem $dest | Select-Object Name, Length | Format-Table
