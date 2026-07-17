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
import { clampCenteredTipX, syncCanvasBuffer } from "./playTimeChartGeometry";

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

const PAD = { top: 36, left: 48, right: 14, bottom: 44 };
/** Approximate tip width for translateX(-50%) edge clamping. */
const TIP_WIDTH = 120;

type PlotMetrics = {
  cssW: number;
  cssH: number;
  plotW: number;
  plotH: number;
  maxY: number;
};

let plotCache: PlotMetrics | null = null;

function readCssVar(name: string, fallback: string): string {
  const v = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return v || fallback;
}

function formatYAxisTickSeconds(sec: number): string {
  if (sec >= 3600) {
    return `${Math.floor(sec / 3600)}${t("chart.hour")}`;
  }
  if (sec >= 60) {
    return `${Math.floor(sec / 60)}${t("chart.minute")}`;
  }
  return `${sec}${t("chart.second")}`;
}

function niceMax(seconds: number): number {
  if (seconds <= 0) return 60;
  const raw = seconds * 1.1;
  const step = raw <= 60 ? 10 : raw <= 3600 ? 300 : 1800;
  return Math.ceil(raw / step) * step;
}

/** Measure plot geometry without touching the canvas buffer. */
function measurePlot(canvas: HTMLCanvasElement): PlotMetrics {
  const cssW = canvas.clientWidth || 300;
  const cssH = canvas.clientHeight || 280;
  const plotW = cssW - PAD.left - PAD.right;
  const plotH = cssH - PAD.top - PAD.bottom;
  const maxY = niceMax(Math.max(0, ...props.series.map((s) => s.seconds)));
  return { cssW, cssH, plotW, plotH, maxY };
}

function refreshPlot(canvas: HTMLCanvasElement): PlotMetrics {
  plotCache = measurePlot(canvas);
  return plotCache;
}

function pointAt(
  i: number,
  n: number,
  plotW: number,
  plotH: number,
  maxY: number,
  seconds: number,
) {
  const x = PAD.left + (n <= 1 ? plotW / 2 : (i / (n - 1)) * plotW);
  const y = PAD.top + plotH - (seconds / maxY) * plotH;
  return { x, y };
}

function draw(): void {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const L = refreshPlot(canvas);
  const dpr = window.devicePixelRatio || 1;
  syncCanvasBuffer(canvas, L.cssW, L.cssH, dpr);
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

  const { cssW, cssH, plotW, plotH, maxY } = L;
  const textMuted = readCssVar("--text-secondary", "#a0a0a0");
  const border = readCssVar("--border", "#333333");
  const accent = readCssVar("--accent", "#5b9bd5");

  ctx.clearRect(0, 0, cssW, cssH);
  if (props.series.length === 0) return;

  const ticks = 4;
  ctx.strokeStyle = border;
  ctx.fillStyle = textMuted;
  ctx.lineWidth = 1;
  ctx.font = "12px sans-serif";
  ctx.textAlign = "right";
  ctx.textBaseline = "middle";
  for (let i = 0; i <= ticks; i++) {
    const v = (maxY * i) / ticks;
    const y = PAD.top + plotH - (v / maxY) * plotH;
    ctx.beginPath();
    ctx.moveTo(PAD.left, y);
    ctx.lineTo(PAD.left + plotW, y);
    ctx.stroke();
    ctx.fillText(formatYAxisTickSeconds(v), PAD.left - 6, y);
  }

  const n = props.series.length;
  ctx.beginPath();
  props.series.forEach((s, i) => {
    const { x, y } = pointAt(i, n, plotW, plotH, maxY, s.seconds);
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.strokeStyle = accent;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = accent;
  props.series.forEach((s, i) => {
    const { x, y } = pointAt(i, n, plotW, plotH, maxY, s.seconds);
    ctx.beginPath();
    ctx.arc(x, y, hoverIndex.value === i ? 5 : 3.5, 0, Math.PI * 2);
    ctx.fill();
  });

  ctx.fillStyle = textMuted;
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  props.series.forEach((s, i) => {
    const { x } = pointAt(i, n, plotW, plotH, maxY, s.seconds);
    ctx.fillText(s.label, x, PAD.top + plotH + 8);
  });
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
    L.plotW,
    L.plotH,
    L.maxY,
    p.seconds,
  );
  return {
    x: clampCenteredTipX(x, L.cssW, TIP_WIDTH, PAD.left, PAD.right),
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
  // Use cached metrics; never resize the buffer on mousemove.
  const L = plotCache ?? refreshPlot(canvas);
  const rect = canvas.getBoundingClientRect();
  const mx = ev.clientX - rect.left;
  let best = 0;
  let bestDist = Infinity;
  props.series.forEach((s, i) => {
    const { x } = pointAt(
      i,
      props.series.length,
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
  border: 1px solid var(--border, #333);
  background: var(--bg-tertiary, #1a1a1a);
  color: var(--text-primary, #e0e0e0);
  font-size: 12px;
  line-height: 1.4;
  white-space: nowrap;
  z-index: 1;
}
</style>
