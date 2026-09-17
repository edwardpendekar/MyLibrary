/**
 * Wraps the selected substring of `text` in `marker` on both sides (e.g. "**"
 * for bold), returning the new full text plus where the selection should
 * land afterward — inside the markers, so typing continues right where the
 * user was rather than after the closing marker.
 */
export function wrapMarkdownSelection(
  text: string,
  selectionStart: number,
  selectionEnd: number,
  marker: string
): { next: string; selectionStart: number; selectionEnd: number } {
  const selected = text.slice(selectionStart, selectionEnd);
  const next = text.slice(0, selectionStart) + marker + selected + marker + text.slice(selectionEnd);
  return {
    next,
    selectionStart: selectionStart + marker.length,
    selectionEnd: selectionEnd + marker.length,
  };
}
