import type { Meta, StoryObj } from "@storybook/vue3-vite";
import styleCss from "../assets/style.css?raw";
import {
  parseStyleSheet,
  readUsedFont,
  resolveCustomProperties,
  type CssCustomProperty,
  type ResolvedCustomProperty,
} from "./cssCustomProperties";
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

const parsed = parseStyleSheet(styleCss);
const rootProps = parsed.customProperties.filter(
  (prop) => prop.selector === ":root",
);
const darkProps = parsed.customProperties.filter(
  (prop) => prop.selector === "html.dark",
);

function rootNamed(prefix: string): CssCustomProperty[] {
  return rootProps.filter((prop) => prop.name.startsWith(prefix));
}

const epFontSizeOrder = [
  "extra-large",
  "large",
  "medium",
  "base",
  "extra-small",
  "small",
];

function epFontSizeRank(name: string): number {
  const key = name.slice("--el-font-size-".length);
  const index = epFontSizeOrder.indexOf(key);
  return index === -1 ? epFontSizeOrder.length : index;
}

const tokenTable = `
  <table class="typography-story-table">
    <thead>
      <tr>
        <th scope="col">Group</th>
        <th scope="col">Token</th>
        <th scope="col">Usage</th>
        <th scope="col">Declared</th>
        <th scope="col">Resolved</th>
        <th scope="col">Selector</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.selector + row.name">
        <td>{{ row.group }}</td>
        <td><code>{{ row.name }}</code></td>
        <td>{{ row.description }}</td>
        <td><code>{{ row.value }}</code></td>
        <td><code v-if="row.referencesVar">{{ row.resolved }}</code><span v-else>—</span></td>
        <td><code>{{ row.selector }}</code></td>
      </tr>
    </tbody>
  </table>
`;

function tokenStory(
  title: string,
  intro: string,
  props: CssCustomProperty[],
): Story {
  return {
    name: title,
    render: () => ({
      setup: () => ({
        title,
        intro,
        rows: resolveCustomProperties(document, props),
      }),
      template: `
        <div class="typography-story">
          <h2>{{ title }}</h2>
          <p>{{ intro }}</p>
          ${tokenTable}
        </div>
      `,
    }),
  };
}

const fontSizeScale = rootNamed("--font-size-").filter((prop) =>
  /^--font-size-\d+$/.test(prop.name),
);
const fontSizeDerivatives = rootNamed("--font-size-").filter(
  (prop) => !/^--font-size-\d+$/.test(prop.name),
);

export const FontSizeScale: Story = {
  name: "Font size scale",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(document, fontSizeScale),
    }),
    template: `
      <div class="typography-story">
        <h2>Font size scale</h2>
        <p>Numeric tokens (<code>--font-size-*</code>, px). Body default is 14.</p>
        <div v-for="row in rows" :key="row.name" class="typography-story-scale-row">
          <span v-if="row.description" class="typography-story-scale-label">{{ row.name }} · {{ row.resolved || row.value }} · {{ row.description }}</span>
          <span v-else class="typography-story-scale-label">{{ row.name }} · {{ row.resolved || row.value }}</span>
          <span :style="{ fontSize: 'var(' + row.name + ')' }">The quick brown fox</span>
        </div>
        ${tokenTable}
      </div>
    `,
  }),
};

export const FontSizeDerivatives: Story = tokenStory(
  "Font size derivatives",
  "Sizes outside the numeric scale. Declared calc()/var() is listed with the used value from getComputedStyle.",
  fontSizeDerivatives,
);

export const LineHeightScale: Story = tokenStory(
  "Line height scale",
  "Unitless ratios. Body default is --line-height-normal (1.5).",
  rootNamed("--line-height-"),
);

export const FontFamilyUi: Story = {
  name: "UI font family",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(document, rootNamed("--font-family-")),
    }),
    template: `
      <div class="typography-story">
        <h2>UI font family</h2>
        <p><code>--font-family-ui</code> — sans-serif stack for body, headings, and labels.</p>
        <p :style="{ fontFamily: 'var(--font-family-ui)' }">
          The quick brown fox jumps over the lazy dog.
        </p>
        ${tokenTable}
      </div>
    `,
  }),
};

export const FontWeightScale: Story = {
  name: "Font weight scale",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(document, rootNamed("--font-weight-")),
    }),
    template: `
      <div class="typography-story">
        <h2>Font weight scale</h2>
        <p>Body default is <code>--font-weight-400</code>.</p>
        <div v-for="row in rows" :key="row.name" class="typography-story-scale-row">
          <span v-if="row.description" class="typography-story-scale-label">{{ row.name }} · {{ row.resolved || row.value }} · {{ row.description }}</span>
          <span v-else class="typography-story-scale-label">{{ row.name }} · {{ row.resolved || row.value }}</span>
          <span :style="{ fontWeight: 'var(' + row.name + ')' }">Typography weight {{ row.value }}</span>
        </div>
        ${tokenTable}
      </div>
    `,
  }),
};

type TextStyleRow = {
  selector: string;
  className: string;
  group: string;
  fontSize: string;
  lineHeight: string;
  fontWeight: string;
  resolvedSize: string;
  resolvedLineHeight: string;
  resolvedWeight: string;
};

function declarationValue(selector: string, property: string): string {
  return (
    parsed.declarations.find(
      (decl) => decl.selector === selector && decl.property === property,
    )?.value ?? ""
  );
}

function textStyleRows(): TextStyleRow[] {
  const selectors: string[] = [];
  for (const decl of parsed.declarations) {
    if (!/^\.(?:text-[a-z0-9-]+|page-title)$/.test(decl.selector)) continue;
    if (!selectors.includes(decl.selector)) selectors.push(decl.selector);
  }
  return selectors.map((selector) => {
    const className = selector.slice(1);
    const used = readUsedFont(document, className);
    const group =
      parsed.declarations.find((decl) => decl.selector === selector)?.group ??
      "";
    return {
      selector,
      className,
      group,
      fontSize: declarationValue(selector, "font-size"),
      lineHeight: declarationValue(selector, "line-height"),
      fontWeight: declarationValue(selector, "font-weight"),
      resolvedSize: used.fontSize,
      resolvedLineHeight: used.lineHeight,
      resolvedWeight: used.fontWeight,
    };
  });
}

export const TextStyleCatalog: Story = {
  name: "Text style catalog",
  render: () => ({
    setup: () => {
      const styles = textStyleRows();
      return {
        pageTitle: styles.find((row) => row.className === "page-title"),
        styles: styles.filter((row) => row.className !== "page-title"),
      };
    },
    template: `
      <div class="typography-story">
        <h2>Text style catalog</h2>
        <p>Shared classes in <code>style.css</code>. Colors are not included (use Color tokens).</p>
        <table class="typography-story-table">
          <caption>Text style catalog</caption>
          <thead>
            <tr>
              <th scope="col">Class</th>
              <th scope="col">Group</th>
              <th scope="col">font-size</th>
              <th scope="col">Resolved size</th>
              <th scope="col">line-height</th>
              <th scope="col">Resolved line-height</th>
              <th scope="col">font-weight</th>
              <th scope="col">Resolved weight</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in styles" :key="row.selector">
              <td><code>{{ row.selector }}</code></td>
              <td>{{ row.group }}</td>
              <td><code>{{ row.fontSize || "—" }}</code></td>
              <td><code v-if="row.fontSize.includes('var(')">{{ row.resolvedSize }}</code><span v-else>—</span></td>
              <td><code>{{ row.lineHeight || "—" }}</code></td>
              <td><code v-if="row.lineHeight.includes('var(')">{{ row.resolvedLineHeight }}</code><span v-else>—</span></td>
              <td><code>{{ row.fontWeight || "—" }}</code></td>
              <td><code v-if="row.fontWeight.includes('var(')">{{ row.resolvedWeight }}</code><span v-else>—</span></td>
            </tr>
          </tbody>
        </table>
        <div class="typography-story-sample">
          <div v-if="pageTitle" class="typography-story-sample-row">
            <span class="typography-story-sample-label">.page-title (H1 size alias; line-height inherits body)</span>
            <h1 class="page-title">Page title sample</h1>
          </div>
          <div v-for="row in styles" :key="'sample-' + row.selector" class="typography-story-sample-row">
            <span class="typography-story-sample-label">{{ row.selector }}</span>
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
      rows: resolveCustomProperties(
        document,
        darkProps.filter((prop) => prop.name.startsWith("--el-font-size-")),
      )
        .slice()
        .sort(
          (a, b) => epFontSizeRank(a.name) - epFontSizeRank(b.name),
        ) satisfies ResolvedCustomProperty[],
    }),
    template: `
      <div class="typography-story">
        <h2>Element Plus typography mapping</h2>
        <p>
          Declared on <code>html.dark</code>.
          <code>var(--…)</code> rows also show the used value from
          <code>getComputedStyle(document.querySelector("html.dark"))</code>.
        </p>
        ${tokenTable}
        <div class="typography-story-ep-samples">
          <p
            v-for="row in rows"
            :key="'sample-' + row.name"
            class="typography-story-ep-sample"
            :style="{ fontSize: 'var(' + row.name + ')' }"
          >
            {{ row.name }} sample
          </p>
        </div>
      </div>
    `,
  }),
};
