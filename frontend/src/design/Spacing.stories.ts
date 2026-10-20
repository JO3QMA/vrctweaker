import type { Meta, StoryObj } from "@storybook/vue3-vite";
import styleCss from "../assets/style.css?raw";
import {
  parseStyleSheet,
  resolveCustomProperties,
} from "./cssCustomProperties";
import "./Spacing.stories.css";

const catalogParameters = {
  controls: { disable: true },
};

const meta = {
  title: "Design System/Spacing",
  tags: ["autodocs"],
  parameters: catalogParameters,
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

const spaceProps = parseStyleSheet(styleCss).customProperties.filter(
  (prop) => prop.selector === ":root" && prop.name.startsWith("--space-"),
);
const scaleProps = spaceProps.filter((prop) => /^--space-\d+$/.test(prop.name));
const patternProps = spaceProps.filter(
  (prop) => !/^--space-\d+$/.test(prop.name),
);

const tokenTable = `
  <table class="spacing-story-table">
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
      <tr v-for="row in rows" :key="row.name">
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

export const Scale: Story = {
  name: "Spacing scale",
  render: () => ({
    setup: () => ({ rows: resolveCustomProperties(document, scaleProps) }),
    template: `
      <div class="spacing-story">
        <h2>Spacing scale</h2>
        <p>{{ rows[0] && rows[0].group }}. Numeric tokens. Use when no pattern fits.</p>
        <div v-for="row in rows" :key="row.name" class="spacing-story-scale-row">
          <div class="spacing-story-scale-bar" :style="{ width: 'var(' + row.name + ')' }" />
          <span class="spacing-story-scale-label">{{ row.name }} · {{ row.value }}</span>
        </div>
        ${tokenTable}
      </div>
    `,
  }),
};

export const Patterns: Story = {
  name: "Spacing pattern catalog",
  render: () => ({
    setup: () => ({ rows: resolveCustomProperties(document, patternProps) }),
    template: `
      <div class="spacing-story">
        <h2>Spacing pattern catalog</h2>
        <p>{{ rows[0] && rows[0].group }}. Prefer a pattern over a raw scale step when the use case matches.</p>
        ${tokenTable}
      </div>
    `,
  }),
};

export const LayoutExample: Story = {
  name: "Layout example",
  render: () => ({
    template: `
      <div class="spacing-story">
        <h2>Layout example</h2>
        <p>
          <code>--space-page</code> padding,
          <code>--space-section</code> between blocks,
          <code>--space-block</code> inside blocks,
          <code>--space-action-group</code> between actions.
        </p>
        <div class="spacing-story-layout-demo">
          <div class="spacing-story-layout-block">
            <strong>Section block</strong>
            <p class="spacing-story-layout-block-text">Content with block spacing below the title.</p>
          </div>
          <div class="spacing-story-layout-block">
            <strong>Form field spacing</strong>
            <div class="spacing-story-field-group">
              <div class="spacing-story-action-chip">Field A</div>
              <div class="spacing-story-action-chip">Field B</div>
            </div>
          </div>
          <div class="spacing-story-action-group">
            <span class="spacing-story-action-chip">Secondary</span>
            <span class="spacing-story-action-chip">Primary</span>
          </div>
        </div>
      </div>
    `,
  }),
};
