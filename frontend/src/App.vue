<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useGame } from './composables/useGame'
import { useTelegram } from './composables/useTelegram'
import { useWebSocket } from './composables/useWebSocket'
import HomeView from './views/HomeView.vue'
import LobbyView from './views/LobbyView.vue'
import GameView from './views/GameView.vue'
import FinishedView from './views/FinishedView.vue'

const { state, applyState, fetchRoom } = useGame()
const { ready, initData } = useTelegram()

const phase = computed(() => state.room?.phase ?? null)

const { connect } = useWebSocket((newState) => {
  applyState(newState)
})

function onJoined() {
  if (state.room?.code) {
    connect(state.room.code, initData)
  }
}

onMounted(() => {
  ready()
  // Try to restore session from localStorage
  const saved = localStorage.getItem('egypt_room_code')
  if (saved) {
    fetchRoom(saved).then(() => {
      if (state.room) connect(saved, initData)
    })
  }
})

// Persist room code so refresh re-joins
const roomCode = computed(() => state.room?.code)
</script>

<template>
  <div id="root">
    <HomeView v-if="!phase" @joined="onJoined" />
    <LobbyView v-else-if="phase === 'lobby'" @started="() => {}" />
    <GameView v-else-if="phase === 'question' || phase === 'voting' || phase === 'seth'" />
    <FinishedView v-else-if="phase === 'finished'" />
  </div>
</template>

<style>
#root { width: 100%; min-height: 100vh; }
</style>
