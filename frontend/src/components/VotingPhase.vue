<script setup lang="ts">
import type { AnswerView, PlayerView } from '../types/game'
import TimerBar from './TimerBar.vue'

defineProps<{
  activePlayer: PlayerView
  answer: AnswerView | undefined
  voteTrust: number
  voteLie: number
  votedIds: number[]
  totalVoters: number
  isActivePlayer: boolean
  hasVoted: boolean
  deadline: string
}>()

const emit = defineEmits<{ (e: 'vote', trust: boolean): void }>()
</script>

<template>
  <div class="phase-wrap">
    <div class="phase-header">
      <div class="active-badge">
        <span>⚖️</span>
        <span class="active-name">{{ activePlayer.username }}</span>
        <span class="active-title">«{{ activePlayer.title }}»</span>
        <span class="phase-label">· Голосование</span>
      </div>
      <TimerBar :deadline="deadline" :total-seconds="30" />
    </div>

    <div class="answer-card">
      <div class="ans-label">Ответ жреца:</div>
      <p v-if="answer?.type === 'text'" class="ans-text">«{{ answer.text }}»</p>
      <p v-else class="ans-text voice"><em>(ответил вслух)</em></p>
    </div>

    <div class="vote-progress">
      <div class="vp-side trust">
        <span class="vp-icon">🪶</span>
        <span class="vp-count">{{ voteTrust }}</span>
        <div class="gems">
          <span v-for="n in voteTrust" :key="n" class="gem trust-gem">🪶</span>
        </div>
      </div>
      <div class="vp-divider">|</div>
      <div class="vp-side lie">
        <div class="gems">
          <span v-for="n in voteLie" :key="n" class="gem lie-gem">👁️</span>
        </div>
        <span class="vp-count">{{ voteLie }}</span>
        <span class="vp-icon">👁️</span>
      </div>
    </div>

    <p class="vp-hint">{{ votedIds.length }} / {{ totalVoters }} проголосовали</p>

    <template v-if="!isActivePlayer">
      <div v-if="hasVoted" class="voted-msg">Твой голос учтён. Весы Маат движутся…</div>
      <div v-else class="vote-btns">
        <button class="btn-trust" @click="emit('vote', true)">
          🪶 Перо Маат<br><small>Говорит правду</small>
        </button>
        <button class="btn-lie" @click="emit('vote', false)">
          👁️ Гнев Аммит<br><small>Лжёт!</small>
        </button>
      </div>
    </template>
    <template v-else>
      <p class="hint">Жрецы судят тебя… Боги наблюдают за весами.</p>
    </template>
  </div>
</template>

<style scoped>
.phase-wrap { display: flex; flex-direction: column; gap: 14px; }

.phase-header { display: flex; flex-direction: column; gap: 8px; }
.active-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(201,146,42,.12);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 20px;
  padding: 6px 14px;
  font-size: 13px;
}
.active-name { color: #e8b84b; font-weight: bold; }
.active-title { color: #8a9bb5; font-style: italic; }
.phase-label { color: #8a9bb5; font-size: 11px; }

.answer-card {
  background: rgba(30,48,84,.7);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 14px;
  padding: 16px 18px;
}
.ans-label { font-size: 11px; color: #8a9bb5; margin-bottom: 8px; letter-spacing: .5px; }
.ans-text { font-size: 15px; color: #fef9e7; line-height: 1.5; font-style: italic; }
.ans-text.voice { color: #8a9bb5; }

.vote-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px;
  background: rgba(255,255,255,.03);
  border-radius: 14px;
  border: 1px solid rgba(255,255,255,.08);
  min-height: 60px;
}
.vp-side { display: flex; align-items: center; gap: 8px; flex: 1; }
.vp-side.trust { justify-content: flex-start; }
.vp-side.lie { justify-content: flex-end; }
.vp-icon { font-size: 20px; }
.vp-count { font-size: 22px; font-weight: bold; color: #e8b84b; min-width: 24px; }
.vp-divider { color: rgba(255,255,255,.2); font-size: 20px; }
.gems { display: flex; flex-wrap: wrap; gap: 3px; }
.gem { font-size: 16px; animation: popIn .3s ease; }
@keyframes popIn { from { transform: scale(0); opacity: 0; } to { transform: scale(1); opacity: 1; } }

.vp-hint { text-align: center; font-size: 12px; color: #8a9bb5; }

.voted-msg {
  text-align: center;
  padding: 14px;
  color: #8a9bb5;
  font-style: italic;
  background: rgba(255,255,255,.03);
  border-radius: 10px;
  border: 1px solid rgba(255,255,255,.08);
}

.vote-btns { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.btn-trust, .btn-lie {
  padding: 14px 10px;
  border-radius: 12px;
  font-family: inherit;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  border: none;
  line-height: 1.4;
  transition: opacity .2s, transform .1s;
}
.btn-trust:active, .btn-lie:active { transform: scale(.96); }
.btn-trust { background: rgba(76,175,80,.85); color: #fff; }
.btn-lie { background: rgba(200,60,60,.85); color: #fff; }

small { font-weight: normal; font-size: 11px; opacity: .85; }
.hint { text-align: center; font-size: 12px; color: #8a9bb5; font-style: italic; }
</style>
