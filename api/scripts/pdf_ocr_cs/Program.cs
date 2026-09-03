using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.WindowsRuntime;
using System.Text.Json;
using System.Threading.Tasks;
using Windows.Graphics.Imaging;
using Windows.Media.Ocr;
using Windows.Storage;
using Windows.Storage.Streams;

class Program
{
    static async Task<int> Main(string[] args)
    {
        if (args.Length < 1)
        {
            Console.Error.WriteLine("Usage: PdfOcr <pdfPath> [pages]");
            return 1;
        }

        string pdfPath = args[0];
        string pages = args.Length > 1 ? args[1] : "all";

        if (!File.Exists(pdfPath))
        {
            PrintError($"File not found: {pdfPath}");
            return 1;
        }

        try
        {
            var result = await OcrPdfAsync(pdfPath, pages);
            Console.WriteLine(JsonSerializer.Serialize(result, new JsonSerializerOptions { WriteIndented = false }));
            return 0;
        }
        catch (Exception ex)
        {
            PrintError(ex.Message);
            return 1;
        }
    }

    static void PrintError(string msg)
    {
        var err = new { success = false, error = msg, pages = Array.Empty<object>(), fullText = "" };
        Console.WriteLine(JsonSerializer.Serialize(err));
    }

    static async Task<OcrResult> OcrPdfAsync(string pdfPath, string pagesSpec)
    {
        var result = new OcrResult();

        // Get page count
        int totalPages;
        using (var img = Image.FromFile(pdfPath))
        {
            totalPages = img.GetFrameCount(FrameDimension.Page);
        }
        if (totalPages <= 0) totalPages = 1;
        result.totalPages = totalPages;

        // Determine pages to process
        int[] pageList;
        if (pagesSpec == "all")
        {
            pageList = Enumerable.Range(0, totalPages).ToArray();
        }
        else
        {
            pageList = new[] { int.Parse(pagesSpec) };
        }

        // Get available OCR language
        var ocrEngine = OcrEngine.TryCreateFromUserProfileLanguages();
        if (ocrEngine == null)
        {
            throw new Exception("Failed to create OcrEngine. No OCR language available.");
        }

        result.pages = new List<PageResult>();

        for (int idx = 0; idx < pageList.Length; idx++)
        {
            int pageIdx = pageList[idx];
            Console.Error.WriteLine($"[OCR] Processing page {pageIdx + 1}/{totalPages}...");

            // Render page to bitmap
            Bitmap? bmp = null;
            try
            {
                bmp = RenderPdfPage(pdfPath, pageIdx);
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"[OCR] Render failed for page {pageIdx + 1}: {ex.Message}");
                result.pages.Add(new PageResult { pageIndex = pageIdx, text = "", charCount = 0, error = ex.Message });
                continue;
            }

            string text = "";
            if (bmp != null)
            {
                try
                {
                    text = await OcrBitmapAsync(bmp, ocrEngine);
                    Console.Error.WriteLine($"[OCR] Page {pageIdx + 1}: extracted {text.Length} chars");
                }
                catch (Exception ex)
                {
                    Console.Error.WriteLine($"[OCR] OCR failed for page {pageIdx + 1}: {ex.Message}");
                }
                finally
                {
                    bmp.Dispose();
                }
            }

            result.pages.Add(new PageResult { pageIndex = pageIdx, text = text, charCount = text.Length });
        }

        result.success = true;
        result.fullText = string.Join("\n\n", result.pages.Select(p => p.text));

        return result;
    }

    static Bitmap? RenderPdfPage(string pdfPath, int pageIndex)
    {
        // System.Drawing.Common 8.0+ supports PDF natively on Windows
        using var source = Image.FromFile(pdfPath);
        source.SelectActiveFrame(FrameDimension.Page, pageIndex);

        // Clone to a new bitmap so we can dispose the source
        var bmp = new Bitmap(source.Width, source.Height, PixelFormat.Format32bppArgb);
        using (var g = Graphics.FromImage(bmp))
        {
            g.Clear(Color.White);
            g.DrawImage(source, 0, 0, source.Width, source.Height);
        }

        return bmp;
    }

    static async Task<string> OcrBitmapAsync(Bitmap bitmap, OcrEngine engine)
    {
        // Convert Bitmap to IRandomAccessStream
        using var ms = new MemoryStream();
        bitmap.Save(ms, ImageFormat.Png);
        ms.Position = 0;

        var randomAccessStream = new InMemoryRandomAccessStream();
        await randomAccessStream.WriteAsync(ms.ToArray().AsBuffer());
        randomAccessStream.Seek(0);

        // Decode as BitmapDecoder -> SoftwareBitmap
        var decoder = await BitmapDecoder.CreateAsync(randomAccessStream);
        var softwareBitmap = await decoder.GetSoftwareBitmapAsync(BitmapPixelFormat.Bgra8, BitmapAlphaMode.Premultiplied);

        // Run OCR
        var ocrResult = await engine.RecognizeAsync(softwareBitmap);
        return ocrResult.Text;
    }
}

class OcrResult
{
    public bool success { get; set; }
    public string? error { get; set; }
    public int totalPages { get; set; }
    public List<PageResult> pages { get; set; } = new();
    public string fullText { get; set; } = "";
}

class PageResult
{
    public int pageIndex { get; set; }
    public string text { get; set; } = "";
    public int charCount { get; set; }
    public string? error { get; set; }
}
