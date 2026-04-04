import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import CollapsibleSectionCard from "./CollapsibleSectionCard.vue";

const meta = {
  title: "Components/CollapsibleSectionCard",
  component: CollapsibleSectionCard,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
  },
  argTypes: {
    modelValue: { control: "boolean" },
    title: { control: "text" },
  },
} satisfies Meta<typeof CollapsibleSectionCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => ({
    components: { CollapsibleSectionCard },
    template: `
      <div style="max-width: 720px">
        <CollapsibleSectionCard title="セクションタイトル">
          <p style="margin: 0">
            ヘッダーをクリックで開閉します。見た目は <code>style.css</code> の
            <code>.section-card__toggle</code> / <code>.section-card--collapsed</code> です。
          </p>
        </CollapsibleSectionCard>
      </div>
    `,
  }),
};

export const InitiallyCollapsed: Story = {
  render: () => ({
    components: { CollapsibleSectionCard },
    setup() {
      const expanded = ref(false);
      return { expanded };
    },
    template: `
      <div style="max-width: 720px">
        <CollapsibleSectionCard v-model="expanded" title="初期は閉じている">
          <p style="margin: 0">開くとこの本文が表示されます。</p>
        </CollapsibleSectionCard>
      </div>
    `,
  }),
};

export const WithActivityVariantClass: Story = {
  name: "Activity layout class",
  render: () => ({
    components: { CollapsibleSectionCard },
    template: `
      <div style="max-width: 720px">
        <CollapsibleSectionCard
          class="section-card--playtime"
          title="プレイ時間（直近14日）"
        >
          <p style="margin: 0">Activity 画面と同じ <code>section-card--playtime</code> を付与した例です。</p>
        </CollapsibleSectionCard>
      </div>
    `,
  }),
};
