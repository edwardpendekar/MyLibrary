#!/usr/bin/env python3
"""Turns an English book manuscript (.docx or .txt) into the Book/Chapter/
Verse/text_en/text_id/title_en/title_id CSV format the Go backend's admin
import pipeline already understands (see internal/service/import_parser.go).

Chapters are detected from the document itself (Word heading styles for
.docx, "Chapter N" / "Pasal N" lines for .txt) and numbered sequentially in
document order. Each chapter's English text is sent to the Gemini API once,
asking it to translate to Indonesian in the style of Alkitab Terjemahan Baru
edisi 2 (TB2) and split the chapter into short numbered verses.

Never called directly by an admin — invoked as a subprocess by
ImportService.Translate in internal/service/import_service.go, which then
feeds the resulting CSV through the normal upload/preview/commit flow.
"""

import argparse
import csv
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request

GEMINI_ENDPOINT = "https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent"

RESPONSE_SCHEMA = {
    "type": "OBJECT",
    "properties": {
        "title_id": {"type": "STRING"},
        "verses": {
            "type": "ARRAY",
            "items": {
                "type": "OBJECT",
                "properties": {
                    "text_en": {"type": "STRING"},
                    "text_id": {"type": "STRING"},
                },
                "required": ["text_en", "text_id"],
            },
        },
    },
    "required": ["title_id", "verses"],
}

PROMPT_TEMPLATE = """Anda adalah penerjemah profesional bahasa Inggris ke Indonesia, berpengalaman menerjemahkan teks keagamaan/sastra klasik.

Tugas:
1. Terjemahkan JUDUL PASAL berikut ke Bahasa Indonesia.
2. Terjemahkan ISI PASAL berikut ke Bahasa Indonesia dengan gaya bahasa seperti Alkitab Terjemahan Baru edisi 2 (TB2) terbitan LAI: formal namun wajar, sastrawi, jelas, dan mudah dibaca dalam Bahasa Indonesia baku kontemporer.
3. Pecah ISI PASAL menjadi ayat-ayat pendek bernomor, mengikuti gaya penomoran ayat Alkitab: setiap ayat berisi satu unit makna yang utuh (umumnya 1-3 kalimat). Pembagian ayat HARUS mengikuti pembagian kalimat/klausa alami dari teks aslinya -- jangan memotong di tengah kalimat, jangan menambah atau menghilangkan isi.
4. Setiap ayat punya pasangan teks asli (text_en, disalin persis dari sumber, TIDAK diterjemahkan) dan teks terjemahan (text_id).

JUDUL PASAL (EN):
{title_en}

ISI PASAL (EN):
{body_en}

Balas HANYA dengan JSON sesuai skema yang diberikan, tanpa markdown code fence, tanpa komentar tambahan."""

CHAPTER_LINE_RE = re.compile(r"^\s*(chapter|pasal)\s+([ivxlcdm\d]+)\b[:.\-]?\s*(.*)$", re.IGNORECASE)
# Fallback for manuscripts that open each chapter with a bare "1. Title" line
# (no "Chapter"/"Pasal" keyword) — since a plain number is far more likely to
# also appear as a numbered list item inside a chapter's body, this is only
# treated as a chapter boundary when its number continues the sequence
# (1, 2, 3, ...) starting from wherever the previous chapter left off.
NUMBER_TITLE_RE = re.compile(r"^\s*(\d+)[.):]\s+(.+?)\s*$")


class TranslationError(Exception):
    pass


def extract_chapters_docx(path):
    from docx import Document

    doc = Document(path)
    lines = [p.text for p in doc.paragraphs]
    is_heading = [bool(p.style and p.style.name.lower().startswith("heading")) for p in doc.paragraphs]
    return split_into_chapters(lines, is_heading, empty_message="no headings or body text found in the .docx file")


def extract_chapters_txt(path):
    with open(path, "r", encoding="utf-8", errors="replace") as f:
        content = f.read()
    return split_into_chapters(content.splitlines(), empty_message="the .txt file is empty")


def split_into_chapters(lines, is_heading=None, empty_message="the document is empty"):
    """Splits a document's lines/paragraphs into chapters, recognizing (in this
    order): a Word heading style (docx only, via is_heading); a "Chapter
    N"/"Pasal N" line; or a bare "N. Title" line whose number continues the
    chapter sequence (see NUMBER_TITLE_RE's comment). Falls back to treating
    the whole document as a single chapter if none of these ever match.
    """
    chapters = []
    current_title = None
    current_lines = []
    found_heading = False
    expected_next_number = 1

    def flush():
        nonlocal current_title, current_lines
        text = "\n".join(l for l in current_lines if l.strip())
        if current_title is not None or text.strip():
            chapters.append({"title_en": (current_title or "").strip(), "body_en": text})
        current_title = None
        current_lines = []

    for i, line in enumerate(lines):
        if is_heading and is_heading[i]:
            found_heading = True
            flush()
            current_title = line.strip()
            continue

        m = CHAPTER_LINE_RE.match(line)
        if m:
            found_heading = True
            flush()
            current_title = m.group(3).strip()
            continue

        m2 = NUMBER_TITLE_RE.match(line)
        if m2 and int(m2.group(1)) == expected_next_number:
            found_heading = True
            flush()
            current_title = m2.group(2).strip()
            expected_next_number += 1
            continue

        if line.strip():
            current_lines.append(line)
    flush()

    if not found_heading:
        text = "\n".join(l for l in lines if l.strip())
        if not text.strip():
            raise TranslationError(empty_message)
        return [{"title_en": "", "body_en": text}]
    return chapters


def translate_chapter(api_key, model, title_en, body_en, max_retries=3):
    prompt = PROMPT_TEMPLATE.format(title_en=title_en or "(tidak ada judul)", body_en=body_en)
    payload = {
        "contents": [{"parts": [{"text": prompt}]}],
        "generationConfig": {
            "temperature": 0.3,
            "maxOutputTokens": 8192,
            "responseMimeType": "application/json",
            "responseSchema": RESPONSE_SCHEMA,
        },
    }
    url = GEMINI_ENDPOINT.format(model=model)
    data = json.dumps(payload).encode("utf-8")
    delay = 5

    for attempt in range(max_retries + 1):
        req = urllib.request.Request(
            url,
            data=data,
            headers={"Content-Type": "application/json", "x-goog-api-key": api_key},
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=120) as resp:
                body = json.loads(resp.read().decode("utf-8"))
            candidate = body["candidates"][0]
            if candidate.get("finishReason") == "MAX_TOKENS":
                raise TranslationError(
                    "chapter terlalu panjang untuk sekali diterjemahkan (respons Gemini terpotong) -- "
                    "pecah pasal ini jadi beberapa bagian lebih kecil di dokumen sumber"
                )
            text = candidate["content"]["parts"][0]["text"]
            parsed = json.loads(text)
            if "title_id" not in parsed or "verses" not in parsed or not parsed["verses"]:
                raise ValueError("model response missing title_id/verses")
            return parsed
        except TranslationError:
            raise
        except urllib.error.HTTPError as e:
            body_text = e.read().decode(errors="replace")
            if (e.code == 429 or e.code >= 500) and attempt < max_retries:
                time.sleep(delay)
                delay *= 2
                continue
            raise TranslationError(f"Gemini API error {e.code}: {body_text}")
        except (urllib.error.URLError, TimeoutError) as e:
            if attempt < max_retries:
                time.sleep(delay)
                delay *= 2
                continue
            raise TranslationError(f"Gemini API request failed: {e}")
        except (KeyError, IndexError, json.JSONDecodeError, ValueError) as e:
            if attempt < max_retries:
                time.sleep(delay)
                delay *= 2
                continue
            raise TranslationError(f"Gemini API returned an unparseable response: {e}")

    raise TranslationError("unreachable")


def run(input_path, book_title, output_path, model, delay_seconds, api_key):
    ext = os.path.splitext(input_path)[1].lower()
    if ext == ".docx":
        chapters = extract_chapters_docx(input_path)
    elif ext == ".txt":
        chapters = extract_chapters_txt(input_path)
    else:
        raise TranslationError(f"unsupported file type {ext!r}: use .docx or .txt")

    with open(output_path, "w", encoding="utf-8", newline="") as out:
        writer = csv.writer(out, quoting=csv.QUOTE_ALL)
        writer.writerow(["book", "chapter", "verse", "text_en", "text_id", "title_en", "title_id"])

        for idx, chapter in enumerate(chapters, start=1):
            print(f"Translating chapter {idx}/{len(chapters)}...", file=sys.stderr, flush=True)
            try:
                result = translate_chapter(api_key, model, chapter["title_en"], chapter["body_en"])
            except TranslationError as e:
                raise TranslationError(f"chapter {idx} ('{chapter['title_en']}'): {e}") from e

            title_id = result.get("title_id", "")
            for v_num, verse in enumerate(result["verses"], start=1):
                writer.writerow(
                    [
                        book_title,
                        idx,
                        v_num,
                        verse.get("text_en", ""),
                        verse.get("text_id", ""),
                        chapter["title_en"],
                        title_id,
                    ]
                )

            if idx < len(chapters):
                time.sleep(delay_seconds)

    print(f"Translated {len(chapters)} chapter(s) -> {output_path}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", required=True, help="path to the .docx or .txt manuscript")
    parser.add_argument("--book", required=True, help="exact title of the existing book to import into")
    parser.add_argument("--output", required=True, help="path to write the generated CSV to")
    parser.add_argument("--model", default=os.environ.get("GEMINI_MODEL", "gemini-3.6-flash"))
    parser.add_argument(
        "--delay",
        type=float,
        default=float(os.environ.get("GEMINI_REQUEST_DELAY_SECONDS", "4.5")),
        help="seconds to sleep between chapters, to stay under the free-tier rate limit",
    )
    args = parser.parse_args()

    api_key = os.environ.get("GEMINI_API_KEY", "")
    if not api_key:
        print("GEMINI_API_KEY environment variable is required", file=sys.stderr)
        sys.exit(1)

    try:
        run(args.input, args.book, args.output, args.model, args.delay, api_key)
    except TranslationError as e:
        print(f"FATAL: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
