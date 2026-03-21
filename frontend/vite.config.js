import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";
import { fileURLToPath } from "node:url";
var __dirname = path.dirname(fileURLToPath(import.meta.url));
export default defineConfig({
    plugins: [vue()],
    root: __dirname,
    base: "./",
    build: {
        outDir: "dist",
        emptyOutDir: true,
    },
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "src"),
        },
    },
    server: {
        port: 5173,
        strictPort: true,
    },
});
