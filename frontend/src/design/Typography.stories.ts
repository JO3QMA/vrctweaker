import type { Meta, StoryObj } from "@storybook/vue3-vite";
import {
  ELEMENT_PLUS_TYPOGRAPHY_DERIVATIVES,
  ELEMENT_PLUS_TYPOGRAPHY_MAPPING,
  FONT_FAMILY_UI_VAR,
  FONT_SIZE_DERIVATIVES,
  FONT_SIZE_SCALE_PX,
  FONT_WEIGHT_SCALE,
  LINE_HEIGHT_PATTERNS,
  TEXT_STYLES,
} from "./typographyTokens";
import "./Typography.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Design System/Typography",
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const FontSizeScale: Story = {
  name: "Font size scale",
  render: () => ({
    setup: () => ({ scale: FONT_SIZE_SCALE_PX }),
    template: `
      <div class="typography-story">
        <h2>Font size scale</h2>
        <p>Numeric tokens (<code>--font-size-*</code>, px). Body default is 14.</p>
        <div v-for="px in scale" :key="px" class="typography-story-scale-row">
          <span class="typography-story-scale-label">--font-size-{{ px }} · {{ px }}px</span>
          <span :style="{ fontSize: px + 'px' }">The quick brown fox</span>
        </div>
      </div>
    `,
  }),
};

export const FontSizeDerivatives: Story = {
  name: "Font size derivatives",
  render: () => ({
    setup: () => ({ derivatives: FONT_SIZE_DERIVATIVES }),
    template: `
      <div class="typography-story">
        <h2>Font size derivatives</h2>
        <p>Sizes outside the numeric scale that preserve legacy computed values.</p>
        <table class="typography-story-table">
          <caption>Font size derivatives</caption>
          <thead>
            <tr>
              <th scope="col">Token</th>
              <th scope="col">Value</th>
              <th scope="col">Note</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in derivatives" :key="row.varName">
              <td><code>{{ row.varName }}</code></td>
              <td><code>{{ row.cssValue }}</code></td>
              <td>{{ row.note }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};

export const LineHeightScale: Story = {
  name: "Line height scale",
  render: () => ({
    setup: () => ({ patterns: LINE_HEIGHT_PATTERNS }),
    template: `
      <div class="typography-story">
        <h2>Line height scale</h2>
        <p>Unitless ratios. Body default is <code>--line-height-normal</code> (1.5).</p>
        <table class="typography-story-table">
          <caption>Line height scale</caption>
          <thead>
            <tr>
              <th scope="col">Token</th>
              <th scope="col">Value</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in patterns" :key="row.varName">
              <td><code>{{ row.varName }}</code></td>
              <td>{{ row.value }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};

export const FontFamilyUi: Story = {
  name: "UI font family",
  render: () => ({
    setup: () => ({ familyVar: FONT_FAMILY_UI_VAR }),
    template: `
      <div class="typography-story">
        <h2>UI font family</h2>
        <p><code>{{ familyVar }}</code> — sans-serif stack for body, headings, and labels.</p>
        <p :style="{ fontFamily: 'var(' + familyVar + ')' }">
          The quick brown fox jumps over the lazy dog.
        </p>
      </div>
    `,
  }),
};

export const FontWeightScale: Story = {
  name: "Font weight scale",
  render: () => ({
    setup: () => ({ weights: FONT_WEIGHT_SCALE }),
    template: `
      <div class="typography-story">
        <h2>Font weight scale</h2>
        <p>Body default is <code>--font-weight-400</code>.</p>
        <div v-for="w in weights" :key="w" class="typography-story-scale-row">
          <span class="typography-story-scale-label">--font-weight-{{ w }}</span>
          <span :style="{ fontWeight: w }">Typography weight {{ w }}</span>
        </div>
      </div>
    `,
  }),
};

export const TextStyleCatalog: Story = {
  name: "Text style catalog",
  render: () => ({
    setup: () => ({ styles: TEXT_STYLES }),
    template: `
      <div class="typography-story">
        <h2>Text style catalog</h2>
        <p>Shared classes in <code>style.css</code>. Colors are not included (use Color tokens).</p>
        <table class="typography-story-table">
          <caption>Text style catalog</caption>
          <thead>
            <tr>
              <th scope="col">Class</th>
              <th scope="col">font-size</th>
              <th scope="col">line-height</th>
              <th scope="col">font-weight</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in styles" :key="row.className">
              <td><code>.{{ row.className }}</code></td>
              <td><code>{{ row.fontSize }}</code></td>
              <td><code>{{ row.lineHeightVar }}</code></td>
              <td><code>{{ row.fontWeightVar }}</code></td>
            </tr>
          </tbody>
        </table>
        <div class="typography-story-sample">
          <div class="typography-story-sample-row">
            <span class="typography-story-sample-label">.page-title (H1 size alias; line-height inherits body)</span>
            <h1 class="page-title">Page title sample</h1>
          </div>
          <div v-for="row in styles" :key="'sample-' + row.className" class="typography-story-sample-row">
            <span class="typography-story-sample-label">.{{ row.className }}</span>
            <p :class="row.className">The quick brown fox jumps over the lazy dog.</p>
          </div>
        </div>
      </div>
    `,
  }),
};

export const ElementPlusMapping: Story = {
  name: "Element Plus typography mapping",
  render: () => ({
    setup: () => ({
      mapping: ELEMENT_PLUS_TYPOGRAPHY_MAPPING,
      derivatives: ELEMENT_PLUS_TYPOGRAPHY_DERIVATIVES,
    }),
    template: `
      <div class="typography-story">
        <h2>Element Plus typography mapping</h2>
        <p>
          Defined in <code>html.dark</code> (enabled globally in Storybook preview);
          the samples below resolve the mapped Element Plus variables.
        </p>
        <table class="typography-story-table">
          <caption>Element Plus typography mapping</caption>
          <thead>
            <tr>
              <th scope="col">EP variable</th>
              <th scope="col">Delegates to</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in mapping" :key="row.elVar">
              <td><code>{{ row.elVar }}</code></td>
              <td><code>var({{ row.appVar }})</code></td>
            </tr>
            <tr v-for="row in derivatives" :key="row.elVar">
              <td><code>{{ row.elVar }}</code></td>
              <td><code>{{ row.valuePx }}px</code> (derivative, outside scale)</td>
            </tr>
          </tbody>
        </table>
        <div class="typography-story-ep-samples">
          <p
            v-for="row in mapping"
            :key="'sample-' + row.elVar"
            class="typography-story-ep-sample"
            :style="{ fontSize: 'var(' + row.elVar + ')' }"
          >
            {{ row.elVar }} sample
          </p>
          <p
            v-for="row in derivatives"
            :key="'sample-' + row.elVar"
            class="typography-story-ep-sample"
            :style="{ fontSize: 'var(' + row.elVar + ')' }"
          >
            {{ row.elVar }} sample
          </p>
        </div>
      </div>
    `,
  }),
};
