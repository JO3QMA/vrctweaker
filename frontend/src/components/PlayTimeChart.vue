<template>
  <div class="playtime-chart-wrap">
    <canvas ref="canvasRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from "vue";
import Chart from "chart.js/auto";
import type { Chart as ChartType } from "chart.js";
import { formatPlayDurationHMS } from "../utils/formatPlayDuration";

export interface PlayTimeDayPoint {
  date: string;
  label: string;
  seconds: number;
}

const props = defineProps<{
  series: PlayTimeDayPoint[];
}>();

const canvasRef = ref<HTMLCanvasElement | null>(null);
let chart: ChartType<"line"> | null = null;

function readCssVar(name: string, fallback: string): string {
  const v = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return v || fallback;
}

function formatYAxisTickSeconds(sec: number): string {
  if (sec >= 3600) {
    return `${Math.floor(sec / 3600)}時間`;
  }
  if (sec >= 60) {
    return `${Math.floor(sec / 60)}分`;
  }
  return `${sec}秒`;
}

function buildChartOptions() {
  const textMuted = readCssVar("--text-secondary", "#a0a0a0");
  const border = readCssVar("--border", "#333333");
  const bgTertiary = readCssVar("--bg-tertiary", "#1a1a1a");

  return {
    responsive: true,
    maintainAspectRatio: false,
    /* 上下: 軸ラベル・ポイントが canvas 端で欠けないよう余白を確保 */
    layout: {
      padding: {
        top: 36,
        left: 10,
        right: 14,
        bottom: 44,
      },
    },
    interaction: {
      mode: "index" as const,
      intersect: false,
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          color: textMuted,
          padding: 6,
          callback: (tickValue: string | number) =>
            formatYAxisTickSeconds(Number(tickValue)),
        },
        grid: { color: border },
      },
      x: {
        ticks: {
          color: textMuted,
          maxRotation: 45,
          minRotation: 0,
          padding: 4,
        },
        grid: { color: border },
      },
    },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: bgTertiary,
        titleColor: textMuted,
        bodyColor: readCssVar("--text-primary", "#e0e0e0"),
        borderColor: border,
        borderWidth: 1,
        callbacks: {
          title: (items: { dataIndex: number }[]) => {
            const idx = items[0]?.dataIndex ?? 0;
            return props.series[idx]?.date ?? "";
          },
          label: (ctx: { parsed: { y: number | null } }) => {
            const y = ctx.parsed.y;
            if (y == null) return "";
            return formatPlayDurationHMS(y);
          },
        },
      },
    },
  };
}

function createOrUpdateChart(): void {
  const canvas = canvasRef.value;
  if (!canvas) {
    return;
  }
  if (props.series.length === 0) {
    chart?.destroy();
    chart = null;
    return;
  }

  const labels = props.series.map((s) => s.label);
  const data = props.series.map((s) => s.seconds);
  const accent = readCssVar("--accent", "#5b9bd5");

  if (chart) {
    chart.data.labels = labels;
    const ds = chart.data.datasets[0];
    if (ds) {
      ds.data = data;
      ds.borderColor = accent;
      ds.backgroundColor = accent + "33";
    }
    chart.options = buildChartOptions();
    chart.update();
    return;
  }

  chart = new Chart(canvas, {
    type: "line",
    data: {
      labels,
      datasets: [
        {
          label: "",
          data,
          borderColor: accent,
          backgroundColor: accent + "33",
          fill: false,
          tension: 0.2,
          pointRadius: 4,
          pointHoverRadius: 6,
          /* 最大値付近の点がチャート領域の上端で欠けないように */
          clip: false,
        },
      ],
    },
    options: buildChartOptions(),
  });
}

watch(
  () => props.series,
  () => {
    void nextTick(() => createOrUpdateChart());
  },
  { deep: true },
);

onMounted(() => {
  void nextTick(() => createOrUpdateChart());
});

onBeforeUnmount(() => {
  chart?.destroy();
  chart = null;
});
</script>

<style scoped>
.playtime-chart-wrap {
  position: relative;
  width: 100%;
  height: 280px;
}
</style>
