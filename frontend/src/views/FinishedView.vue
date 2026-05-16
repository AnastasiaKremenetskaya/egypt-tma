<script setup lang="ts">
import { computed } from 'vue'
import { useGame } from '../composables/useGame'
import ScoreBoard from '../components/ScoreBoard.vue'

const { state, userId } = useGame()
const room = computed(() => state.room!)

const winner = computed(() => {
  if (!room.value) return null
  return [...room.value.players].sort((a, b) => b.score - a.score)[0]
})

const isWinner = computed(() => winner.value?.user_id === userId)
</script>

<template>
  <div class="finished">
    <div class="bg-layer" />

    <div class="winner-hero">
      <div class="winner-crown">{{ isWinner ? '👑' : '🏛️' }}</div>
      <h1 class="winner-title">{{ isWinner ? 'Ты победил!' : 'Суд завершён' }}</h1>
      <p v-if="winner" class="winner-name">
        {{ winner.username }} «{{ winner.title }}»
        <br>
        <span class="winner-score">{{ winner.score }} очков</span>
      </p>
      <p class="winner-sub">{{ isWinner ? 'Маат торжествует! Твоё сердце чисто.' : 'Боги вынесли приговор.' }}</p>
    </div>

    <div class="score-wrap">
      <ScoreBoard :players="room.players" :my-user-id="userId" />
    </div>
  </div>
</template>

<style scoped>
.finished {
  min-height: 100vh;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 16px;
  gap: 24px;
  position: relative;
}
.bg-layer { position: fixed; inset: 0; background: inherit; z-index: -1; }

.winner-hero { text-align: center; }
.winner-crown { font-size: 64px; margin-bottom: 12px; filter: drop-shadow(0 0 20px rgba(201,146,42,.6)); }
.winner-title { font-size: 26px; letter-spacing: 2px; color: #e8b84b; font-weight: bold; }
.winner-name { margin-top: 12px; color: #fef9e7; font-size: 15px; line-height: 1.6; }
.winner-score { font-size: 24px; font-weight: bold; color: #e8b84b; }
.winner-sub { margin-top: 10px; color: #8a9bb5; font-size: 13px; font-style: italic; }

.score-wrap { width: 100%; max-width: 400px; }
</style>
