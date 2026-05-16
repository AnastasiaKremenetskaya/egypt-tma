<script setup lang="ts">
import { computed } from 'vue'
import { useGame } from '../composables/useGame'

const emit = defineEmits<{ (e: 'started'): void }>()

const { state, isAdmin, userId, startGame } = useGame()
const room = computed(() => state.room!)

const botUsername = import.meta.env.VITE_BOT_USERNAME ?? ''
const inviteLink = computed(() =>
  botUsername ? `https://t.me/${botUsername}?start=${room.value.code}` : ''
)

async function onStart() {
  await startGame()
  emit('started')
}

function copyCode() {
  navigator.clipboard.writeText(room.value.code)
}
</script>

<template>
  <div class="lobby">
    <!-- bg stars -->
    <div class="bg-layer" />

    <header class="lobby-header">
      <div class="header-icon">🏛️</div>
      <h1 class="header-title">ХРАМ</h1>
      <p class="header-sub">Суд Осириса</p>
    </header>

    <div class="code-card" @click="copyCode">
      <span class="code-label">Код комнаты</span>
      <span class="code-value">{{ room.code }}</span>
      <span class="code-hint">нажми чтобы скопировать</span>
    </div>

    <div v-if="inviteLink" class="invite-card">
      <a :href="inviteLink" class="invite-link">🔗 Пригласить в Храм</a>
    </div>

    <div class="players-card">
      <div class="players-header">
        <span>Участники</span>
        <span class="players-count">{{ room.players.length }}</span>
      </div>
      <div class="players-list">
        <div
          v-for="p in room.players"
          :key="p.user_id"
          class="player-row"
          :class="{ me: p.user_id === userId, admin: p.user_id === room.admin_id }"
        >
          <div class="player-avatar">{{ p.username[0]?.toUpperCase() }}</div>
          <div class="player-info">
            <span class="player-name">{{ p.username }}</span>
            <span class="player-title">«{{ p.title }}»</span>
          </div>
          <span v-if="p.user_id === room.admin_id" class="admin-badge">Жрец</span>
        </div>
      </div>
    </div>

    <div class="action-area">
      <template v-if="isAdmin">
        <p v-if="room.players.length < 2" class="wait-hint">
          Нужно минимум 2 участника для начала суда
        </p>
        <button class="btn-gold" :disabled="room.players.length < 2" @click="onStart">
          ⚖️ Начать Суд
        </button>
      </template>
      <template v-else>
        <p class="wait-hint">
          Ждём когда организатор начнёт суд…
        </p>
        <div class="waiting-dots">
          <span v-for="i in 3" :key="i" class="dot" :style="{ animationDelay: (i * .2) + 's' }" />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.lobby {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 100vh;
  padding: 20px 16px 32px;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
  position: relative;
  overflow-y: auto;
}

.bg-layer {
  position: fixed;
  inset: 0;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
  z-index: -1;
}

.lobby-header { text-align: center; padding: 16px 0 8px; }
.header-icon { font-size: 40px; margin-bottom: 6px; }
.header-title { font-size: 22px; letter-spacing: 4px; color: #e8b84b; }
.header-sub { font-size: 13px; color: #8a9bb5; font-style: italic; }

.code-card {
  background: rgba(201,146,42,.1);
  border: 1px solid rgba(201,146,42,.35);
  border-radius: 14px;
  padding: 16px;
  text-align: center;
  cursor: pointer;
  transition: background .2s;
}
.code-card:hover { background: rgba(201,146,42,.16); }
.code-label { display: block; font-size: 11px; color: #8a9bb5; letter-spacing: 1px; margin-bottom: 6px; }
.code-value {
  display: block;
  font-size: 30px;
  font-weight: bold;
  color: #e8b84b;
  letter-spacing: 6px;
  text-shadow: 0 0 16px rgba(232,184,75,.4);
}
.code-hint { display: block; font-size: 10px; color: #8a9bb5; margin-top: 6px; opacity: .6; }

.invite-card {
  text-align: center;
  background: rgba(30,48,84,.5);
  border: 1px solid rgba(201,146,42,.2);
  border-radius: 12px;
  padding: 12px;
}
.invite-link {
  color: #e8b84b;
  text-decoration: none;
  font-size: 14px;
  font-weight: bold;
}

.players-card {
  background: rgba(26,39,68,.6);
  border: 1px solid rgba(201,146,42,.2);
  border-radius: 14px;
  padding: 16px;
}
.players-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  color: #8a9bb5;
  letter-spacing: .5px;
}
.players-count {
  background: rgba(201,146,42,.2);
  color: #e8b84b;
  border-radius: 20px;
  padding: 2px 10px;
  font-size: 12px;
  font-weight: bold;
}
.players-list { display: flex; flex-direction: column; gap: 8px; }

.player-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(255,255,255,.04);
  border-radius: 10px;
  border: 1px solid transparent;
  transition: border-color .2s;
}
.player-row.me { border-color: rgba(201,146,42,.35); background: rgba(201,146,42,.07); }
.player-row.admin .player-name::after { content: ' ✦'; color: #c9922a; }

.player-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: rgba(201,146,42,.2);
  border: 1px solid rgba(201,146,42,.35);
  color: #e8b84b;
  font-size: 15px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.player-info { flex: 1; min-width: 0; }
.player-name { display: block; font-size: 14px; color: #fef9e7; }
.player-title { display: block; font-size: 12px; color: #8a9bb5; font-style: italic; }
.admin-badge {
  font-size: 10px;
  color: #c9922a;
  border: 1px solid rgba(201,146,42,.4);
  border-radius: 6px;
  padding: 2px 7px;
  white-space: nowrap;
}

.action-area { margin-top: 4px; }
.btn-gold {
  width: 100%;
  padding: 15px;
  background: linear-gradient(135deg, #c9922a, #e8b84b);
  color: #04060f;
  border: none;
  border-radius: 12px;
  font-family: inherit;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  transition: opacity .2s, transform .1s;
}
.btn-gold:disabled { opacity: .35; cursor: default; }
.btn-gold:active { transform: scale(.97); }

.wait-hint { text-align: center; color: #8a9bb5; font-size: 13px; font-style: italic; margin-bottom: 14px; }

.waiting-dots { display: flex; justify-content: center; gap: 8px; }
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(201,146,42,.5);
  animation: bounce .8s infinite alternate;
}
@keyframes bounce { to { opacity: 1; transform: translateY(-6px); } }
</style>
