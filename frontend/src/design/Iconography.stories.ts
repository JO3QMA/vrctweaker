import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { CaretRight, Search } from "@element-plus/icons-vue";
import VtIcon from "../components/VtIcon.vue";
import SampleIcon from "../icons/SampleIcon.vue";
import {
  ICON_SIZE_LEGACY,
  ICON_SIZE_PATTERNS,
  VT_ICON_SIZES,
} from "./iconTokens";
import "./Iconography.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Design System/Iconography",
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const Scale: Story = {
  name: "Icon size scale",
  render: () => ({
    components: { VtIcon, Search },
    setup: () => ({ patterns: ICON_SIZE_PATTERNS }),
    template: `
      <div class="iconography-story">
        <h2>Icon size scale</h2>
        <p>Square slot via <code>font-size</code> on <code>VtIcon</code> / <code>el-icon</code>.</p>
        <div v-for="pattern in patterns" :key="pattern.px" class="iconography-story-size-row">
          <VtIcon :size="pattern.name">
            <Search />
          </VtIcon>
          <span>{{ pattern.scaleVar }} · {{ pattern.px }}px</span>
        </div>
      </div>
    `,
  }),
};

export const Patterns: Story = {
  name: "Icon size pattern catalog",
  render: () => ({
    setup: () => ({ patterns: ICON_SIZE_PATTERNS, legacy: ICON_SIZE_LEGACY }),
    template: `
      <div class="iconography-story">
        <h2>Icon size pattern catalog</h2>
        <table class="iconography-story-table">
          <thead>
            <tr>
              <th>Pattern</th>
              <th>Delegates to</th>
              <th>px</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in patterns" :key="row.varName">
              <td><code>{{ row.varName }}</code> (VtIcon <code>{{ row.name }}</code>)</td>
              <td><code>{{ row.scaleVar }}</code></td>
              <td>{{ row.px }}</td>
            </tr>
            <tr>
              <td><code>{{ legacy.toggle.varName }}</code> (legacy)</td>
              <td><code>{{ legacy.toggle.delegatesTo }}</code></td>
              <td>{{ legacy.toggle.px }} → adoption: <code>{{ legacy.toggle.adoptionTarget }}</code></td>
            </tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};

export const VtIconSizes: Story = {
  name: "VtIcon sizes",
  render: () => ({
    components: { VtIcon, Search, SampleIcon },
    setup: () => ({ sizes: VT_ICON_SIZES }),
    template: `
      <div class="iconography-story">
        <h2>VtIcon</h2>
        <p><code>size</code> is required. Decorative icons default to <code>aria-hidden="true"</code>.</p>
        <div class="iconography-story-examples">
          <div v-for="size in sizes" :key="size" class="iconography-story-example-block">
            <h3>{{ size }}</h3>
            <div class="iconography-story-inline">
              <VtIcon :size="size"><Search /></VtIcon>
              <span>Element Plus Search</span>
            </div>
            <div class="iconography-story-inline">
              <VtIcon :size="size"><SampleIcon /></VtIcon>
              <span>Custom SampleIcon</span>
            </div>
          </div>
        </div>
      </div>
    `,
  }),
};

export const ColorRule: Story = {
  name: "Icon color rule",
  render: () => ({
    components: { VtIcon, CaretRight },
    template: `
      <div class="iconography-story">
        <h2>Icon color rule</h2>
        <p>Icons use <code>currentColor</code>. Parent sets Neutral text tokens.</p>
        <div class="iconography-story-color-swatch iconography-story-color--secondary">
          <VtIcon size="default"><CaretRight /></VtIcon>
          <span>secondary (decorative)</span>
        </div>
        <div class="iconography-story-color-swatch iconography-story-color--primary">
          <VtIcon size="default"><CaretRight /></VtIcon>
          <span>primary (hover / active parent)</span>
        </div>
        <div class="iconography-story-color-swatch iconography-story-color--muted">
          <VtIcon size="default"><CaretRight /></VtIcon>
          <span>muted (disabled context)</span>
        </div>
      </div>
    `,
  }),
};

export const NavigationGlyph: Story = {
  name: "Navigation glyph (emoji)",
  render: () => ({
    template: `
      <div class="iconography-story">
        <h2>Navigation glyph</h2>
        <p>Emoji keep intrinsic colors. Align size with <code>.nav-glyph-size-default</code> until SVG migration.</p>
        <div class="iconography-story-nav-row">
          <span class="nav-glyph-size-legacy" aria-hidden="true">🏠</span>
          <span>Dashboard (legacy 14px — current sidebar)</span>
        </div>
        <div class="iconography-story-nav-row">
          <span class="nav-glyph-size-default" aria-hidden="true">🏠</span>
          <span>Dashboard (default 16px — adoption target)</span>
        </div>
        <div class="iconography-story-nav-row">
          <span class="nav-glyph-size-compact" aria-hidden="true">📊</span>
          <span>Activity (compact)</span>
        </div>
      </div>
    `,
  }),
};
