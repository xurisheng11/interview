import sys
import fitz
import json
import io
import tempfile
import os

def extract_text_from_pdf(pdf_path: str) -> dict:
    """Extract text from PDF using PyMuPDF. Returns dict with text and metadata."""
    try:
        doc = fitz.open(pdf_path)
        total_pages = doc.page_count
        all_text = []

        for page_num in range(doc.page_count):
            page = doc[page_num]

            # Method 1: Try get_text("text") - most reliable for text PDFs
            text = page.get_text("text")
            if text and text.strip():
                chinese_count = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
                readable_count = sum(1 for c in text if c.isprintable() and c not in '��')

                if chinese_count > 5 or readable_count > len(text) * 0.5:
                    all_text.append(text.strip())
                    continue

            # Method 2: Try HTML extraction
            html = page.get_text("html")
            if html:
                import re
                spans = re.findall(r'<span[^>]*>([^<]+)</span>', html)
                html_text = []
                for span in spans:
                    decoded = html.unescape(span)
                    decoded = re.sub(r'\s+', ' ', decoded).strip()
                    if decoded:
                        html_text.append(decoded)
                if html_text:
                    combined = ''.join(html_text)
                    chinese_count = sum(1 for c in combined if '\u4e00' <= c <= '\u9fff')
                    if chinese_count > 5:
                        all_text.append(combined)
                        continue

            # Method 3: Try rawdict
            rawdict = page.get_text("rawdict")
            raw_text = []
            for block in rawdict.get("blocks", []):
                if block.get("type") == 0:
                    for line in block.get("lines", []):
                        for span in line.get("spans", []):
                            t = span.get("text", "")
                            if t.strip():
                                raw_text.append(t)
            if raw_text:
                combined = ''.join(raw_text)
                chinese_count = sum(1 for c in combined if '\u4e00' <= c <= '\u9fff')
                if chinese_count > 5:
                    all_text.append(combined)
                    continue

            if text:
                all_text.append(text.strip())

        doc.close()
        total_pages = len(all_text)  # Use all_text length as proxy

        full_text = '\n\n'.join(all_text)
        chinese_count = sum(1 for c in full_text if '\u4e00' <= c <= '\u9fff')

        return {
            'success': True,
            'text': full_text,
            'charCount': len(full_text),
            'chineseCount': chinese_count,
            'method': 'pymupdf_text',
            'pages': total_pages
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {
            'success': False,
            'error': str(e),
            'text': '',
            'charCount': 0,
            'chineseCount': 0
        }


if __name__ == "__main__":
    # Construct path using UTF-8 bytes
    dir_bytes = b'\xe6\x9b\xb9\xe4\xbc\xa0\xe6\x80\xa1\xe7\xae\x80\xe5\x8e\x86'
    pdf_path = "e:/" + dir_bytes.decode('utf-8') + "/7.27版本(1).pdf"

    result = extract_text_from_pdf(pdf_path)

    # Save to file to avoid encoding issues
    with open("e:/interview/api/scripts/extract_result.json", "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)

    print(f"Success: {result['success']}")
    print(f"Chinese chars: {result.get('chineseCount', 0)}/{result.get('charCount', 0)}")
    if result.get('text'):
        print(f"First 200 chars: {result['text'][:200]}")
