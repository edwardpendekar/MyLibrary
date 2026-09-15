import { describe, expect, it } from "vitest";
import { sanitizeSnippet } from "./sanitize-snippet";

describe("sanitizeSnippet", () => {
  it("preserves ts_headline's <b> highlight tags", () => {
    expect(sanitizeSnippet("In the <b>beginning</b> God created")).toBe(
      "In the <b>beginning</b> God created"
    );
  });

  it("escapes a script tag embedded in verse text (stored-XSS regression)", () => {
    const malicious = 'Hello <script>alert(1)</script> <b>world</b>';
    const result = sanitizeSnippet(malicious);

    expect(result).not.toContain("<script>");
    expect(result).toContain("&lt;script&gt;");
    // The legitimate highlight tag must still survive untouched.
    expect(result).toContain("<b>world</b>");
  });

  it("escapes stray angle brackets and quotes that are not <b>/</b>", () => {
    const input = `He said "5 < 10" & <i>emphasis</i>`;
    const result = sanitizeSnippet(input);

    expect(result).toContain("&quot;5 &lt; 10&quot;");
    expect(result).toContain("&amp;");
    expect(result).toContain("&lt;i&gt;emphasis&lt;/i&gt;");
  });

  it("handles uppercase or malformed tag casing the same as lowercase", () => {
    const result = sanitizeSnippet("<B>shout</B>");
    expect(result).toBe("<b>shout</b>");
  });
});
