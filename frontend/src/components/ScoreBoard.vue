<script setup lang="ts">
import type { PlayerView } from '../types/game'

defineProps<{ players: PlayerView[]; myUserId: number }>()

const WIN_SCORE = 50

function scrollWidth(score: number) {
  return Math.min(100, Math.round((score / WIN_SCORE) * 100)) + '%'
}
</script>

<template>
  <div class="scoreboard">
    <div class="sb-header">
      <span class="sb-title">Свиток Осириса</span>
      <span class="sb-goal">до {{ WIN_SCORE }} очков</span>
    </div>

    <div
      v-for="(p, i) in players"
      :key="p.user_id"
      class="sb-row"
      :class="{ me: p.user_id === myUserId }"
    >
      <span class="sb-rank">{{ i === 0 ? '👑' : i + 1 }}</span>
      <div class="sb-info">
        <div class="sb-name">{{ p.username }} <span class="sb-title-badge">«{{ p.title }}»</span></div>
        <div class="sb-bar-wrap">
          <div class="sb-bar-fill" :style="{ width: scrollWidth(p.score) }" />
        </div>
      </div>
      <span class="sb-score">{{ p.score }}</span>
    </div>
  </div>
</template>

<style scoped>
.scoreboard {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 0;
}
.sb-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 0 4px 6px;
  border-bottom: 1px solid rgba(201,146,42,.25);
}
.sb-title { font-size: 15px; color: #e8b84b; font-weight: bold; letter-spacing: .5px; }
.sb-goal { font-size: 11px; color: #8a9bb5; }

.sb-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(255,255,255,.04);
  border-radius: 10px;
  border: 1px solid rgba(201,146,42,.12);
  transition: border-color .2s;
}
.sb-row.me {
  border-color: rgba(201,146,42,.45);
  background: rgba(201,146,42,.08);
}
.sb-rank { font-size: 18px; min-width: 26px; text-align: center; }
.sb-info { flex: 1; min-width: 0; }
.sb-name { font-size: 13px; color: #fef9e7; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sb-title-badge { color: #8a9bb5; font-style: italic; }
.sb-bar-wrap {
  margin-top: 5px;
  height: 4px;
  background: rgba(255,255,255,.1);
  border-radius: 2px;
  overflow: hidden;
}
.sb-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #c9922a, #f5d070);
  border-radius: 2px;
  transition: width .6s ease;
}
.sb-score {
  font-size: 20px;
  font-weight: bold;
  color: #e8b84b;
  min-width: 38px;
  text-align: right;
}
</style>
