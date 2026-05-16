<script setup lang="ts">
import type { QuestionView, PlayerView } from '../types/game'
import TimerBar from './TimerBar.vue'

defineProps<{
  question: QuestionView
  sethAnsweredIds: number[]
  players: PlayerView[]
  hasSethAnswered: boolean
  deadline: string
}>()

const emit = defineEmits<{ (e: 'seth', optIdx: number): void }>()
</script>

<template>
  <div class="phase-wrap">
    <div class="seth-header">
      <span class="seth-icon">🌪</span>
      <div>
        <div class="seth-title">Вопрос Сета!</div>
        <div class="seth-sub">Штрафной раунд · Знания Египта</div>
      </div>
    </div>

    <TimerBar :deadline="deadline" :total-seconds="20" />

    <div class="question-card">
      <p class="q-text">{{ question.text }}</p>
    </div>

    <div class="scoring-hint">
      <span>Активный: верно <b>+2</b> / неверно <b>0</b></span>
      <span>Остальные: верно <b>+1</b> / неверно <b>−1</b></span>
    </div>

    <div v-if="hasSethAnswered" class="answered-msg">
      Твой ответ записан. Ждём остальных…
      <div class="answered-count">{{ sethAnsweredIds.length }} / {{ players.length }}</div>
    </div>

    <div v-else-if="question.options" class="options">
      <button
        v-for="(opt, i) in question.options"
        :key="i"
        class="opt-btn"
        @click="emit('seth', i)"
      >
        <span class="opt-letter">{{ String.fromCharCode(65 + i) }}</span>
        <span class="opt-text">{{ opt }}</span>
      </button>
    </div>

    <div class="waiting-players">
      <span
        v-for="p in players"
        :key="p.user_id"
        class="player-dot"
        :class="{ answered: sethAnsweredIds.includes(p.user_id) }"
        :title="p.username"
      >{{ p.username[0].toUpperCase() }}</span>
    </div>
  </div>
</template>

<style scoped>
.phase-wrap { display: flex; flex-direction: column; gap: 14px; }

.seth-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(180,60,60,.15);
  border: 1px solid rgba(180,60,60,.35);
  border-radius: 12px;
}
.seth-icon { font-size: 28px; }
.seth-title { font-size: 16px; font-weight: bold; color: #e84b4b; }
.seth-sub { font-size: 11px; color: #8a9bb5; }

.question-card {
  background: rgba(50,20,20,.6);
  border: 1px solid rgba(180,60,60,.4);
  border-radius: 14px;
  padding: 18px;
}
.q-text { font-size: 15px; color: #fef9e7; line-height: 1.55; font-style: italic; }

.scoring-hint {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #8a9bb5;
  padding: 0 4px;
}

.options { display: flex; flex-direction: column; gap: 8px; }
.opt-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  background: rgba(255,255,255,.05);
  border: 1px solid rgba(201,146,42,.2);
  border-radius: 10px;
  font-family: inherit;
  font-size: 14px;
  color: #fef9e7;
  cursor: pointer;
  text-align: left;
  transition: background .2s, border-color .2s, transform .1s;
}
.opt-btn:hover { background: rgba(201,146,42,.12); border-color: rgba(201,146,42,.4); }
.opt-btn:active { transform: scale(.97); }
.opt-letter {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(201,146,42,.25);
  color: #e8b84b;
  font-weight: bold;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.opt-text { flex: 1; }

.answered-msg {
  text-align: center;
  padding: 20px;
  color: #8a9bb5;
  font-style: italic;
  background: rgba(255,255,255,.03);
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,.08);
}
.answered-count { margin-top: 8px; font-size: 20px; font-weight: bold; color: #e8b84b; }

.waiting-players {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}
.player-dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255,255,255,.08);
  border: 2px solid rgba(255,255,255,.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: bold;
  color: #8a9bb5;
  transition: background .3s, border-color .3s;
}
.player-dot.answered {
  background: rgba(76,175,80,.25);
  border-color: rgba(76,175,80,.6);
  color: #4caf50;
}
</style>
