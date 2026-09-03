# PDF OCR using Tesseract
# Prerequisites:
#   - Tesseract 5.x installed (https://github.com/tesseract-ocr/tesseract)
#   - Ghostscript installed (https://www.ghostscript.com/) for PDF->image conversion
#   - Language data in tessdata folder

param(
    [Parameter(Mandatory=$true)][string]$PdfPath,
    [Parameter(Mandatory=$false)][string]$Pages = "all"
)

$ErrorActionPreference = "Stop"
$tesseract = "C:\Program Files\Tesseract-OCR\tesseract.exe"

if (-not (Test-Path $tesseract)) {
    $tesseract = "tesseract"
}

# Find Ghostscript
$gsExe = $null
$gsPaths = @(
    "C:\Program Files\gs\gs10.03.1\bin\gswin64c.exe",
    "C:\Program Files\gs\gs10.02.2\bin\gswin64c.exe",
    "C:\Program Files\gs\gs9.56.1\bin\gswin64c.exe",
    "C:\Program Files\gs\gs9.55.0\bin\gswin64c.exe",
    "C:\Program Files\gs\bin\gswin64c.exe",
    "C:\Program Files (x86)\gs\gs9.56.1\bin\gswin64c.exe",
    "C:\Program Files (x86)\gs\bin\gswin64c.exe"
)
foreach ($p in $gsPaths) {
    if (Test-Path $p) { $gsExe = $p; break }
}

if (-not $gsExe) {
    # Try to find any gswin64c
    $gsExe = (Get-Command gswin64c -ErrorAction SilentlyContinue).Source
    if (-not $gsExe) { $gsExe = (Get-Command gswin32c -ErrorAction SilentlyContinue).Source }
}

if (-not $gsExe) {
    @{
        success = $false
        error = "Ghostscript not found. Please install from https://www.ghostscript.com/"
        pages = @()
    } | ConvertTo-Json -Compress
    exit 1
}

Write-Host "Using Ghostscript: $gsExe"

# Setup
$workDir = Join-Path $env:TEMP "pdf_ocr_$(Get-Random)"
New-Item -ItemType Directory -Path $workDir -Force | Out-Null

# Find tessdata
$tessdata = $null
$tessPaths = @(
    "C:\tessdata",
    "C:\Program Files\Tesseract-OCR\tessdata",
    "C:\Program Files (x86)\Tesseract-OCR\tessdata",
    ".\tessdata"
)
foreach ($p in $tessPaths) {
    if ((Test-Path $p) -and (Get-ChildItem $p -Filter "chi_sim.traineddata" -ErrorAction SilentlyContinue)) {
        $tessdata = $p; break
    }
}

if (-not $tessdata) {
    # Try to find in PATH
    $tessExe = (Get-Command tesseract -ErrorAction SilentlyContinue).Source
    if ($tessExe) {
        $tessdata = Split-Path $tessExe
        $tessdata = Split-Path $tessdata
        $tessdata = Join-Path $tessdata "tessdata"
    }
}

if (-not $tessdata) {
    @{
        success = $false
        error = "Tesseract language data (chi_sim.traineddata) not found. Download from https://github.com/tesseract-ocr/tessdata and place in tessdata folder."
        pages = @()
    } | ConvertTo-Json -Compress
    exit 1
}

Write-Host "Using tessdata: $tessdata"

# Get page count using Ghostscript
function Get-PdfPageCount($pdfPath) {
    $output = & $gsExe -q -dNODISPLAY -dBATCH -dNOPAUSE -sPDFname="$pdfPath" -c "(PDFname) run (r) file closefile /PageCount get = quit" 2>&1
    $output -match '\d+' | Out-Null
    if ($matches) { return [int]$matches[0] }
    return 1
}

# Convert PDF page to image using Ghostscript
function Convert-PdfPageToPng($pdfPath, $pageIndex, $outputPath) {
    # Ghostscript: render at 300 DPI for good OCR quality
    $gsArgs = @(
        "-dNOPAUSE",
        "-dBATCH",
        "-sDEVICE=png16m",
        "-r300",
        "-dPDFFitPage",
        "-dTextAlphaBits=4",
        "-sOutputFile=$outputPath",
        "-dFirstPage=$($pageIndex + 1)",
        "-dLastPage=$($pageIndex + 1)",
        "`"$pdfPath`""
    )
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $gsExe
    $psi.Arguments = $gsArgs -join " "
    $psi.UseShellExecute = $false
    $psi.RedirectStandardError = $true
    $psi.RedirectStandardOutput = $true
    $psi.CreateNoWindow = $true
    $proc = [System.Diagnostics.Process]::Start($psi)
    $stderr = $proc.StandardError.ReadToEnd()
    $proc.WaitForExit()
    if ($proc.ExitCode -ne 0) {
        Write-Warning "Ghostscript failed: $stderr"
        return $false
    }
    return Test-Path $outputPath
}

# Run Tesseract OCR on image
function Get-TesseractOcr($imagePath) {
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $tesseract
    $psi.Arguments = "`"$imagePath`" stdout -l chi_sim eng --psm 6"
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.EnvironmentVariables["TESSDATA_PREFIX"] = $tessdata
    $psi.CreateNoWindow = $true
    $proc = [System.Diagnostics.Process]::Start($psi)
    $stdout = $proc.StandardOutput.ReadToEnd()
    $stderr = $proc.StandardError.ReadToEnd()
    $proc.WaitForExit()
    return $stdout.Trim()
}

# Main
try {
    if (-not (Test-Path $PdfPath)) {
        throw "File not found: $PdfPath"
    }

    $totalPages = Get-PdfPageCount -pdfPath $PdfPath
    Write-Host "[OCR] PDF has $totalPages pages"

    $result = @{
        pages = @()
        totalPages = $totalPages
        success = $true
    }

    if ($Pages -eq "all") {
        $pageList = 0..($totalPages - 1)
    } else {
        $pageList = @([int]$Pages)
    }

    foreach ($pageIdx in $pageList) {
        Write-Host "[OCR] Processing page $($pageIdx + 1)/$totalPages..."

        $pngPath = Join-Path $workDir "page_$($pageIdx + 1).png"

        $ok = Convert-PdfPageToPng -pdfPath $PdfPath -pageIndex $pageIdx -outputPath $pngPath

        if ($ok -and (Test-Path $pngPath)) {
            $text = Get-TesseractOcr -imagePath $pngPath
            Remove-Item $pngPath -Force -ErrorAction SilentlyContinue

            $result.pages += @{
                pageIndex = $pageIdx
                text = $text
                charCount = $text.Length
            }
            Write-Host "[OCR] Page $($pageIdx + 1): extracted $($text.Length) chars"
        } else {
            $result.pages += @{
                pageIndex = $pageIdx
                text = ""
                charCount = 0
                error = "PDF render failed"
            }
            Write-Host "[OCR] Page $($pageIdx + 1): render failed"
        }
    }

    $result.fullText = ($result.pages | ForEach-Object { $_.text }) -join "`n`n"
    $result | ConvertTo-Json -Depth 5 -Compress

} catch {
    @{
        success = $false
        error = $_.Exception.Message
        pages = @()
    } | ConvertTo-Json -Compress
    exit 1
} finally {
    Remove-Item $workDir -Recurse -Force -ErrorAction SilentlyContinue
}
