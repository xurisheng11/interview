# Advanced PDF OCR: PyMuPDF + PIL preprocessing + Tesseract
param(
    [Parameter(Mandatory=$true)][string]$PdfPath,
    [Parameter(Mandatory=$false)][string]$Pages = "all"
)

$ErrorActionPreference = "Stop"

function Debug($msg) {
    [Console]::Error.WriteLine("[OCR] $msg")
}

# Find Tesseract
$tesseract = $null
$tesseractPaths = @("C:\Program Files\Tesseract-OCR\tesseract.exe", "C:\Program Files (x86)\Tesseract-OCR\tesseract.exe")
foreach ($p in $tesseractPaths) { if (Test-Path $p) { $tesseract = $p; break } }
if (-not $tesseract) { $tesseract = (Get-Command tesseract -ErrorAction SilentlyContinue).Source }
if (-not $tesseract) {
    @{ success = $false; error = "Tesseract not found"; pages = @(); fullText = "" } | ConvertTo-Json -Compress
    exit 1
}

# Find Python
$python = $null
$pythonPaths = @("C:\Users\xuris\AppData\Local\Programs\Python\Python312\python.exe","C:\Users\xuris\AppData\Local\Programs\Python\Python313\python.exe","python","python3")
foreach ($p in $pythonPaths) {
    try {
        $result = & $p -c "import fitz; import PIL; print('ok')" 2>$null
        if ($result -eq "ok") { $python = $p; break }
    } catch {}
}
if (-not $python) {
    @{ success = $false; error = "Python not found"; pages = @(); fullText = "" } | ConvertTo-Json -Compress
    exit 1
}

$workDir = Join-Path $env:TEMP "pdf_ocr_$(Get-Random)"
New-Item -ItemType Directory -Path $workDir -Force | Out-Null

# Write OCR Python script
$ocrPy = Join-Path $workDir "ocr.py"
$ocrCode = @'
import sys, os, json, io
import fitz
from PIL import Image, ImageEnhance, ImageFilter
import pytesseract

pytesseract.pytesseract.tesseract_cmd = sys.argv[4]

pdf_path = sys.argv[1]
output_dir = sys.argv[2]
page_idx = int(sys.argv[3])

def preprocess_and_ocr(img):
    results = []
    g = img.convert('L')
    e = ImageEnhance.Contrast(g)
    g = e.enhance(2.5)
    g = g.filter(ImageFilter.SHARPEN)
    
    configs = [
        ("--psm 6 --oem 3", "chi_sim eng"),
        ("--psm 4 --oem 3", "chi_sim eng"),
        ("--psm 6 --oem 3", "chi_sim"),
    ]
    
    for config, lang in configs:
        try:
            text = pytesseract.image_to_string(g, lang=lang, config=config)
            if text and len(text.strip()) > 5:
                results.append((text.strip(), len(text)))
        except:
            pass
    
    if results:
        return max(results, key=lambda x: x[1])
    return ("", 0)

try:
    doc = fitz.open(pdf_path)
    total_pages = doc.page_count
    
    if page_idx < 0 or page_idx >= total_pages:
        print(json.dumps({"success": False, "error": "Invalid page"}))
        sys.exit(1)

    page = doc[page_idx]
    mat = fitz.Matrix(3.0, 3.0)
    pix = page.get_pixmap(matrix=mat, alpha=False)
    img_data = pix.tobytes("png")
    pix = None
    page = None
    doc.close()

    img = Image.open(io.BytesIO(img_data))
    text, char_count = preprocess_and_ocr(img)
    img.close()

    print(json.dumps({
        "success": True, 
        "text": text, 
        "charCount": char_count,
        "totalPages": total_pages
    }))

except Exception as e:
    import traceback
    traceback.print_exc()
    print(json.dumps({"success": False, "error": str(e)}))
'@
$ocrCode | Out-File -FilePath $ocrPy -Encoding UTF8

try {
    if (-not (Test-Path $PdfPath)) { throw "File not found: $PdfPath" }

    # Get page count
    $metaPy = Join-Path $workDir "meta.py"
    $metaCode = @'
import sys, fitz, json
try:
    doc = fitz.open(sys.argv[1])
    print(json.dumps({"totalPages": doc.page_count}))
    doc.close()
except Exception as e:
    print(json.dumps({"error": str(e)}))
'@
    $metaCode | Out-File -FilePath $metaPy -Encoding UTF8

    $metaOut = & $python $metaPy $PdfPath 2>$null
    $meta = $metaOut | ConvertFrom-Json
    if ($meta.error) { throw "Failed: $($meta.error)" }

    $totalPages = $meta.totalPages
    Debug "PDF has $totalPages pages"

    $result = @{ pages = @(); totalPages = $totalPages; success = $true }
    if ($Pages -eq "all") { $pageList = 0..($totalPages - 1) } else { $pageList = @([int]$Pages) }

    foreach ($pageIdx in $pageList) {
        Debug "Processing page $($pageIdx + 1)/$totalPages..."
        $ocrOut = & $python $ocrPy $PdfPath $workDir $pageIdx $tesseract 2>$null
        $ocrResult = $ocrOut | ConvertFrom-Json
        
        if ($ocrResult.success) {
            Debug "Page $($pageIdx + 1): $($ocrResult.charCount) chars"
            $result.pages += @{
                pageIndex = $pageIdx
                text = $ocrResult.text
                charCount = $ocrResult.charCount
            }
        } else {
            Debug "Page $($pageIdx + 1): failed - $($ocrResult.error)"
            $result.pages += @{
                pageIndex = $pageIdx
                text = ""
                charCount = 0
                error = $ocrResult.error
            }
        }
    }

    $result.fullText = ($result.pages | ForEach-Object { $_.text }) -join "`n`n"
    $result | ConvertTo-Json -Depth 5 -Compress

} catch {
    @{ success = $false; error = $_.Exception.Message; pages = @(); fullText = "" } | ConvertTo-Json -Compress
    exit 1
} finally {
    Remove-Item $workDir -Recurse -Force -ErrorAction SilentlyContinue
}
