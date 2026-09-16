import { describe, expect, it } from "vitest";
import { applyTextHighlight } from "./pdf-text-highlight";

function makeDivs(texts: string[]): HTMLElement[] {
  return texts.map((t) => {
    const div = document.createElement("div");
    div.textContent = t;
    return div;
  });
}

describe("applyTextHighlight", () => {
  it("wraps a single match in a mark and counts it", () => {
    const divs = makeDivs(["In the beginning God created"]);
    const result = applyTextHighlight(divs, ["In the beginning God created"], "beginning", 0);

    expect(result.totalMatches).toBe(1);
    expect(divs[0].innerHTML).toBe(
      'In the <mark class="pdfSearchMatch pdfSearchMatchActive" data-match-index="0">beginning</mark> God created'
    );
    expect(result.activeElement).toBe(divs[0].querySelector("mark"));
  });

  it("is case-insensitive", () => {
    const divs = makeDivs(["BEGINNING of all things"]);
    const result = applyTextHighlight(divs, ["BEGINNING of all things"], "beginning", 0);
    expect(result.totalMatches).toBe(1);
  });

  it("counts multiple matches within one text run", () => {
    const divs = makeDivs(["cat and cat and cat"]);
    const result = applyTextHighlight(divs, ["cat and cat and cat"], "cat", 1);

    expect(result.totalMatches).toBe(3);
    const marks = divs[0].querySelectorAll("mark");
    expect(marks).toHaveLength(3);
    expect(marks[1].className).toContain("pdfSearchMatchActive");
    expect(marks[0].className).not.toContain("pdfSearchMatchActive");
    expect(marks[2].className).not.toContain("pdfSearchMatchActive");
  });

  it("counts matches across multiple text divs with a running global index", () => {
    const texts = ["the cat sat", "on the cat mat"];
    const divs = makeDivs(texts);
    const result = applyTextHighlight(divs, texts, "cat", 1);

    expect(result.totalMatches).toBe(2);
    // Global index 1 (the second "cat") is in the second div.
    expect(divs[0].querySelector("mark.pdfSearchMatchActive")).toBeNull();
    expect(divs[1].querySelector("mark.pdfSearchMatchActive")).not.toBeNull();
    expect(result.activeElement).toBe(divs[1].querySelector("mark.pdfSearchMatchActive"));
  });

  it("escapes HTML special characters so a match never breaks out of text content", () => {
    // A verse/PDF containing a literal "<script>" must not become real markup
    // once wrapped — only our own single <mark> element may exist afterward.
    const texts = ['before <script>alert(1)</script> cat after'];
    const divs = makeDivs(texts);
    applyTextHighlight(divs, texts, "cat", 0);

    expect(divs[0].querySelectorAll("script")).toHaveLength(0);
    expect(divs[0].querySelectorAll("mark")).toHaveLength(1);
    // Round-tripping through the DOM recovers the exact original text once
    // the <mark> wrapper is accounted for.
    expect(divs[0].textContent).toBe(texts[0]);
  });

  it("restores plain text and reports zero matches for an empty query", () => {
    const texts = ["some text here"];
    const divs = makeDivs(texts);
    divs[0].innerHTML = '<mark class="pdfSearchMatch">some</mark> text here';

    const result = applyTextHighlight(divs, texts, "", 0);

    expect(result.totalMatches).toBe(0);
    expect(result.activeElement).toBeNull();
    expect(divs[0].textContent).toBe("some text here");
    expect(divs[0].innerHTML).not.toContain("<mark");
  });

  it("leaves non-matching divs as plain text", () => {
    const texts = ["no match here", "cat found here"];
    const divs = makeDivs(texts);
    const result = applyTextHighlight(divs, texts, "cat", 0);

    expect(result.totalMatches).toBe(1);
    expect(divs[0].innerHTML).toBe("no match here");
    expect(divs[0].querySelector("mark")).toBeNull();
  });
});
