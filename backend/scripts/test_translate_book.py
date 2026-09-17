"""Tests for the pure logic in translate_book.py — never the Gemini API call
itself, which needs network access and a real key. Run with:

    python3 -m unittest discover -s backend/scripts
"""

import os
import tempfile
import unittest
from unittest.mock import patch

from translate_book import TranslationError, run


class RunTests(unittest.TestCase):
    def write_tmp(self, content):
        fd, path = tempfile.mkstemp(suffix=".txt")
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(content)
        self.addCleanup(os.remove, path)
        return path

    def output_path(self):
        fd, path = tempfile.mkstemp(suffix=".csv")
        os.close(fd)
        os.remove(path)
        self.addCleanup(lambda: os.path.exists(path) and os.remove(path))
        return path

    def test_writes_one_row_per_translated_verse(self):
        input_path = self.write_tmp("In the beginning there was light. Then came the rest.")
        output_path = self.output_path()
        fake_result = {
            "title_id": "Permulaan",
            "verses": [
                {"text_en": "In the beginning there was light.", "text_id": "Pada mulanya ada terang."},
                {"text_en": "Then came the rest.", "text_id": "Kemudian datanglah selebihnya."},
            ],
        }
        with patch("translate_book.translate_chapter", return_value=fake_result) as mocked:
            run(input_path, "Some Book", 3, "The Beginning", output_path, "gemini-3.6-flash", "fake-key")
            mocked.assert_called_once_with(
                "fake-key", "gemini-3.6-flash", "The Beginning", "In the beginning there was light. Then came the rest."
            )

        with open(output_path, encoding="utf-8") as f:
            content = f.read()
        rows = content.strip().splitlines()
        self.assertEqual(len(rows), 3)  # header + 2 verses
        self.assertIn('"Some Book","3","1"', rows[1])
        self.assertIn("Pada mulanya ada terang.", rows[1])
        self.assertIn('"Some Book","3","2"', rows[2])
        self.assertIn("Permulaan", rows[2])

    def test_empty_chapter_body_raises(self):
        input_path = self.write_tmp("   \n\n  ")
        output_path = self.output_path()
        with self.assertRaises(TranslationError):
            run(input_path, "Some Book", 1, "", output_path, "gemini-3.6-flash", "fake-key")


if __name__ == "__main__":
    unittest.main()
