<template>
  <span class="bio-link-site-icon" :data-site="site" aria-hidden="true">
    <svg
      v-if="site === 'twitter'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"
      />
    </svg>
    <svg
      v-else-if="site === 'booth'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M4 4h16v16H4V4zm3.2 3.2v9.6h2.4V9.6h1.8l2.2 4.2 2.2-4.2h1.8v7.2h2.4V7.2H7.2z"
      />
    </svg>
    <svg
      v-else-if="site === 'misskey'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M12 3c-1.8 0-3.3.9-4.2 2.3C6 4.2 4.4 3.5 2.7 3.5 2.3 7.2 3.4 10.2 5.7 12.5 5.2 13.8 5 15.2 5 16.6c0 2.2 1.2 4.1 3 5.2.5-1.5 1.4-2.8 2.6-3.8 1.2 1 2.1 2.3 2.6 3.8 1.8-1.1 3-3 3-5.2 0-1.4-.2-2.8-.7-4.1 2.3-2.3 3.4-5.3 3-9-.9 0-2.3.7-3.1 1.8C15.3 3.9 13.8 3 12 3z"
      />
    </svg>
    <svg
      v-else-if="site === 'github'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M12 2C6.477 2 2 6.484 2 12.021c0 4.428 2.865 8.184 6.839 9.504.5.092.682-.217.682-.483 0-.237-.009-.868-.014-1.703-2.782.605-3.369-1.342-3.369-1.342-.454-1.157-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.004.07 1.532 1.032 1.532 1.032.892 1.53 2.341 1.089 2.91.833.092-.647.35-1.089.636-1.339-2.221-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.378.202 2.397.1 2.65.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.021C22 6.484 17.523 2 12 2z"
      />
    </svg>
    <svg
      v-else-if="site === 'youtube'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M23.5 6.2a3 3 0 0 0-2.1-2.1C19.5 3.5 12 3.5 12 3.5s-7.5 0-9.4.6A3 3 0 0 0 .5 6.2 31.5 31.5 0 0 0 0 12a31.5 31.5 0 0 0 .5 5.8 3 3 0 0 0 2.1 2.1c1.9.6 9.4.6 9.4.6s7.5 0 9.4-.6a3 3 0 0 0 2.1-2.1A31.5 31.5 0 0 0 24 12a31.5 31.5 0 0 0-.5-5.8zM9.75 15.5v-7l6 3.5-6 3.5z"
      />
    </svg>
    <svg
      v-else-if="site === 'discord'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M20.3 4.4A19.7 19.7 0 0 0 15.5 3c-.2.4-.5 1-.7 1.4a18.2 18.2 0 0 0-5.6 0C8.9 4 8.6 3.4 8.4 3a19.8 19.8 0 0 0-4.8 1.4C1.2 8.8.4 13.1.9 17.3a19.9 19.9 0 0 0 6.1 3.1c.5-.7.9-1.4 1.2-2.2-.7-.3-1.3-.6-1.9-1 .2-.1.3-.2.5-.3 3.6 1.7 7.5 1.7 11 0 .2.1.3.2.5.3-.6.4-1.2.7-1.9 1 .3.8.7 1.5 1.2 2.2a19.8 19.8 0 0 0 6.1-3.1c.6-4.8-.1-9.1-2.7-12.9zM8.3 14.9c-1.1 0-2-.9-2-2.1s.9-2.1 2-2.1 2 1 2.1 2-1 2.1-2 2.1zm7.4 0c-1.1 0-2-.9-2-2.1s.9-2.1 2-2.1 2 1 2.1 2-1 2.1-2 2.1z"
      />
    </svg>
    <svg
      v-else-if="site === 'twitch'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M4 3 2 7.2v12.8h5v3l3-3h4.2L22 13.2V3H4zm16 8.8-3.2 3.2H12l-2.4 2.4v-2.4H6.8V5h13.2v6.8zM15 7.8h2v4.4h-2V7.8zm-5 0h2v4.4h-2V7.8z"
      />
    </svg>
    <svg
      v-else-if="site === 'vrchat'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M4 5h16l-1.4 12.2a2 2 0 0 1-2 1.8H7.4a2 2 0 0 1-2-1.8L4 5zm4.2 3.2 1.6 6.4 1.6-3.2 1.6 3.2 1.6-6.4h2.4l-2.4 9.6-1.6-3.2-1.6 3.2-2.4-9.6H8.2z"
      />
    </svg>
    <svg
      v-else-if="site === 'pixiv'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M4 4h16v16H4V4zm3 3h4.2c3.4 0 5.8 2.1 5.8 5.3S14.6 17.6 11.2 17.6H7V7z"
      />
    </svg>
    <svg
      v-else-if="site === 'instagram'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M7 3h10a4 4 0 0 1 4 4v10a4 4 0 0 1-4 4H7a4 4 0 0 1-4-4V7a4 4 0 0 1 4-4zm10 2H7a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2zm-5 3.5A4.5 4.5 0 1 1 7.5 13 4.5 4.5 0 0 1 12 8.5zm0 2A2.5 2.5 0 1 0 14.5 13 2.5 2.5 0 0 0 12 10.5zM17.8 7.6a1 1 0 1 1-1 1 1 1 0 0 1 1-1z"
      />
    </svg>
    <svg
      v-else-if="site === 'tiktok'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M16.5 3h2.7c.2 1.8 1.2 3.4 2.8 4.3v2.5c-1.4-.1-2.7-.5-3.8-1.2v6.9a6.4 6.4 0 1 1-5.6-6.3v2.7a3.7 3.7 0 1 0 2.6 3.5V3z"
      />
    </svg>
    <svg
      v-else-if="site === 'bluesky'"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M12 10.8c-1.9-3.5-7.1-9.6-12-3.6 2.4 4.6 6.1 8.8 12 8.8s9.6-4.2 12-8.8c-4.9-6-10.1.1-12 3.6z"
      />
    </svg>
    <svg
      v-else
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M10.59 13.41a1.996 1.996 0 0 0 .03-2.82l-2.5-2.5a2 2 0 0 0-2.83 2.83l.79.79-1.06 1.06-.79-.79a4 4 0 1 1 5.66-5.66l2.5 2.5a4 4 0 0 1 0 5.66l-1.42 1.41-1.06-1.06 1.42-1.41zM13.41 10.59a1.996 1.996 0 0 0-.03 2.82l2.5 2.5a2 2 0 0 0 2.83-2.83l-.79-.79 1.06-1.06.79.79a4 4 0 1 1-5.66 5.66l-2.5-2.5a4 4 0 0 1 0-5.66l1.42-1.41 1.06 1.06-1.42 1.41z"
      />
    </svg>
  </span>
</template>

<script setup lang="ts">
import type { BioLinkSite } from "../utils/bioLinkHost";

defineProps<{
  site: BioLinkSite;
}>();
</script>

<style scoped>
.bio-link-site-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--icon-size-16);
  height: var(--icon-size-16);
  flex-shrink: 0;
}

.bio-link-site-icon svg {
  width: 100%;
  height: 100%;
}
</style>
