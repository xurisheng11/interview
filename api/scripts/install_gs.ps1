# Download and install Ghostscript for PDF rendering
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$gsDir = "C:\gs"
if (-not (Test-Path $gsDir)) {
    New-Item -ItemType Directory -Path $gsDir -Force | Out-Null
}

# Download Ghostscript 10.03.1 64-bit
$url = "https://github.com/ArtifexSoftware/ghostpdl-downloads/releases/download/gs10031/gs10031w64.exe"
$outFile = Join-Path $gsDir "gs_install.exe"

Write-Host "Downloading Ghostscript..."
try {
    Invoke-WebRequest -Uri $url -OutFile $outFile -UseBasicParsing -TimeoutSec 120
    Write-Host "Downloaded to $outFile"
    Write-Host "File size: $((Get-Item $outFile).Length / 1MB) MB"
} catch {
    Write-Host "Download failed: $_"
    Write-Host "Please manually install Ghostscript from: https://www.ghostscript.com/"
}
