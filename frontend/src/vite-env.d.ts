/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_LEGAL_DOCS_BASE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

import "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    title?: string;
    bare?: boolean;
  }
}

declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<object, object, unknown>;
  export default component;
}

declare module "virtual:legal-md/*" {
  const content: string;
  export default content;
}
