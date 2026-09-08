import type { Meta, StoryObj } from "@storybook/vue3-vite";
import GalleryView from "./GalleryView.vue";
import { withWailsApp } from "../stories/wailsDecorator";

const galleryWailsBase = {
  Screenshots: () => Promise.resolve([]),
  SearchScreenshots: () => Promise.resolve([]),
  IsGalleryScanning: () => Promise.resolve(false),
  GetVRChatConfig: () =>
    Promise.resolve({
      cameraResWidth: 1920,
      cameraResHeight: 1080,
      screenshotResWidth: 1920,
      screenshotResHeight: 1080,
      pictureOutputFolder: "C:/Pictures/VRChat",
      pictureOutputSplitByDate: true,
      fpvSteadycamFov: 90,
      cacheDirectory: "",
      cacheSize: 0,
      cacheExpiryDelay: 0,
    }),
  DefaultVRChatPictureFolder: () => Promise.resolve("C:/Pictures/VRChat"),
};

const meta = {
  title: "Views/GalleryView",
  component: GalleryView,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Flat thumbnail grid with sticky day section headers. Year dividers appear when the calendar year changes; day labels omit the year (e.g. 9月4日). Filter toolbar: criteria group (world search + date range) and actions group (refresh + scan folder), height-aligned at default control size.",
      },
    },
  },
} satisfies Meta<typeof GalleryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  decorators: [withWailsApp(galleryWailsBase)],
};

export const Loading: Story = {
  decorators: [
    withWailsApp({
      ...galleryWailsBase,
      Screenshots: () => new Promise(() => {}),
    }),
  ],
};
