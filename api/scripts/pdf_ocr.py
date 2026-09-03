# PDF OCR using PyMuPDF + Tesseract
# Usage: python pdf_ocr.py <pdf_path> [pages]
import sys
import os
import json
import fitz
from PIL import Image, ImageEnhance, ImageFilter
import pytesseract
import io
import tempfile
import shutil

# Tesseract configuration
TESSERACT_CMD = r"C:\Program Files\Tesseract-OCR\tesseract.exe"
pytesseract.pytesseract.tesseract_cmd = TESSERACT_CMD

# TESSDATA_PREFIX environment
os.environ["TESSDATA_PREFIX"] = r"C:\tessdata"

def preprocess_image(img):
    """Preprocess image for better OCR"""
    # Convert to grayscale
    gray = img.convert('L')
    # Enhance contrast
    enhancer = ImageEnhance.Contrast(gray)
    enhanced = enhancer.enhance(2.5)
    # Sharpen
    sharpened = enhanced.filter(ImageFilter.SHARPEN)
    return sharpened

def ocr_page(img, lang="chi_sim eng"):
    """OCR a single page with multiple config attempts"""
    configs = [
        ("--psm 6 --oem 3", lang),
        ("--psm 4 --oem 3", lang),
        ("--psm 6 --oem 3", "chi_sim"),
        ("--psm 6 --oem 3", "eng"),
    ]
    
    results = []
    for config, lang_cfg in configs:
        try:
            text = pytesseract.image_to_string(img, lang=lang_cfg, config=config)
            if text and len(text.strip()) > 10:
                chinese_count = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
                results.append((text.strip(), len(text), chinese_count))
        except Exception as e:
            continue
    
    if results:
        # Prefer result with most Chinese characters
        results.sort(key=lambda x: x[2], reverse=True)
        return results[0][0]
    return ""

def pdf_to_images(pdf_path, dpi=300):
    """Convert PDF pages to images"""
    images = []
    doc = fitz.open(pdf_path)
    for page_num in range(doc.page_count):
        page = doc[page_num]
        # Render at high DPI for better OCR
        mat = fitz.Matrix(dpi/72, dpi/72)
        pix = page.get_pixmap(matrix=mat, alpha=False)
        img_data = pix.tobytes("png")
        img = Image.open(io.BytesIO(img_data))
        images.append((page_num, img))
    doc.close()
    return images

def extract_text_by_ocr(pdf_path):
    """Main OCR extraction function"""
    result = {
        "success": False,
        "pages": [],
        "fullText": "",
        "totalPages": 0,
        "error": ""
    }
    
    try:
        # First try to extract text directly (fastest)
        doc = fitz.open(pdf_path)
        total_pages = doc.page_count
        result["totalPages"] = total_pages
        
        # Check if PDF has selectable text
        has_text = False
        for i in range(min(3, total_pages)):  # Check first 3 pages
            page = doc[i]
            text = page.get_text("text")
            if text and len(text.strip()) > 50:
                has_text = True
                break
        
        if has_text:
            # Extract text directly (fast, works for text PDFs)
            for page_num in range(total_pages):
                page = doc[page_num]
                text = page.get_text("text")
                chinese_count = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
                result["pages"].append({
                    "pageIndex": page_num,
                    "text": text.strip(),
                    "charCount": len(text),
                    "chineseCount": chinese_count
                })
            doc.close()
            result["success"] = True
            result["fullText"] = "\n\n".join(p["text"] for p in result["pages"])
            return result
        
        doc.close()
        
        # PDF has no selectable text - use OCR
        images = pdf_to_images(pdf_path, dpi=300)
        
        for page_num, img in images:
            processed = preprocess_image(img)
            text = ocr_page(processed)
            chinese_count = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
            result["pages"].append({
                "pageIndex": page_num,
                "text": text,
                "charCount": len(text),
                "chineseCount": chinese_count
            })
            img.close()
        
        result["success"] = True
        result["fullText"] = "\n\n".join(p["text"] for p in result["pages"])
        
    except Exception as e:
        import traceback
        traceback.print_exc()
        result["error"] = str(e)
    
    return result

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"success": False, "error": "Usage: python pdf_ocr.py <pdf_path>"}))
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    if not os.path.exists(pdf_path):
        print(json.dumps({"success": False, "error": f"File not found: {pdf_path}"}))
        sys.exit(1)
    
    result = extract_text_by_ocr(pdf_path)
    print(json.dumps(result, ensure_ascii=False))
