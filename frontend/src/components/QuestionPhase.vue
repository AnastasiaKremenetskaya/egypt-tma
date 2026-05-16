<script setup lang="ts">
import { ref } from 'vue'
import type { QuestionView, PlayerView } from '../types/game'
import TimerBar from './TimerBar.vue'

const props = defineProps<{
  question: QuestionView
  activePlayer: PlayerView
  isActivePlayer: boolean
  deadline: string
}>()

const emit = defineEmits<{
  (e: 'answer', text: string): void
  (e: 'voice'): void
}>()

const answerText = ref('')

function sendAnswer() {
  const t = answerText.value.trim()
  if (!t) return
  emit('answer', t)
  answerText.value = ''
}
</script>

<template>
  <div class="phase-wrap">
    <div class="phase-header">
      <div class="active-badge">
        <span class="active-icon">{{ isActivePlayer ? '🪶' : '👁' }}</span>
        <span class="active-name">{{ activePlayer.username }}</span>
        <span class="active-title">«{{ activePlayer.title }}»</span>
      </div>
      <TimerBar :deadline="deadline" :total-seconds="60" />
    </div>

    <div class="question-card" :class="question.type">
      <div class="q-badge">
        {{ question.type === 'maat' ? '🕊️ Папирус Маат · Личный вопрос' : '⚡ Карта Сета · Знания' }}
      </div>
      <p class="q-text">{{ question.text }}</p>
    </div>

    <template v-if="isActivePlayer">
      <div class="answer-area">
        <textarea
          v-model="answerText"
          class="answer-input"
          placeholder="Напиши ответ жрецам…"
          rows="3"
          @keydown.enter.exact.prevent="sendAnswer"
        />
        <button class="btn-gold" :disabled="!answerText.trim()" @click="sendAnswer">
          ⚖️ Отправить ответ
        </button>
      </div>
      <button class="btn-ghost" @click="emit('voice')">🗣️ Я ответил вслух</button>
      <p class="hint">Боги слышат каждое слово. Жрецы будут судить твоё сердце.</p>
    </template>

    <template v-else>
      <div class="witness-card">
        <span class="witness-icon">👁</span>
        <p>Глаз Гора наблюдает.<br><b>{{ activePlayer.username }}</b> держит ответ.</p>
      </div>
      <p class="hint">Жди — скоро встанешь перед весами Маат.</p>
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
.active-icon { font-size: 16px; }
.active-name { color: #e8b84b; font-weight: bold; }
.active-title { color: #8a9bb5; font-style: italic; }

.question-card {
  background: rgba(30,48,84,.7);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 14px;
  padding: 18px;
  backdrop-filter: blur(8px);
}
.question-card.seth { border-color: rgba(180,60,60,.4); background: rgba(50,20,20,.6); }
.q-badge { font-size: 11px; color: #8a9bb5; letter-spacing: .5px; margin-bottom: 10px; }
.q-text { font-size: 15px; line-height: 1.55; color: #fef9e7; font-style: italic; }

.answer-area { display: flex; flex-direction: column; gap: 8px; }
.answer-input {
  width: 100%;
  background: rgba(255,255,255,.06);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 10px;
  color: #fef9e7;
  font-family: inherit;
  font-size: 14px;
  padding: 12px;
  resize: none;
  outline: none;
}
.answer-input:focus { border-color: #c9922a; }
.answer-input::placeholder { color: #8a9bb5; }

.btn-gold {
  width: 100%;
  padding: 13px;
  background: linear-gradient(135deg, #c9922a, #e8b84b);
  color: #04060f;
  border: none;
  border-radius: 10px;
  font-family: inherit;
  font-size: 15px;
  font-weight: bold;
  cursor: pointer;
  transition: opacity .2s;
}
.btn-gold:disabled { opacity: .4; cursor: default; }
.btn-ghost {
  width: 100%;
  padding: 12px;
  background: transparent;
  border: 1px solid rgba(201,146,42,.4);
  border-radius: 10px;
  color: #e8b84b;
  font-family: inherit;
  font-size: 14px;
  cursor: pointer;
  transition: background .2s;
}
.btn-ghost:hover { background: rgba(201,146,42,.1); }

.witness-card {
  text-align: center;
  padding: 24px;
  background: rgba(255,255,255,.04);
  border-radius: 14px;
  border: 1px solid rgba(255,255,255,.08);
  line-height: 1.6;
  color: #8a9bb5;
}
.witness-icon { font-size: 32px; display: block; margin-bottom: 10px; }

.hint { text-align: center; font-size: 12px; color: #8a9bb5; font-style: italic; }
</style>
