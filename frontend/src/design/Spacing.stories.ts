import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { SPACING_PATTERNS, SPACING_SCALE_PX } from "./spacingTokens";
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

export const Scale: Story = {
  name: "Spacing scale",
  render: () => ({
    setup: () => ({ scale: SPACING_SCALE_PX }),
    template: `
      <div class="spacing-story">
        <h2>Spacing scale</h2>
        <p>Numeric tokens (<code>--space-*</code>, px). Use for layout when no pattern fits.</p>
        <div v-for="px in scale" :key="px" class="spacing-story-scale-row">
          <div class="spacing-story-scale-bar" :style="{ width: px + 'px' }" />
          <span class="spacing-story-scale-label">--space-{{ px }} · {{ px }}px</span>
        </div>
      </div>
    `,
  }),
};

export const Patterns: Story = {
  name: "Spacing pattern catalog",
  render: () => ({
    setup: () => ({ patterns: SPACING_PATTERNS }),
    template: `
      <div class="spacing-story">
        <h2>Spacing pattern catalog</h2>
        <p>Prefer semantic patterns over raw scale tokens when the use case matches.</p>
        <table class="spacing-story-table">
          <thead>
            <tr>
              <th>Pattern</th>
              <th>Delegates to</th>
              <th>px</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in patterns" :key="row.varName">
              <td><code>{{ row.varName }}</code></td>
              <td><code>{{ row.scaleVar }}</code></td>
              <td>{{ row.px }}</td>
            </tr>
          </tbody>
        </table>
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
            <p style="margin: var(--space-block) 0 0">Content with block spacing below the title.</p>
          </div>
          <div class="spacing-story-layout-block">
            <strong>Form field spacing</strong>
            <div style="display: flex; flex-direction: column; gap: var(--space-form-field); margin-top: var(--space-block)">
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
