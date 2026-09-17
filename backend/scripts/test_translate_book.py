"""Tests for the pure text-splitting logic in translate_book.py — never the
Gemini API call itself, which needs network access and a real key. Run with:

    python3 -m unittest discover -s backend/scripts
"""

import os
import tempfile
import unittest

from translate_book import TranslationError, extract_chapters_txt


class ExtractChaptersTxtTests(unittest.TestCase):
    def write_tmp(self, content):
        fd, path = tempfile.mkstemp(suffix=".txt")
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(content)
        self.addCleanup(os.remove, path)
        return path

    def test_splits_on_chapter_headings(self):
        path = self.write_tmp(
            "Chapter 1: The Beginning\n"
            "In the beginning there was light.\n"
            "\n"
            "Chapter 2: The Middle\n"
            "Then came the middle of the story.\n"
        )
        chapters = extract_chapters_txt(path)
        self.assertEqual(len(chapters), 2)
        self.assertEqual(chapters[0]["title_en"], "The Beginning")
        self.assertIn("In the beginning", chapters[0]["body_en"])
        self.assertEqual(chapters[1]["title_en"], "The Middle")
        self.assertIn("middle of the story", chapters[1]["body_en"])

    def test_recognizes_pasal_prefix_case_insensitively(self):
        path = self.write_tmp("PASAL 1 Judul\nIsi pasal pertama.\n")
        chapters = extract_chapters_txt(path)
        self.assertEqual(len(chapters), 1)
        self.assertEqual(chapters[0]["title_en"], "Judul")

    def test_no_headings_treats_whole_document_as_one_chapter(self):
        path = self.write_tmp("Just a plain paragraph with no chapter markers at all.\n")
        chapters = extract_chapters_txt(path)
        self.assertEqual(len(chapters), 1)
        self.assertEqual(chapters[0]["title_en"], "")
        self.assertIn("plain paragraph", chapters[0]["body_en"])

    def test_empty_document_raises(self):
        path = self.write_tmp("   \n\n  ")
        with self.assertRaises(TranslationError):
            extract_chapters_txt(path)

    def test_heading_without_trailing_title_text(self):
        path = self.write_tmp("Chapter 1\nBody text only, no title after the number.\n")
        chapters = extract_chapters_txt(path)
        self.assertEqual(len(chapters), 1)
        self.assertEqual(chapters[0]["title_en"], "")
        self.assertIn("Body text only", chapters[0]["body_en"])


if __name__ == "__main__":
    unittest.main()
