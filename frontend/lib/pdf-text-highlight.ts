/**
 * Wraps every case-insensitive occurrence of `query` inside a PDF page's
 * rendered text layer in a `<mark>`, so a search actually shows *where* on
 * the page the match is instead of only which page it's on. Operates on the
 * parallel arrays pdfjs-dist's `TextLayer` exposes after rendering:
 * `textDivs[i]` is the DOM element for the text run `textContentItemsStr[i]`.
 */
export interface HighlightResult {
  /** Total occurrences of `query` found across every text div on the page. */
  totalMatches: number;
  /** The `<mark>` element for `activeMatchIndex`, if it was found. */
  activeElement: HTMLElement | null;
}

export function applyTextHighlight(
  textDivs: HTMLElement[],
  textContentItemsStr: string[],
  query: string,
  activeMatchIndex: number
): HighlightResult {
  const trimmed = query.trim();
  if (!trimmed) {
    for (let i = 0; i < textDivs.length; i++) {
      textDivs[i].textContent = textContentItemsStr[i];
    }
    return { totalMatches: 0, activeElement: null };
  }

  const needle = trimmed.toLowerCase();
  let totalMatches = 0;
  let activeElement: HTMLElement | null = null;

  for (let i = 0; i < textDivs.length; i++) {
    const div = textDivs[i];
    const text = textContentItemsStr[i];
    const lower = text.toLowerCase();

    if (!lower.includes(needle)) {
      div.textContent = text;
      continue;
    }

    let html = "";
    let cursor = 0;
    let idx = lower.indexOf(needle, cursor);
    while (idx !== -1) {
      const isActive = totalMatches === activeMatchIndex;
      html += escapeHtml(text.slice(cursor, idx));
      html += `<mark class="pdfSearchMatch${isActive ? " pdfSearchMatchActive" : ""}" data-match-index="${totalMatches}">${escapeHtml(
        text.slice(idx, idx + needle.length)
      )}</mark>`;
      cursor = idx + needle.length;
      totalMatches++;
      idx = lower.indexOf(needle, cursor);
    }
    html += escapeHtml(text.slice(cursor));
    div.innerHTML = html;

    if (!activeElement) {
      const found = div.querySelector<HTMLElement>("mark.pdfSearchMatchActive");
      if (found) activeElement = found;
    }
  }

  return { totalMatches, activeElement };
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}
