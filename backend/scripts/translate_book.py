#!/usr/bin/env python3
"""Translates one chapter's English text into the Book/Chapter/Verse/text_en/
text_id/title_en/title_id CSV format the Go backend's admin import pipeline
already understands (see internal/service/import_parser.go).

The admin pastes a single chapter's title + body (not a whole-book document —
splitting a long manuscript into many sequential Gemini calls turned out to
be slow and fragile against transient API errors). This script makes exactly
one Gemini call, asking it to translate to Indonesian in the style of
Alkitab Terjemahan Baru edisi 2 and split the chapter into short numbered
verses.

Never called directly by an admin — invoked as a subprocess by
ImportService.Translate in internal/service/import_service.go, which then
feeds the resulting single-chapter CSV through the normal upload/preview/
commit flow.
"""

import argparse
import csv
import json
import os
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


class TranslationError(Exception):
    pass


def translate_chapter(api_key, model, title_en, body_en, max_retries=5):
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
    max_delay = 30

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
                    "pasal terlalu panjang untuk sekali diterjemahkan (respons Gemini terpotong) -- "
                    "coba pecah jadi dua input pasal yang lebih pendek"
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
                delay = min(delay * 2, max_delay)
                continue
            raise TranslationError(f"Gemini API error {e.code}: {body_text}")
        except (urllib.error.URLError, TimeoutError) as e:
            if attempt < max_retries:
                time.sleep(delay)
                delay = min(delay * 2, max_delay)
                continue
            raise TranslationError(f"Gemini API request failed: {e}")
        except (KeyError, IndexError, json.JSONDecodeError, ValueError) as e:
            if attempt < max_retries:
                time.sleep(delay)
                delay = min(delay * 2, max_delay)
                continue
            raise TranslationError(f"Gemini API returned an unparseable response: {e}")

    raise TranslationError("unreachable")


def run(input_path, book_title, chapter_number, title_en, output_path, model, api_key):
    with open(input_path, "r", encoding="utf-8", errors="replace") as f:
        body_en = f.read().strip()
    if not body_en:
        raise TranslationError("chapter body is empty")

    result = translate_chapter(api_key, model, title_en, body_en)
    title_id = result.get("title_id", "")

    with open(output_path, "w", encoding="utf-8", newline="") as out:
        writer = csv.writer(out, quoting=csv.QUOTE_ALL)
        writer.writerow(["book", "chapter", "verse", "text_en", "text_id", "title_en", "title_id"])
        for v_num, verse in enumerate(result["verses"], start=1):
            writer.writerow(
                [book_title, chapter_number, v_num, verse.get("text_en", ""), verse.get("text_id", ""), title_en, title_id]
            )

    print(f"Translated chapter {chapter_number} ({len(result['verses'])} verses) -> {output_path}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", required=True, help="path to a plain UTF-8 text file with the chapter's English body")
    parser.add_argument("--book", required=True, help="exact title of the existing book to import into")
    parser.add_argument("--chapter", required=True, type=int, help="chapter number")
    parser.add_argument("--title", default="", help="chapter title in English (optional)")
    parser.add_argument("--output", required=True, help="path to write the generated CSV to")
    parser.add_argument("--model", default=os.environ.get("GEMINI_MODEL", "gemini-3.6-flash"))
    args = parser.parse_args()

    api_key = os.environ.get("GEMINI_API_KEY", "")
    if not api_key:
        print("GEMINI_API_KEY environment variable is required", file=sys.stderr)
        sys.exit(1)

    try:
        run(args.input, args.book, args.chapter, args.title, args.output, args.model, api_key)
    except TranslationError as e:
        print(f"FATAL: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
