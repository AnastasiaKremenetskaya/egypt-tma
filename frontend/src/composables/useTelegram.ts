declare global {
  interface Window {
    Telegram?: {
      WebApp: {
        ready(): void
        expand(): void
        initData: string
        initDataUnsafe: {
          user?: {
            id: number
            username?: string
            first_name?: string
          }
        }
        colorScheme: 'light' | 'dark'
        themeParams: Record<string, string>
      }
    }
  }
}

const tg = window.Telegram?.WebApp

// ─── Dev mode helpers (only active when Telegram SDK is not available) ─────────

const DEV_USER_KEY = 'egypt_dev_user'

interface DevUser {
  id: number
  username: string
}

const DEFAULT_DEV_USER: DevUser = { id: 111111, username: 'dev_player1' }

function loadDevUser(): DevUser {
  try {
    const raw = localStorage.getItem(DEV_USER_KEY)
    if (raw) return JSON.parse(raw) as DevUser
  } catch { /* ignore */ }
  return DEFAULT_DEV_USER
}

// ─── Composable ───────────────────────────────────────────────────────────────

// The Telegram SDK script always creates window.Telegram.WebApp, even in a regular
// browser. initData is only non-empty when the page is genuinely opened inside Telegram.
const inTelegram = !!(tg && tg.initData)

// Resolve identity once at module load so it's stable across components.
const _devUser = inTelegram ? null : loadDevUser()

const initData: string = inTelegram
  ? (tg!.initData ?? '')
  : `dev:${JSON.stringify(_devUser)}`

const userId: number = inTelegram
  ? (tg?.initDataUnsafe?.user?.id ?? 0)
  : (_devUser!.id)

const username: string = inTelegram
  ? (tg?.initDataUnsafe?.user?.username ?? tg?.initDataUnsafe?.user?.first_name ?? 'Странник')
  : (_devUser!.username)

export function useTelegram() {
  function ready() {
    tg?.ready()
    tg?.expand()
  }

  /**
   * Switch the local dev user and reload the page.
   * Call from browser console: useTelegramSetDevUser(222222, 'player2')
   * Only works when not running inside Telegram.
   */
  function setDevUser(id: number, uname: string) {
    if (inTelegram) {
      console.warn('setDevUser has no effect inside Telegram')
      return
    }
    const user: DevUser = { id, username: uname }
    localStorage.setItem(DEV_USER_KEY, JSON.stringify(user))
    localStorage.removeItem('egypt_room_code') // clear saved room so new user starts fresh
    location.reload()
  }

  return { initData, userId, username, ready, inTelegram, setDevUser }
}
