<template>
  <div class="playtime-chart-wrap">
    <canvas ref="canvasRef" @mousemove="onMove" @mouseleave="onLeave" />
    <div
      v-if="hoverIndex != null && tip"
      class="playtime-chart-tip"
      :style="{ left: tip.x + 'px', top: tip.y + 'px' }"
    >
      <div>{{ tip.date }}</div>
      <div>{{ tip.duration }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  watch,
  onMounted,
  onBeforeUnmount,
  nextTick,
  computed,
} from "vue";
import { useI18n } from "vue-i18n";
import { formatPlayDurationHMS } from "../utils/formatPlayDuration";
import {
  clampCenteredTipX,
  clampedPlotSize,
  syncCanvasBuffer,
} from "./playTimeChartGeometry";
import {
  formatPlayTimeAxisTickLabel,
  playTimeChartAxisTickUnit,
  playTimeChartMaxY,
  playTimeChartYAxisTicks,
} from "./playTimeChartScale";
import {
  playTimeChartVisibleXLabelIndices,
  playTimeChartXLabelHorizontalPad,
} from "./playTimeChartXLabels";

export interface PlayTimeDayPoint {
  date: string;
  label: string;
  seconds: number;
}

const props = defineProps<{
  series: PlayTimeDayPoint[];
}>();

const { t, locale } = useI18n();

const canvasRef = ref<HTMLCanvasElement | null>(null);
const hoverIndex = ref<number | null>(null);
let resizeObserver: ResizeObserver | null = null;

const PAD_BASE = { top: 36, bottom: 44 };
const Y_AXIS_LABEL_GAP = 6;
/** Fallback left inset when Y labels are narrow. */
const MIN_PLOT_LEFT = 48;
const X_LABEL_FONT = "12px sans-serif";
/** Approximate tip width for translateX(-50%) edge clamping. */
const TIP_WIDTH = 120;

type PlotMetrics = {
  cssW: number;
  cssH: number;
  plotW: number;
  plotH: number;
  maxY: number;
  padLeft: number;
  padRight: number;
  xLabelIndices: number[];
};

let plotCache: PlotMetrics | null = null;

function readCssVar(name: string, fallback: string): string {
  const v = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return v || fallback;
}

function formatYAxisTickSeconds(
  sec: number,
  axisUnit: ReturnType<typeof playTimeChartAxisTickUnit>,
): string {
  return formatPlayTimeAxisTickLabel(
    sec,
    {
      hour: t("chart.hour"),
      minute: t("chart.minute"),
    },
    axisUnit,
  );
}

function computePlotMetrics(
  canvas: HTMLCanvasElement,
  ctx: CanvasRenderingContext2D,
): PlotMetrics {
  const cssW = canvas.clientWidth || 300;
  const cssH = canvas.clientHeight || 280;
  const maxSeconds = props.series.reduce(
    (max, s) => Math.max(max, s.seconds),
    0,
  );
  const maxY = playTimeChartMaxY(Math.max(0, maxSeconds));
  const yAxisTicks = playTimeChartYAxisTicks(maxY);
  const yAxisUnit = playTimeChartAxisTickUnit(maxY);

  ctx.font = X_LABEL_FONT;
  const labelWidths = props.series.map((s) => ctx.measureText(s.label).width);
  const maxXLabelW = labelWidths.reduce((m, w) => Math.max(m, w), 0);
  const xHorizPad = playTimeChartXLabelHorizontalPad(maxXLabelW);

  let maxYLabelW = 0;
  for (const v of yAxisTicks) {
    maxYLabelW = Math.max(
      maxYLabelW,
      ctx.measureText(formatYAxisTickSeconds(v, yAxisUnit)).width,
    );
  }

  const padLeft = Math.max(
    MIN_PLOT_LEFT,
    maxYLabelW + Y_AXIS_LABEL_GAP,
    xHorizPad,
  );
  const padRight = xHorizPad;
  const { plotW, plotH } = clampedPlotSize(cssW, cssH, {
    top: PAD_BASE.top,
    left: padLeft,
    right: padRight,
    bottom: PAD_BASE.bottom,
  });

  const xLabelIndices =
    props.series.length === 0
      ? []
      : playTimeChartVisibleXLabelIndices({
          pointCount: props.series.length,
          plotLeft: padLeft,
          plotWidth: plotW,
          labelWidths,
        });

  return {
    cssW,
    cssH,
    plotW,
    plotH,
    maxY,
    padLeft,
    padRight,
    xLabelIndices,
  };
}

/** Measure plot geometry without touching the canvas buffer. */
function measurePlot(canvas: HTMLCanvasElement): PlotMetrics {
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    const cssW = canvas.clientWidth || 300;
    const cssH = canvas.clientHeight || 280;
    const maxY = playTimeChartMaxY(0);
    const padLeft = MIN_PLOT_LEFT;
    const padRight = playTimeChartXLabelHorizontalPad(36);
    const { plotW, plotH } = clampedPlotSize(cssW, cssH, {
      top: PAD_BASE.top,
      left: padLeft,
      right: padRight,
      bottom: PAD_BASE.bottom,
    });
    return {
      cssW,
      cssH,
      plotW,
      plotH,
      maxY,
      padLeft,
      padRight,
      xLabelIndices: [],
    };
  }
  return computePlotMetrics(canvas, ctx);
}

function refreshPlot(
  canvas: HTMLCanvasElement,
  ctx: CanvasRenderingContext2D,
): PlotMetrics {
  plotCache = computePlotMetrics(canvas, ctx);
  return plotCache;
}

function pointAt(
  i: number,
  n: number,
  padLeft: number,
  plotW: number,
  plotH: number,
  maxY: number,
  seconds: number,
) {
  const x = padLeft + (n <= 1 ? plotW / 2 : (i / (n - 1)) * plotW);
  const y = PAD_BASE.top + plotH - (seconds / maxY) * plotH;
  return { x, y };
}

function draw(): void {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  const L = refreshPlot(canvas, ctx);
  const dpr = window.devicePixelRatio || 1;
  syncCanvasBuffer(canvas, L.cssW, L.cssH, dpr);
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  const { cssW, cssH, plotW, plotH, maxY, padLeft, xLabelIndices } = L;
  const textMuted = readCssVar("--color-text-secondary", "#a0a0a0");
  const border = readCssVar("--color-border", "#333333");
  const accent = readCssVar("--color-brand", "#5b9bd5");

  ctx.clearRect(0, 0, cssW, cssH);
  if (props.series.length === 0) return;

  const yAxisTicks = playTimeChartYAxisTicks(maxY);
  const yAxisUnit = playTimeChartAxisTickUnit(maxY);
  ctx.strokeStyle = border;
  ctx.fillStyle = textMuted;
  ctx.lineWidth = 1;
  ctx.font = X_LABEL_FONT;
  ctx.textAlign = "right";
  ctx.textBaseline = "middle";
  for (const v of yAxisTicks) {
    const y = PAD_BASE.top + plotH - (v / maxY) * plotH;
    ctx.beginPath();
    ctx.moveTo(padLeft, y);
    ctx.lineTo(padLeft + plotW, y);
    ctx.stroke();
    ctx.fillText(formatYAxisTickSeconds(v, yAxisUnit), padLeft - 6, y);
  }

  const n = props.series.length;
  ctx.beginPath();
  props.series.forEach((s, i) => {
    const { x, y } = pointAt(i, n, padLeft, plotW, plotH, maxY, s.seconds);
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.strokeStyle = accent;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = accent;
  props.series.forEach((s, i) => {
    const { x, y } = pointAt(i, n, padLeft, plotW, plotH, maxY, s.seconds);
    ctx.beginPath();
    ctx.arc(x, y, hoverIndex.value === i ? 5 : 3.5, 0, Math.PI * 2);
    ctx.fill();
  });

  ctx.fillStyle = textMuted;
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  const xLabelY = PAD_BASE.top + plotH + 8;
  for (const i of xLabelIndices) {
    const s = props.series[i];
    if (!s) continue;
    const { x } = pointAt(i, n, padLeft, plotW, plotH, maxY, s.seconds);
    ctx.fillText(s.label, x, xLabelY);
  }
}

const tip = computed(() => {
  const i = hoverIndex.value;
  const canvas = canvasRef.value;
  if (i == null || !props.series[i] || !canvas) return null;
  const L = plotCache ?? measurePlot(canvas);
  const p = props.series[i];
  const { x, y } = pointAt(
    i,
    props.series.length,
    L.padLeft,
    L.plotW,
    L.plotH,
    L.maxY,
    p.seconds,
  );
  return {
    x: clampCenteredTipX(x, L.cssW, TIP_WIDTH, L.padLeft, L.padRight),
    y: Math.max(8, y - 48),
    date: p.date,
    duration: formatPlayDurationHMS(p.seconds, {
      hour: t("chart.hour"),
      minute: t("chart.minute"),
      second: t("chart.second"),
    }),
  };
});

function onMove(ev: MouseEvent): void {
  const canvas = canvasRef.value;
  if (!canvas || props.series.length === 0) {
    hoverIndex.value = null;
    return;
  }
  const ctx = canvas.getContext("2d");
  const L = plotCache ?? (ctx ? refreshPlot(canvas, ctx) : measurePlot(canvas));
  const rect = canvas.getBoundingClientRect();
  const mx = ev.clientX - rect.left;
  let best = 0;
  let bestDist = Infinity;
  props.series.forEach((s, i) => {
    const { x } = pointAt(
      i,
      props.series.length,
      L.padLeft,
      L.plotW,
      L.plotH,
      L.maxY,
      s.seconds,
    );
    const d = Math.abs(x - mx);
    if (d < bestDist) {
      bestDist = d;
      best = i;
    }
  });
  const next = bestDist < 40 ? best : null;
  if (hoverIndex.value === next) return;
  hoverIndex.value = next;
  draw();
}

function onLeave(): void {
  if (hoverIndex.value == null) return;
  hoverIndex.value = null;
  draw();
}

watch(
  () => [props.series, locale.value] as const,
  () => {
    void nextTick(() => draw());
  },
  { deep: true },
);

onMounted(() => {
  void nextTick(() => draw());
  const canvas = canvasRef.value;
  if (canvas && typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(() => draw());
    resizeObserver.observe(canvas);
  }
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  resizeObserver = null;
});
</script>

<style scoped>
.playtime-chart-wrap {
  position: relative;
  width: 100%;
  height: 280px;
}

.playtime-chart-wrap canvas {
  width: 100%;
  height: 100%;
  display: block;
}

.playtime-chart-tip {
  position: absolute;
  transform: translateX(-50%);
  pointer-events: none;
  padding: 6px 8px;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  /* Preserve 1.4: Typography scale has 1.25/1.5/1.75 only; tip offset assumes ~1.4 */
  line-height: 1.4;
  white-space: nowrap;
  z-index: 1;
}
</style>
