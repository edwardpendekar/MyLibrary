import { describe, expect, it } from "vitest";
import { wrapMarkdownSelection } from "./wrap-markdown-selection";

describe("wrapMarkdownSelection", () => {
  it("wraps a selected word in the marker", () => {
    const result = wrapMarkdownSelection("hello world", 6, 11, "**");
    expect(result.next).toBe("hello **world**");
  });

  it("places the cursor selection inside the markers, after the shift", () => {
    // "world" is at [6, 11); after inserting "**" before it, the same text
    // now starts 2 characters later.
    const result = wrapMarkdownSelection("hello world", 6, 11, "**");
    expect(result.selectionStart).toBe(8);
    expect(result.selectionEnd).toBe(13);
    expect(result.next.slice(result.selectionStart, result.selectionEnd)).toBe("world");
  });

  it("inserts an empty pair of markers when nothing is selected (cursor only)", () => {
    const result = wrapMarkdownSelection("hello ", 6, 6, "*");
    expect(result.next).toBe("hello **");
    // Cursor lands between the two markers, ready to type.
    expect(result.selectionStart).toBe(7);
    expect(result.selectionEnd).toBe(7);
  });

  it("wraps a selection in the middle of longer text", () => {
    const result = wrapMarkdownSelection("The quick brown fox", 4, 9, "**");
    expect(result.next).toBe("The **quick** brown fox");
  });

  it("supports a multi-character marker for italic vs bold distinctly", () => {
    const bold = wrapMarkdownSelection("text", 0, 4, "**");
    const italic = wrapMarkdownSelection("text", 0, 4, "*");
    expect(bold.next).toBe("**text**");
    expect(italic.next).toBe("*text*");
  });
});
