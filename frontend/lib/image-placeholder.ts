// A tiny shared base64 blur placeholder (1x1 neutral gray pixel scaled by the
// browser) used for every book cover. Generating a real per-image blurhash
// would require a server-side image-processing step (e.g. sharp) at upload
// time; this is a deliberate simplification that still gives the "blur while
// loading" UX the spec asks for without that extra pipeline.
export const COVER_BLUR_DATA_URL =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=";
