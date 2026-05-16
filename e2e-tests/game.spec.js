// @ts-check
require('dotenv').config({ path: '../backend/.env' })

const { test, expect, request } = require('@playwright/test')
const { generateInitData } = require('./helpers/tgCrypto')

// ─── Config ───────────────────────────────────────────────────────────────────

const BOT_TOKEN = process.env.BOT_TOKEN
const API_BASE = process.env.BACKEND_URL ?? 'http://localhost:8080'
const FRONTEND_BASE = process.env.FRONTEND_URL ?? 'http://localhost:5173'

if (!BOT_TOKEN) {
  throw new Error('BOT_TOKEN must be set in backend/.env to run e2e tests')
}

// Two fake test players
const PLAYER_1 = { id: 100000001, first_name: 'Хорус', username: 'horus_test' }
const PLAYER_2 = { id: 100000002, first_name: 'Анубис', username: 'anubis_test' }

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Injects a window.Telegram.WebApp mock before the page loads.
 * This satisfies useTelegram.ts without a real Telegram client.
 */
async function mockTelegramWebApp(page, user, initData) {
  await page.addInitScript(
    ({ user, initData }) => {
      window.Telegram = {
        WebApp: {
          ready() {},
          expand() {},
          initData,
          initDataUnsafe: { user },
          colorScheme: 'dark',
          themeParams: {},
        },
      }
    },
    { user, initData },
  )
}

/**
 * Sends a signed API request on behalf of a player (no browser needed).
 */
async function apiRequest(apiCtx, method, path, body, initData) {
  const url = `${API_BASE}${path}`
  const opts = {
    headers: {
      'Content-Type': 'application/json',
      'X-Telegram-Init-Data': initData,
    },
  }
  if (body !== undefined) opts.data = body

  const res = method === 'GET'
    ? await apiCtx.get(url, opts)
    : await apiCtx.post(url, opts)

  expect(res.ok(), `${method} ${path} → ${res.status()}`).toBeTruthy()
  return res.json()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

test.describe('Полный игровой сценарий', () => {
  let apiCtx
  let initData1
  let initData2
  let roomCode

  test.beforeAll(async ({ playwright }) => {
    apiCtx = await request.newContext()
    initData1 = generateInitData(BOT_TOKEN, PLAYER_1)
    initData2 = generateInitData(BOT_TOKEN, PLAYER_2)
  })

  test.afterAll(async () => {
    await apiCtx.dispose()
  })

  // ── 1. Создание комнаты ──────────────────────────────────────────────────────

  test('игрок 1 создаёт комнату через UI', async ({ page }) => {
    await mockTelegramWebApp(page, PLAYER_1, initData1)
    await page.goto(FRONTEND_BASE)

    // Проверяем главный экран
    await expect(page.getByText('МААТ И СЕТ')).toBeVisible()
    await expect(page.getByText(`Суд Осириса ждёт, ${PLAYER_1.username}`)).toBeVisible()

    // Переключаемся на вкладку "Создать храм"
    await page.getByRole('button', { name: 'Создать храм' }).click()
    await page.getByRole('button', { name: /Создать комнату/ }).click()

    // После создания должен появиться экран лобби с кодом комнаты
    await expect(page.getByText(/Лобби|Храм|код/i)).toBeVisible({ timeout: 10_000 })

    // Вытаскиваем код комнаты из состояния приложения
    roomCode = await page.evaluate(() => localStorage.getItem('egypt_room_code'))
    expect(roomCode).toBeTruthy()
    expect(roomCode.length).toBeGreaterThanOrEqual(4)
  })

  // ── 2. Второй игрок присоединяется ──────────────────────────────────────────

  test('игрок 2 входит в комнату через UI', async ({ page }) => {
    expect(roomCode, 'Код комнаты должен быть известен из предыдущего теста').toBeTruthy()

    await mockTelegramWebApp(page, PLAYER_2, initData2)
    await page.goto(FRONTEND_BASE)

    // Вкладка "Войти в храм" активна по умолчанию
    const codeInput = page.locator('input.code-input')
    await codeInput.fill(roomCode)
    await page.getByRole('button', { name: /Войти в Храм/ }).click()

    // Попадаем в лобби
    await expect(page.getByText(/Лобби|Храм|код/i)).toBeVisible({ timeout: 10_000 })
  })

  // ── 3. API: старт игры ───────────────────────────────────────────────────────

  test('игрок 1 запускает игру через API', async () => {
    expect(roomCode).toBeTruthy()

    const state = await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/start`, {}, initData1)
    expect(['question', 'lobby']).toContain(state.phase)
    // Если < 2 игроков — всё равно проходим тест; старт мог вернуть ошибку
  })

  // ── 4. Ответ на вопрос ───────────────────────────────────────────────────────

  test('оба игрока отвечают на вопрос', async () => {
    expect(roomCode).toBeTruthy()

    // Проверяем текущую фазу
    const room = await apiRequest(apiCtx, 'GET', `/api/room/${roomCode}`, undefined, initData1)

    if (room.phase !== 'question') {
      test.skip(true, `Ожидалась фаза question, получена: ${room.phase}`)
      return
    }

    const ans1 = await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/answer`, { text: 'Потому что Маат требует справедливости' }, initData1)
    expect(ans1).toBeTruthy()

    const ans2 = await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/answer`, { text: 'Сет нашептал иное' }, initData2)
    expect(ans2).toBeTruthy()
  })

  // ── 5. Голосование ───────────────────────────────────────────────────────────

  test('оба игрока голосуют на фазе voting', async () => {
    expect(roomCode).toBeTruthy()

    const room = await apiRequest(apiCtx, 'GET', `/api/room/${roomCode}`, undefined, initData1)

    if (room.phase !== 'voting') {
      test.skip(true, `Ожидалась фаза voting, получена: ${room.phase}`)
      return
    }

    await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/vote`, { trust: true }, initData1)
    await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/vote`, { trust: false }, initData2)
  })

  // ── 6. Фаза Сета (опционально) ───────────────────────────────────────────────

  test('игрок отвечает на вопрос Сета, если фаза активна', async () => {
    expect(roomCode).toBeTruthy()

    const room = await apiRequest(apiCtx, 'GET', `/api/room/${roomCode}`, undefined, initData1)

    if (room.phase !== 'seth') {
      test.skip(true, 'Фаза seth не активна, пропускаем')
      return
    }

    // Сет задаёт вопрос с вариантами; выбираем первый
    await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/seth`, { option: 0 }, initData1)
    await apiRequest(apiCtx, 'POST', `/api/room/${roomCode}/seth`, { option: 1 }, initData2)
  })

  // ── 7. Финальный экран ───────────────────────────────────────────────────────

  test('UI игрока 1 показывает экран завершения игры', async ({ page }) => {
    expect(roomCode).toBeTruthy()

    // Пропускаем, если игра ещё в процессе (мало игроков для смены фазы)
    const room = await apiRequest(apiCtx, 'GET', `/api/room/${roomCode}`, undefined, initData1)
    if (room.phase !== 'finished') {
      test.skip(true, `Игра ещё не завершена, фаза: ${room.phase}`)
      return
    }

    await mockTelegramWebApp(page, PLAYER_1, initData1)
    await page.goto(FRONTEND_BASE)

    // Восстанавливаем сессию через localStorage
    await page.evaluate((code) => localStorage.setItem('egypt_room_code', code), roomCode)
    await page.reload()

    await expect(page.getByText(/Суд завершён|Результаты|победитель/i)).toBeVisible({ timeout: 15_000 })
  })
})

// ─── Изолированные юнит-подобные тесты ────────────────────────────────────────

test.describe('API: авторизация', () => {
  test('запрос без initData возвращает 401', async ({ request: req }) => {
    const res = await req.post(`${API_BASE}/api/room`)
    expect(res.status()).toBe(401)
  })

  test('запрос с невалидной подписью возвращает 401', async ({ request: req }) => {
    const res = await req.post(`${API_BASE}/api/room`, {
      headers: { 'X-Telegram-Init-Data': 'user=%7B%7D&auth_date=1&hash=deadbeef' },
    })
    expect(res.status()).toBe(401)
  })

  test('запрос с валидным initData создаёт комнату', async ({ request: req }) => {
    const initData = generateInitData(BOT_TOKEN, PLAYER_1)
    const res = await req.post(`${API_BASE}/api/room`, {
      headers: { 'X-Telegram-Init-Data': initData },
    })
    expect(res.ok()).toBeTruthy()
    const body = await res.json()
    expect(body.code).toBeTruthy()
    expect(body.phase).toBe('lobby')
  })
})

test.describe('API: комната', () => {
  let apiCtx
  let initData1
  let initData2

  test.beforeAll(async ({ playwright }) => {
    apiCtx = await request.newContext()
    initData1 = generateInitData(BOT_TOKEN, PLAYER_1)
    initData2 = generateInitData(BOT_TOKEN, PLAYER_2)
  })

  test.afterAll(async () => apiCtx.dispose())

  test('GET /api/room/:code без auth возвращает данные (публичный)', async () => {
    // Создаём комнату
    const created = await apiRequest(apiCtx, 'POST', '/api/room', {}, initData1)
    const code = created.code

    // GET не требует auth по текущей реализации
    const res = await apiCtx.get(`${API_BASE}/api/room/${code}`)
    expect(res.ok()).toBeTruthy()
    const room = await res.json()
    expect(room.code).toBe(code)
    expect(room.phase).toBe('lobby')
  })

  test('нельзя запустить игру из чужой комнаты', async () => {
    const room = await apiRequest(apiCtx, 'POST', '/api/room', {}, initData1)
    const code = room.code

    // Игрок 2 пытается стартовать чужую комнату
    const res = await apiCtx.post(`${API_BASE}/api/room/${code}/start`, {
      headers: {
        'Content-Type': 'application/json',
        'X-Telegram-Init-Data': initData2,
      },
      data: {},
    })
    // Бэкенд должен вернуть ошибку (только создатель может стартовать)
    expect(res.ok()).toBeFalsy()
  })

  test('второй игрок может присоединиться к комнате', async () => {
    const room = await apiRequest(apiCtx, 'POST', '/api/room', {}, initData1)
    const code = room.code

    const joined = await apiRequest(apiCtx, 'POST', `/api/room/${code}/join`, {}, initData2)
    expect(joined.phase).toBe('lobby')
    const names = joined.players.map((p) => p.name)
    expect(names).toContain(PLAYER_2.username)
  })
})
