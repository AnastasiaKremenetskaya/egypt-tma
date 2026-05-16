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

export function useTelegram() {
  const initData = tg?.initData ?? ''

  const rawUser = tg?.initDataUnsafe?.user
  const userId = rawUser?.id ?? 0
  const username = rawUser?.username ?? rawUser?.first_name ?? 'Странник'

  function ready() {
    tg?.ready()
    tg?.expand()
  }

  return { initData, userId, username, ready }
}
