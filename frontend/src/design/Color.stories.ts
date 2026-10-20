import type { Meta, StoryObj } from "@storybook/vue3-vite";
import styleCss from "../assets/style.css?raw";
import {
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
};

function swatchRows(props: CssCustomProperty[]): SwatchRow[] {
  return resolveCustomProperties(document, props).map((row) => ({
    ...row,
    kind: row.name.includes("--color-text-")
      ? "text"
      : row.name === "--color-border" || row.name.endsWith("-border")
        ? "border"
        : "fill",
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

function swatchStory(title: string, props: CssCustomProperty[]): Story {
  return {
    name: title,
    render: () => ({
      setup: () => ({
        title,
        description: props[0]?.group ?? "",
        rows: swatchRows(props),
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
          Chips use <code>--color-text-inverse</code> on every state, matching the production
          Presence change section (ADR 0019 v1 visual policy).
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

export const ElementPlusMapping: Story = {
  name: "Element Plus mapping",
  render: () => ({
    setup: () => ({
      rows: resolveCustomProperties(
        document,
        darkProps.filter(
          (prop) =>
            !prop.name.includes("font-size") &&
            (prop.name.includes("color") ||
              prop.value.includes("var(--color-")),
        ),
      ),
    }),
    template: `
      <div class="color-story">
        <h2>Element Plus color mapping</h2>
        <p>
          App tokens in <code>:root</code> are canonical.
          Rows below are custom properties declared on <code>html.dark</code>.
          Declared <code>var(--…)</code> is shown beside the value read with
          <code>getComputedStyle</code> from the <code>html.dark</code> element.
        </p>
        ${tokenTable}
      </div>
    `,
  }),
};
