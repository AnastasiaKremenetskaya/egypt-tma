import { ref, onUnmounted } from 'vue'
import type { WSMessage, RoomState } from '../types/game'
import { API_BASE } from './useGame'

export function useWebSocket(onState: (state: RoomState) => void) {
  const connected = ref(false)
  let ws: WebSocket | null = null
  let roomCode = ''
  let initData = ''
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  function connect(code: string, tgInitData: string) {
    roomCode = code
    initData = tgInitData
    _open()
  }

  function disconnect() {
    destroyed = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
  }

  function _open() {
    if (destroyed) return

    const encoded = encodeURIComponent(initData)
    const base = API_BASE
      ? API_BASE.replace(/^http/, 'ws')
      : (location.protocol === 'https:' ? 'wss' : 'ws') + '://' + location.host
    const url = `${base}/ws/room/${roomCode}?init_data=${encoded}`

    ws = new WebSocket(url)

    ws.onopen = () => { connected.value = true }

    ws.onmessage = (evt) => {
      try {
        const msg: WSMessage = JSON.parse(evt.data)
        if (msg.type === 'state') onState(msg.state)
      } catch { /* ignore */ }
    }

    ws.onclose = () => {
      connected.value = false
      if (!destroyed) {
        reconnectTimer = setTimeout(_open, 2000)
      }
    }

    ws.onerror = () => ws?.close()
  }

  onUnmounted(disconnect)

  return { connected, connect, disconnect }
}
