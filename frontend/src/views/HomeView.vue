<script setup lang="ts">
import { ref } from 'vue'
import { useGame } from '../composables/useGame'

const emit = defineEmits<{ (e: 'joined'): void }>()

const { createRoom, joinRoom, state, username } = useGame()

const code = ref('')
const tab = ref<'join' | 'create'>('join')
const localError = ref('')

async function onJoin() {
  localError.value = ''
  const c = code.value.trim().toUpperCase()
  if (c.length < 4) { localError.value = 'Введи код комнаты'; return }
  try {
    await joinRoom(c)
    emit('joined')
  } catch (e: unknown) {
    localError.value = e instanceof Error ? e.message : 'Ошибка'
  }
}

async function onCreate() {
  localError.value = ''
  try {
    await createRoom()
    emit('joined')
  } catch (e: unknown) {
    localError.value = e instanceof Error ? e.message : 'Ошибка'
  }
}
</script>

<template>
  <div class="home">
    <!-- Stars bg -->
    <div class="bg-stars">
      <div v-for="i in 30" :key="i" class="star" :style="{
        left: Math.random() * 100 + '%',
        top: Math.random() * 60 + '%',
        width: (Math.random() * 2 + 1) + 'px',
        height: (Math.random() * 2 + 1) + 'px',
        animationDelay: Math.random() * 4 + 's',
        animationDuration: (2 + Math.random() * 3) + 's',
      }" />
    </div>

    <div class="hero">
      <div class="hero-icon">⚖️</div>
      <h1 class="hero-title">МААТ И СЕТ</h1>
      <p class="hero-sub">Суд Осириса ждёт, {{ username }}</p>
    </div>

    <div class="card">
      <div class="tabs">
        <button :class="['tab', { active: tab === 'join' }]" @click="tab = 'join'">Войти в храм</button>
        <button :class="['tab', { active: tab === 'create' }]" @click="tab = 'create'">Создать храм</button>
      </div>

      <div v-if="tab === 'join'" class="tab-content">
        <p class="tab-hint">Введи код комнаты, полученный от организатора</p>
        <input
          v-model="code"
          class="code-input"
          placeholder="ANKH42"
          maxlength="8"
          spellcheck="false"
          @keydown.enter="onJoin"
        />
        <button class="btn-gold" :disabled="state.loading" @click="onJoin">
          {{ state.loading ? 'Входим…' : '🚪 Войти в Храм' }}
        </button>
      </div>

      <div v-else class="tab-content">
        <p class="tab-hint">Ты станешь организатором нового суда Осириса</p>
        <button class="btn-gold" :disabled="state.loading" @click="onCreate">
          {{ state.loading ? 'Создаём…' : '✨ Создать комнату' }}
        </button>
      </div>

      <p v-if="localError || state.error" class="error-msg">
        {{ localError || state.error }}
      </p>
    </div>

    <p class="footer-hint">Игра-трибунал в стиле Древнего Египта · 2–8 игроков</p>
  </div>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px 20px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(180deg, #050a16 0%, #10203a 60%, #1a2744 100%);
}

.bg-stars { position: absolute; inset: 0; pointer-events: none; }
.star {
  position: absolute;
  background: #f5d070;
  border-radius: 50%;
  animation: twinkle 3s infinite alternate;
}
@keyframes twinkle { from { opacity: .2; transform: scale(1); } to { opacity: .9; transform: scale(1.4); } }

.hero { text-align: center; margin-bottom: 32px; position: relative; z-index: 1; }
.hero-icon { font-size: 56px; margin-bottom: 12px; filter: drop-shadow(0 0 16px rgba(201,146,42,.5)); }
.hero-title { font-size: 28px; letter-spacing: 4px; color: #e8b84b; font-weight: bold; text-shadow: 0 0 20px rgba(232,184,75,.4); }
.hero-sub { margin-top: 8px; color: #8a9bb5; font-size: 14px; font-style: italic; }

.card {
  width: 100%;
  max-width: 360px;
  background: rgba(26,39,68,.85);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 18px;
  padding: 22px;
  backdrop-filter: blur(12px);
  position: relative;
  z-index: 1;
}

.tabs { display: flex; border-radius: 10px; overflow: hidden; background: rgba(0,0,0,.3); margin-bottom: 18px; }
.tab {
  flex: 1;
  padding: 10px;
  background: none;
  border: none;
  color: #8a9bb5;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: background .2s, color .2s;
}
.tab.active { background: rgba(201,146,42,.2); color: #e8b84b; }

.tab-content { display: flex; flex-direction: column; gap: 12px; }
.tab-hint { font-size: 13px; color: #8a9bb5; text-align: center; line-height: 1.4; }

.code-input {
  width: 100%;
  padding: 14px;
  background: rgba(255,255,255,.06);
  border: 1px solid rgba(201,146,42,.3);
  border-radius: 10px;
  color: #fef9e7;
  font-family: 'Georgia', serif;
  font-size: 20px;
  letter-spacing: 4px;
  text-align: center;
  text-transform: uppercase;
  outline: none;
  transition: border-color .2s;
}
.code-input:focus { border-color: #c9922a; }
.code-input::placeholder { color: #8a9bb5; letter-spacing: 2px; font-size: 16px; }

.btn-gold {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #c9922a, #e8b84b);
  color: #04060f;
  border: none;
  border-radius: 10px;
  font-family: inherit;
  font-size: 15px;
  font-weight: bold;
  cursor: pointer;
  transition: opacity .2s, transform .1s;
}
.btn-gold:hover { opacity: .9; }
.btn-gold:active { transform: scale(.97); }
.btn-gold:disabled { opacity: .4; cursor: default; }

.error-msg {
  text-align: center;
  color: #e84b4b;
  font-size: 13px;
  padding: 10px;
  background: rgba(232,75,75,.1);
  border-radius: 8px;
  border: 1px solid rgba(232,75,75,.25);
}

.footer-hint { margin-top: 20px; font-size: 11px; color: #8a9bb5; text-align: center; position: relative; z-index: 1; }
</style>
