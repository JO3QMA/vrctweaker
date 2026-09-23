import { copyFileSync, mkdirSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const frontendRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const source = resolve(frontendRoot, "../build/appicon.png");
const publicDir = resolve(frontendRoot, "public");
const dest = resolve(publicDir, "appicon.png");

mkdirSync(publicDir, { recursive: true });
copyFileSync(source, dest);
