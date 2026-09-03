# PDF OCR using Windows.Media.Ocr (built-in Windows 10+ OCR engine)
# Best for PDFs with custom fonts - uses Windows' own OCR which handles more font types
param(
    [Parameter(Mandatory=$true)][string]$PdfPath,
    [Parameter(Mandatory=$false)][string]$Pages = "all"
)

$ErrorActionPreference = "Stop"

function Debug($msg) {
    [Console]::Error.WriteLine($msg)
}

# Create temp directory
$workDir = Join-Path $env:TEMP "pdf_winrt_ocr_$(Get-Random)"
New-Item -ItemType Directory -Path $workDir -Force | Out-Null

# Write Python render script
$renderPy = Join-Path $workDir "render.py"
$renderScript = @"
import sys, os, json, io
import fitz

pdf_path = sys.argv[1]
output_dir = sys.argv[2]
page_idx = int(sys.argv[3])

try:
    doc = fitz.open(pdf_path)
    total_pages = doc.page_count
    if page_idx < 0 or page_idx >= total_pages:
        print(json.dumps({'success': False, 'error': f'Invalid page index: {page_idx}'}))
        sys.exit(1)

    page = doc[page_idx]
    mat = fitz.Matrix(3.0, 3.0)  # 3x zoom for good quality
    pix = page.get_pixmap(matrix=mat, alpha=False)
    png_path = os.path.join(output_dir, f'page_{page_idx + 1}.png')
    pix.save(png_path)
    pix = None
    page = None
    doc.close()

    result = {'success': True, 'path': png_path, 'totalPages': total_pages}
    print(json.dumps(result))
except Exception as e:
    import traceback
    traceback.print_exc()
    print(json.dumps({'success': False, 'error': str(e)}))
"@
$renderScript | Out-File -FilePath $renderPy -Encoding UTF8

# Write WinRT OCR PowerShell script
$winrtPy = Join-Path $workDir "winrt_ocr.py"
$winrtScript = @"
import sys, os, json, io
import fitz

pdf_path = sys.argv[1]
output_dir = sys.argv[2]
page_idx = int(sys.argv[3])

try:
    doc = fitz.open(pdf_path)
    total_pages = doc.page_count
    if page_idx < 0 or page_idx >= total_pages:
        print(json.dumps({'success': False, 'error': f'Invalid page'}))
        sys.exit(1)

    page = doc[page_idx]
    # High DPI for better OCR
    mat = fitz.Matrix(4.0, 4.0)
    pix = page.get_pixmap(matrix=mat, alpha=False)
    png_path = os.path.join(output_dir, f'page_{page_idx + 1}.png')
    pix.save(png_path)
    pix = None
    page = None
    doc.close()

    result = {'success': True, 'path': png_path, 'totalPages': total_pages}
    print(json.dumps(result))
except Exception as e:
    import traceback
    traceback.print_exc()
    print(json.dumps({'success': False, 'error': str(e)}))
"@
$winrtScript | Out-File -FilePath $winrtPy -Encoding UTF8

try {
    if (-not (Test-Path $PdfPath)) {
        throw "File not found: $PdfPath"
    }

    # Find Python
    $python = $null
    $pythonPaths = @(
        "C:\Users\xuris\AppData\Local\Programs\Python\Python312\python.exe",
        "C:\Users\xuris\AppData\Local\Programs\Python\Python313\python.exe",
        "python",
        "python3"
    )
    foreach ($p in $pythonPaths) {
        try {
            $result = & $p -c "import fitz; print('ok')" 2>$null
            if ($result -eq "ok") { $python = $p; break }
        } catch {}
    }

    if (-not $python) {
        throw "Python with PyMuPDF not found"
    }

    # Get page count
    $metaPy = Join-Path $workDir "meta.py"
    $metaScript = @"
import sys
import fitz, json
try:
    doc = fitz.open(sys.argv[1])
    print(json.dumps({'totalPages': doc.page_count}))
    doc.close()
except Exception as e:
    print(json.dumps({'error': str(e)}))
"@
    $metaScript | Out-File -FilePath $metaPy -Encoding UTF8

    $metaOut = & $python $metaPy $PdfPath 2>$null
    $meta = $metaOut | ConvertFrom-Json
    if ($meta.error) {
        throw "Failed to open PDF: $($meta.error)"
    }
    $totalPages = $meta.totalPages
    Debug "[WinRT OCR] PDF has $totalPages pages"

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
        Debug "[WinRT OCR] Processing page $($pageIdx + 1)/$totalPages..."

        # Render PDF page to PNG
        $renderOut = & $python $renderPy $PdfPath $workDir $pageIdx 2>$null
        $renderResult = $renderOut | ConvertFrom-Json

        if (-not $renderResult.success) {
            $result.pages += @{
                pageIndex = $pageIdx
                text = ""
                charCount = 0
                error = $renderResult.error
            }
            continue
        }

        $pngPath = $renderResult.path
        if (-not (Test-Path $pngPath)) {
            $result.pages += @{
                pageIndex = $pageIdx
                text = ""
                charCount = 0
                error = "PNG not created"
            }
            continue
        }

        # Use Windows.Media.Ocr via Add-Type
        $ocrScript = @"
Add-Type -AssemblyName System.Runtime.WindowsRuntime
Add-Type -Path 'C:\Windows\System32\WinMetadata\Windows.Media.winmd'
Add-Type -AssemblyName Windows.Graphics
Add-Type -AssemblyName Windows.Storage

`$pngPath = '$pngPath'

# Load the PNG as a stream
`$file = [Windows.Storage.StorageFile]::GetFileFromPathAsync(`$pngPath).GetAwaiter().GetResult()
`$stream = `$file.OpenAsync([Windows.Storage.FileAccessMode]::Read).GetAwaiter().GetResult()
`$decoder = [Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync(`$stream).GetAwaiter().GetResult()
`$softwareBitmap = `$decoder.GetSoftwareBitmapAsync().GetAwaiter().GetResult()
`$stream.Close()

# Create OCR engine - try Chinese first, then English
`$langs = [Windows.Media.Ocr.OcrLanguage]::GetAvailableLanguages()
Debug "[WinRT] Available languages: `$(`$langs | ForEach-Object { `$_.DisplayName })"

`$engine = $null
# Try to use Chinese
foreach (`$lang in `$langs) {
    if (`$lang.LanguageTag -eq 'zh-Hans' -or `$lang.LanguageTag -eq 'zh-CN') {
        `$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage(`$lang)
        if (`$engine) {
            Debug "[WinRT] Using Chinese OCR: `$(`$lang.DisplayName)"
            break
        }
    }
}

if (-not `$engine) {
    # Fallback to any available language
    foreach (`$lang in `$langs) {
        if (`$lang.LanguageTag -eq 'en') {
            `$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromLanguage(`$lang)
            if (`$engine) {
                Debug "[WinRT] Using English OCR as fallback"
                break
            }
        }
    }
}

if (-not `$engine) {
    # Last resort: try from user profile languages
    `$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
    Debug "[WinRT] Using user profile languages"
}

if (`$engine) {
    `$ocrResult = `$engine.RecognizeAsync(`$softwareBitmap).GetAwaiter().GetResult()
    `$text = `$ocrResult.Text
    Debug "[WinRT] OCR extracted `$(`$text.Length) chars"
    Write-Output `$text
} else {
    Write-Error "No OCR engine available"
}
"@

        $ocrResult = powershell -ExecutionPolicy Bypass -Command $ocrScript 2>&1
        $ocrText = ($ocrResult | Where-Object { $_ -isnot [System.Management.Automation.ErrorRecord] }) -join "`n"

        if ($ocrText -and $ocrText.Length -gt 0) {
            $result.pages += @{
                pageIndex = $pageIdx
                text = $ocrText.Trim()
                charCount = $ocrText.Trim().Length
            }
            Debug "[WinRT OCR] Page $($pageIdx + 1): extracted $($ocrText.Trim().Length) chars"
        } else {
            $result.pages += @{
                pageIndex = $pageIdx
                text = ""
                charCount = 0
                error = "WinRT OCR returned empty"
            }
            Debug "[WinRT OCR] Page $($pageIdx + 1): empty result"
        }

        # Clean up PNG
        Remove-Item $pngPath -Force -ErrorAction SilentlyContinue
    }

    $result.fullText = ($result.pages | ForEach-Object { $_.text }) -join "`n`n"
    $result | ConvertTo-Json -Depth 5 -Compress

} catch {
    @{
        success = $false
        error = $_.Exception.Message
        pages = @()
        fullText = ""
    } | ConvertTo-Json -Compress
    exit 1
} finally {
    Remove-Item $workDir -Recurse -Force -ErrorAction SilentlyContinue
}
