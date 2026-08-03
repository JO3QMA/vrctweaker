import type { Meta, StoryObj } from "@storybook/vue3-vite";
import type { SimpleColorToken } from "./colorTokens";
import {
  BRAND_COLOR_TOKENS,
  NEUTRAL_COLOR_TOKENS,
  PRESENCE_COLOR_TOKENS,
  SEMANTIC_COLOR_TOKENS,
  SERVER_STATUS_COLOR_TOKENS,
} from "./colorTokens";
import "./Color.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Design System/Color",
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

function swatchStory(
  title: string,
  description: string,
  rows: ReadonlyArray<Pick<SimpleColorToken, "varName">>,
): Story {
  return {
    name: title,
    render: () => ({
      setup: () => ({ title, description, rows }),
      template: `
        <div class="color-story">
          <h2>{{ title }}</h2>
          <p>{{ description }}</p>
          <div class="color-story-grid">
            <div v-for="row in rows" :key="row.varName" class="color-story-swatch">
              <div
                class="color-story-swatch-chip"
                :style="{ background: 'var(' + row.varName + ')' }"
              />
              <span class="color-story-swatch-label">{{ row.varName }}</span>
            </div>
          </div>
        </div>
      `,
    }),
  };
}

export const Brand: Story = swatchStory(
  "Brand color",
  "Main accent. Primary button fill and links use --color-brand.",
  BRAND_COLOR_TOKENS,
);

export const Neutral: Story = {
  name: "Neutral color",
  render: () => ({
    setup: () => ({ rows: NEUTRAL_COLOR_TOKENS }),
    template: `
      <div class="color-story">
        <h2>Neutral color</h2>
        <p>Surfaces, text, and border. Prefer these over legacy --bg-* / --text-* aliases.</p>
        <div class="color-story-grid">
          <div v-for="row in rows" :key="row.varName" class="color-story-swatch">
            <div
              v-if="row.name.startsWith('text-')"
              class="color-story-swatch-chip color-story-swatch-chip--text"
              :style="{ color: 'var(' + row.varName + ')' }"
            >Aa</div>
            <div
              v-else-if="row.name === 'border'"
              class="color-story-swatch-chip"
              :style="{ background: 'var(--color-bg-base)', border: '3px solid var(' + row.varName + ')' }"
            />
            <div
              v-else
              class="color-story-swatch-chip"
              :style="{ background: 'var(' + row.varName + ')' }"
            />
            <span class="color-story-swatch-label">{{ row.varName }}</span>
            <span v-if="row.legacyAlias" class="color-story-swatch-label">alias: {{ row.legacyAlias }}</span>
          </div>
        </div>
      </div>
    `,
  }),
};

export const Semantic: Story = swatchStory(
  "Semantic color catalog",
  "Feedback for tags, messages, and validation. Not for Server status or Presence.",
  SEMANTIC_COLOR_TOKENS,
);

export const ServerStatus: Story = swatchStory(
  "Server status color",
  "Domain colors for Dashboard Server status. Do not reuse as Semantic feedback.",
  SERVER_STATUS_COLOR_TOKENS,
);

export const Presence: Story = {
  name: "Presence color",
  render: () => ({
    setup: () => ({ rows: PRESENCE_COLOR_TOKENS }),
    template: `
      <div class="color-story">
        <h2>Presence color</h2>
        <p>Domain colors for Presence change Semantic buttons.</p>
        <div class="color-story-presence-row">
          <div
            v-for="row in rows"
            :key="row.name"
            class="color-story-presence-chip"
            :style="{
              background: 'var(' + row.bgVar + ')',
              borderColor: 'var(' + row.borderVar + ')',
            }"
          >{{ row.name }}</div>
        </div>
        <table class="color-story-table">
          <thead>
            <tr><th>Name</th><th>Background</th><th>Border</th></tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.name">
              <td>{{ row.name }}</td>
              <td><code>{{ row.bgVar }}</code></td>
              <td><code>{{ row.borderVar }}</code></td>
            </tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};

export const ElementPlusMapping: Story = {
  name: "Element Plus mapping",
  render: () => ({
    template: `
      <div class="color-story">
        <h2>Element Plus color mapping</h2>
        <p>
          App tokens in <code>:root</code> are canonical.
          <code>html.dark</code> maps <code>--el-color-*</code> and neutrals from them.
          <code>light-*</code> / <code>dark-*</code> derivatives stay in the mapping block only.
        </p>
        <table class="color-story-table">
          <thead>
            <tr><th>App token</th><th>Element Plus (examples)</th></tr>
          </thead>
          <tbody>
            <tr><td><code>--color-brand</code></td><td><code>--el-color-primary</code></td></tr>
            <tr><td><code>--color-brand-hover</code></td><td><code>--el-color-primary-light-3</code></td></tr>
            <tr><td><code>--color-danger</code></td><td><code>--el-color-danger</code>, <code>--el-color-error</code></td></tr>
            <tr><td><code>--color-success</code></td><td><code>--el-color-success</code></td></tr>
            <tr><td><code>--color-warning</code></td><td><code>--el-color-warning</code></td></tr>
            <tr><td><code>--color-info</code></td><td><code>--el-color-info</code></td></tr>
            <tr><td><code>--color-bg-base</code></td><td><code>--el-bg-color-page</code></td></tr>
            <tr><td><code>--color-text-muted</code></td><td><code>--el-text-color-placeholder</code></td></tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};
