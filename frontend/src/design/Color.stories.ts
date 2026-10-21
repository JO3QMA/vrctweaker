import type { Meta, StoryObj } from "@storybook/vue3-vite";
import styleCss from "../assets/style.css?raw";
import {
  joinUniqueValues,
  parseStyleSheet,
  resolveCustomProperties,
  type CssCustomProperty,
  type ResolvedCustomProperty,
} from "./cssCustomProperties";
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

const parsed = parseStyleSheet(styleCss);
const rootProps = parsed.customProperties.filter(
  (prop) => prop.selector === ":root",
);
const darkProps = parsed.customProperties.filter(
  (prop) => prop.selector === "html.dark",
);

function colorGroup(prefix: string): CssCustomProperty[] {
  return rootProps.filter(
    (prop) => prop.group.startsWith(prefix) && prop.name.startsWith("--color-"),
  );
}

type SwatchRow = ResolvedCustomProperty & {
  kind: "text" | "border" | "fill";
  alias: string;
};

function legacyAliasFor(target: string): string {
  return rootProps
    .filter(
      (prop) =>
        prop.group.startsWith("Legacy") && prop.value === `var(${target})`,
    )
    .map((prop) => prop.name)
    .join(", ");
}

function swatchKind(name: string): SwatchRow["kind"] {
  if (name.includes("--color-text-")) return "text";
  if (name === "--color-border" || name.endsWith("-border")) return "border";
  return "fill";
}

function swatchRows(
  props: CssCustomProperty[],
  options?: { alias?: boolean },
): SwatchRow[] {
  return resolveCustomProperties(document, props).map((row) => ({
    ...row,
    alias: options?.alias ? legacyAliasFor(row.name) : "",
    kind: swatchKind(row.name),
  }));
}

const tokenTable = `
  <table class="color-story-table">
    <thead>
      <tr>
        <th>Group</th>
        <th>Token</th>
        <th>Usage</th>
        <th>Declared</th>
        <th>Resolved</th>
        <th>Selector</th>
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

function swatchStory(
  title: string,
  props: CssCustomProperty[],
  options?: { alias?: boolean },
): Story {
  return {
    name: title,
    render: () => ({
      setup: () => ({
        title,
        description: props[0]?.group ?? "",
        rows: swatchRows(props, options),
      }),
      template: `
        <div class="color-story">
          <h2>{{ title }}</h2>
          <p>{{ description }}</p>
          <div class="color-story-grid">
            <div v-for="row in rows" :key="row.name" class="color-story-swatch">
              <div
                v-if="row.kind === 'text'"
                class="color-story-swatch-chip color-story-swatch-chip--text"
                :style="{ color: 'var(' + row.name + ')' }"
              >Aa</div>
              <div
                v-else-if="row.kind === 'border'"
                class="color-story-swatch-chip"
                :style="{ background: 'var(--color-bg-base)', border: '3px solid var(' + row.name + ')' }"
              />
              <div
                v-else
                class="color-story-swatch-chip"
                :style="{ background: 'var(' + row.name + ')' }"
              />
              <span class="color-story-swatch-label">{{ row.name }}</span>
              <span v-if="row.description" class="color-story-swatch-label">{{ row.description }}</span>
              <span v-if="row.alias" class="color-story-swatch-label">alias: {{ row.alias }}</span>
            </div>
          </div>
          ${tokenTable}
        </div>
      `,
    }),
  };
}

export const Brand: Story = swatchStory("Brand color", colorGroup("Brand"));

export const Neutral: Story = swatchStory(
  "Neutral color",
  colorGroup("Neutral"),
  { alias: true },
);

export const Semantic: Story = swatchStory(
  "Semantic color catalog",
  colorGroup("Semantic"),
);

export const ServerStatus: Story = swatchStory(
  "Server status color",
  colorGroup("Server status"),
);

export const Presence: Story = {
  name: "Presence color",
  render: () => ({
    setup: () => {
      const rows = swatchRows(colorGroup("Presence"));
      const chips = rows.filter((row) => !row.name.endsWith("-border"));
      return { rows, chips, description: rows[0]?.group ?? "" };
    },
    template: `
      <div class="color-story">
        <h2>Presence color</h2>
        <p>{{ description }}</p>
        <p>
          チップはどの状態でも <code>--color-text-inverse</code> を使います。本番のプレゼンス変更セクション（ADR 0019 v1 の見た目方針）と同じです。きいてみてねのコントラストは WCAG AA を下回ります。状態ごとの文字色は、このトークン化とは別の後続対応です。
        </p>
        <div class="color-story-presence-row">
          <div
            v-for="row in chips"
            :key="row.name"
            class="color-story-presence-chip"
            :style="{
              background: 'var(' + row.name + ')',
              borderColor: 'var(' + row.name + '-border)',
            }"
          >{{ row.description }}</div>
        </div>
        ${tokenTable}
      </div>
    `,
  }),
};

export const LegacyAliases: Story = {
  name: "Legacy aliases",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(
        document,
        rootProps.filter((prop) => prop.group.startsWith("Legacy")),
      ),
    }),
    template: `
      <div class="color-story">
        <h2>Legacy aliases</h2>
        <p>{{ rows[0] && rows[0].group }}</p>
        ${tokenTable}
      </div>
    `,
  }),
};

/** Same eight correspondences as the previous catalog. Values are read from style.css. */
const ELEMENT_PLUS_COLOR_EXAMPLES = [
  { app: "--color-brand", el: ["--el-color-primary"] },
  { app: "--color-brand-hover", el: ["--el-color-primary-light-3"] },
  { app: "--color-danger", el: ["--el-color-danger", "--el-color-error"] },
  { app: "--color-success", el: ["--el-color-success"] },
  { app: "--color-warning", el: ["--el-color-warning"] },
  { app: "--color-info", el: ["--el-color-info"] },
  { app: "--color-bg-base", el: ["--el-bg-color-page"] },
  { app: "--color-text-muted", el: ["--el-text-color-placeholder"] },
] as const;

export const ElementPlusMapping: Story = {
  name: "Element Plus mapping",
  render: () => ({
    setup: () => ({
      rows: ELEMENT_PLUS_COLOR_EXAMPLES.map((spec) => {
        const props = spec.el.map((name) =>
          darkProps.find((prop) => prop.name === name),
        );
        const resolved = resolveCustomProperties(
          document,
          props.filter((prop): prop is CssCustomProperty => prop != null),
        );
        return {
          app: spec.app,
          examples: spec.el.join(", "),
          declared: joinUniqueValues(props.map((prop) => prop?.value ?? "")),
          resolved: joinUniqueValues(resolved.map((prop) => prop.resolved)),
        };
      }),
    }),
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
            <tr>
              <th>App token</th>
              <th>Element Plus (examples)</th>
              <th>Declared</th>
              <th>Resolved</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.app">
              <td><code>{{ row.app }}</code></td>
              <td><code>{{ row.examples }}</code></td>
              <td><code>{{ row.declared }}</code></td>
              <td><code>{{ row.resolved }}</code></td>
            </tr>
          </tbody>
        </table>
      </div>
    `,
  }),
};

export const ElementPlusVariables: Story = {
  name: "Element Plus variables",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(
        document,
        darkProps.filter((prop) => !prop.name.includes("font-size")),
      ),
    }),
    template: `
      <div class="color-story">
        <h2>Element Plus variables</h2>
        <p>
          Every custom property declared on <code>html.dark</code> except font sizes.
          The mapping story above is the eight-row correspondence.
        </p>
        ${tokenTable}
      </div>
    `,
  }),
};
