import { reactive, computed } from 'vue'
import type { RoomState, PlayerView } from '../types/game'
import { useTelegram } from './useTelegram'

const { initData, userId, username } = useTelegram()

const state = reactive<{
  room: RoomState | null
  loading: boolean
  error: string | null
}>({
  room: null,
  loading: false,
  error: null,
})

// ─── Derived state ─────────────────────────────────────────────────────────────

const myPlayer = computed<PlayerView | undefined>(() =>
  state.room?.players.find((p) => p.user_id === userId)
)

const activePlayer = computed<PlayerView | undefined>(() => {
  const r = state.room
  if (!r) return undefined
  return r.players[r.active_idx]
})

const isAdmin = computed(() => state.room?.admin_id === userId)
const isActivePlayer = computed(() => activePlayer.value?.user_id === userId)
const hasVoted = computed(() => state.room?.voted_ids.includes(userId) ?? false)
const hasSethAnswered = computed(() => state.room?.seth_answered_ids.includes(userId) ?? false)

// ─── API helpers ──────────────────────────────────────────────────────────────

const headers = () => ({
  'Content-Type': 'application/json',
  'X-Telegram-Init-Data': initData,
})

async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: headers(),
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? 'Ошибка сервера')
  return data as T
}

async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(path, { headers: headers() })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? 'Ошибка сервера')
  return data as T
}

// ─── Actions ──────────────────────────────────────────────────────────────────

async function createRoom(): Promise<string> {
  state.loading = true
  state.error = null
  try {
    const s = await apiPost<RoomState>('/api/room')
    state.room = s
    return s.code
  } catch (e: unknown) {
    state.error = e instanceof Error ? e.message : String(e)
    throw e
  } finally {
    state.loading = false
  }
}

async function joinRoom(code: string): Promise<void> {
  state.loading = true
  state.error = null
  try {
    const s = await apiPost<RoomState>(`/api/room/${code}/join`)
    state.room = s
  } catch (e: unknown) {
    state.error = e instanceof Error ? e.message : String(e)
    throw e
  } finally {
    state.loading = false
  }
}

async function fetchRoom(code: string): Promise<void> {
  try {
    const s = await apiGet<RoomState>(`/api/room/${code}`)
    state.room = s
  } catch { /* ignore */ }
}

async function startGame(): Promise<void> {
  if (!state.room) return
  await apiPost(`/api/room/${state.room.code}/start`)
}

async function submitAnswer(text: string): Promise<void> {
  if (!state.room) return
  await apiPost(`/api/room/${state.room.code}/answer`, { text })
}

async function submitVoice(): Promise<void> {
  if (!state.room) return
  await apiPost(`/api/room/${state.room.code}/voice`)
}

async function submitVote(trust: boolean): Promise<void> {
  if (!state.room) return
  await apiPost(`/api/room/${state.room.code}/vote`, { trust })
}

async function submitSeth(option: number): Promise<void> {
  if (!state.room) return
  await apiPost(`/api/room/${state.room.code}/seth`, { option })
}

function applyState(s: RoomState) {
  state.room = s
}

export function useGame() {
  return {
    state,
    myPlayer,
    activePlayer,
    isAdmin,
    isActivePlayer,
    hasVoted,
    hasSethAnswered,
    userId,
    username,
    createRoom,
    joinRoom,
    fetchRoom,
    startGame,
    submitAnswer,
    submitVoice,
    submitVote,
    submitSeth,
    applyState,
  }
}
