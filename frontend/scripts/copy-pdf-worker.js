// Copies the pdf.js worker into public/ so the PDF viewer can load it from
// same-origin, versioned to match the installed pdfjs-dist package exactly
// (mismatched worker/API versions throw at runtime).
const fs = require("fs");
const path = require("path");

const src = require.resolve("pdfjs-dist/build/pdf.worker.min.mjs");
const destDir = path.join(__dirname, "..", "public", "pdfjs");
const dest = path.join(destDir, "pdf.worker.min.mjs");

fs.mkdirSync(destDir, { recursive: true });
fs.copyFileSync(src, dest);
console.log(`Copied pdf.js worker to ${path.relative(process.cwd(), dest)}`);
