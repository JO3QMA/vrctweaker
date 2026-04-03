import path from "node:path";
import { fileURLToPath } from "node:url";
import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** Default Wails dev bridge (see wails.json `devServer`, default localhost:34115). */
const defaultWailsDevTarget =
  process.env.VITE_WAILS_DEVSERVER_URL ?? "http://localhost:34115";

export default defineConfig(({ command }) => ({
  plugins: [
    vue(),
    {
      name: "wails-dev-ipc-scripts",
      transformIndexHtml(html) {
        // Production `wails build` lets the Go asset server inject these; avoid duplicates.
        if (command !== "serve") {
          return html;
        }
        if (html.includes('src="/wails/ipc.js"')) {
          return html;
        }
        return html.replace(
          "</head>",
          '  <script src="/wails/ipc.js"></script>\n  <script src="/wails/runtime.js"></script>\n</head>',
        );
      },
    },
  ],
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
    proxy: {
      // WebSocket IPC must match `window.location.host + "/wails/ipc"` (see wails ipc_websocket).
      "/wails/ipc": {
        target: defaultWailsDevTarget,
        changeOrigin: true,
        ws: true,
      },
      "/wails": {
        target: defaultWailsDevTarget,
        changeOrigin: true,
        ws: true,
      },
    },
  },
}));
