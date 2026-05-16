<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps<{ deadline: string; totalSeconds: number }>()

const pct = ref(100)
let raf = 0

function tick() {
  const now = Date.now()
  const end = new Date(props.deadline).getTime()
  const total = props.totalSeconds * 1000
  const remaining = Math.max(0, end - now)
  pct.value = Math.round((remaining / total) * 100)
  if (remaining > 0) raf = requestAnimationFrame(tick)
}

function start() {
  cancelAnimationFrame(raf)
  tick()
}

watch(() => props.deadline, start)
onMounted(start)
onUnmounted(() => cancelAnimationFrame(raf))

const secondsLeft = ref(0)
function tickSeconds() {
  const end = new Date(props.deadline).getTime()
  secondsLeft.value = Math.max(0, Math.ceil((end - Date.now()) / 1000))
}
let secInterval = 0
onMounted(() => { tickSeconds(); secInterval = setInterval(tickSeconds, 500) as unknown as number })
onUnmounted(() => clearInterval(secInterval))
</script>

<template>
  <div class="timer-wrap">
    <div class="timer-bar" :class="{ warn: pct < 30, crit: pct < 10 }">
      <div class="timer-fill" :style="{ width: pct + '%' }" />
    </div>
    <span class="timer-seconds">{{ secondsLeft }}с</span>
  </div>
</template>

<style scoped>
.timer-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.timer-bar {
  flex: 1;
  height: 6px;
  background: rgba(255,255,255,.12);
  border-radius: 3px;
  overflow: hidden;
}
.timer-fill {
  height: 100%;
  background: linear-gradient(90deg, #c9922a, #f5d070);
  border-radius: 3px;
  transition: width .2s linear, background .3s;
}
.timer-bar.warn .timer-fill { background: linear-gradient(90deg, #c9922a, #e84b4b); }
.timer-bar.crit .timer-fill { background: #e84b4b; animation: pulse .5s infinite alternate; }
@keyframes pulse { to { opacity: .5; } }
.timer-seconds {
  font-size: 12px;
  color: #8a9bb5;
  min-width: 28px;
  text-align: right;
}
</style>
