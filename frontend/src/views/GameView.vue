<script setup lang="ts">
import { computed, ref } from 'vue'
import { useGame } from '../composables/useGame'
import QuestionPhase from '../components/QuestionPhase.vue'
import VotingPhase from '../components/VotingPhase.vue'
import SethPhase from '../components/SethPhase.vue'
import ScoreBoard from '../components/ScoreBoard.vue'

const {
  state,
  activePlayer,
  isActivePlayer,
  hasVoted,
  hasSethAnswered,
  userId,
  submitAnswer,
  submitVoice,
  submitVote,
  submitSeth,
} = useGame()

const room = computed(() => state.room!)
const tab = ref<'game' | 'score'>('game')

const nonActiveCount = computed(() =>
  room.value.players.filter(p => p.user_id !== activePlayer.value?.user_id).length
)
</script>

<template>
  <div class="game-view">
    <div class="bg-layer" />

    <!-- Header with tabs -->
    <header class="game-header">
      <div class="round-info">
        <span class="round-label">Раунд</span>
        <span class="round-num">{{ room.round }}</span>
      </div>
      <div class="tabs">
        <button :class="['tab', { active: tab === 'game' }]" @click="tab = 'game'">
          {{ room.phase === 'question' ? '🪶 Вопрос' : room.phase === 'voting' ? '⚖️ Суд' : '🌪 Сет' }}
        </button>
        <button :class="['tab', { active: tab === 'score' }]" @click="tab = 'score'">
          📊 Счёт
        </button>
      </div>
    </header>

    <main class="game-main">
      <!-- Score tab -->
      <template v-if="tab === 'score'">
        <ScoreBoard :players="room.players" :my-user-id="userId" />
      </template>

      <!-- Game tab -->
      <template v-else>
        <QuestionPhase
          v-if="room.phase === 'question' && room.question"
          :question="room.question"
          :active-player="activePlayer!"
          :is-active-player="isActivePlayer"
          :deadline="room.phase_deadline"
          @answer="submitAnswer"
          @voice="submitVoice"
        />

        <VotingPhase
          v-else-if="room.phase === 'voting'"
          :active-player="activePlayer!"
          :answer="room.answer"
          :vote-trust="room.vote_trust"
          :vote-lie="room.vote_lie"
          :voted-ids="room.voted_ids"
          :total-voters="nonActiveCount"
          :is-active-player="isActivePlayer"
          :has-voted="hasVoted"
          :deadline="room.phase_deadline"
          @vote="submitVote"
        />

        <SethPhase
          v-else-if="room.phase === 'seth' && room.question"
          :question="room.question"
          :seth-answered-ids="room.seth_answered_ids"
          :players="room.players"
          :has-seth-answered="hasSethAnswered"
          :deadline="room.phase_deadline"
          @seth="submitSeth"
        />

        <div v-else class="loading-state">
          <span class="loading-icon">⚖️</span>
          <p>Суд продолжается…</p>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.game-view {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
  position: relative;
}
.bg-layer {
  position: fixed;
  inset: 0;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
  z-index: -1;
}

.game-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(201,146,42,.15);
  background: rgba(8,14,28,.6);
  backdrop-filter: blur(8px);
  position: sticky;
  top: 0;
  z-index: 10;
}
.round-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 4px 12px;
  background: rgba(201,146,42,.1);
  border: 1px solid rgba(201,146,42,.25);
  border-radius: 8px;
  min-width: 48px;
}
.round-label { font-size: 9px; color: #8a9bb5; letter-spacing: 1px; }
.round-num { font-size: 18px; font-weight: bold; color: #e8b84b; line-height: 1; }

.tabs {
  flex: 1;
  display: flex;
  background: rgba(0,0,0,.3);
  border-radius: 10px;
  overflow: hidden;
}
.tab {
  flex: 1;
  padding: 8px 6px;
  background: none;
  border: none;
  color: #8a9bb5;
  font-family: inherit;
  font-size: 12px;
  cursor: pointer;
  transition: background .2s, color .2s;
}
.tab.active { background: rgba(201,146,42,.2); color: #e8b84b; }

.game-main {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  min-height: 200px;
  color: #8a9bb5;
  font-style: italic;
}
.loading-icon { font-size: 40px; animation: spin 2s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
