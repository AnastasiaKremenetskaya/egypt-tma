Техническая выжимка: «Маат и Сет»

1. Утверждённые механики
Сессия. Один игрок создаёт комнату → получает код (напр. ANKH42). Остальные входят по коду. Бот пишет в личку каждому — требует предварительного /start. Войти после старта нельзя. Создатель комнаты = админ, только он видит «Начать суд»; остальные видят «Игра начнётся когда @name начнёт её».
Ход. Игроки ходят по очереди (round-robin). Активный игрок тянет карту — Маат (личный вопрос) или Сет (интеллектуальный, с однозначным ответом). Тип карты определяется случайно или чередуется.
Ответ. 60 секунд. Два варианта: текст (виден всем) или «Ответил вслух» (бот фиксирует устный ответ). Таймаут = пропуск хода, −1 очко.
Голосование. Все кроме активного игрока голосуют: «Перо Маат» (чисто) / «Гнев Аммит» (лжёт). 30 секунд. Голоса атомарны — один игрок, один голос.
Подсчёт после голосования.
* Большинство «чисто» → активный +3.
* Большинство «ложь» → активный получает вопрос Сета (штраф).
Вопрос Сета (штраф). 20 секунд, один правильный ответ. Активный: верно → +2, неверно → 0. Остальные параллельно: верно → +1, неверно → −1.
Победа. Первый игрок с ≥ 50 очками побеждает. Проверка после каждого начисления.
Титулы. При входе каждому случайно назначается титул из фиксированного списка (31 вариант). Титул отображается рядом с ником везде в UI.

2. Архитектура файлов
cmd/
  bot/
    main.go             — точка входа, webhook/polling, регистрация хендлеров

internal/
  bot/
    handler.go          — роутинг команд (/start, /новая_комната, /войти, /старт)
                          и callback_query (vote:*, seth:*)
    middleware.go       — извлечение userID, проверка фазы комнаты
    notify.go           — рассылка сообщений группе игроков (личка + групповой чат)

  game/
    room.go             — структура Room, CRUD, список игроков, индекс хода
    fsm.go              — машина состояний: Lobby→Question→Voting→Seth→NextTurn→Finished
    turn.go             — логика одного хода: выбор карты, переход фаз
    vote.go             — AddVote(), CountVotes(), атомарность через mutex
    score.go            — начисление очков, CheckWinner()
    timer.go            — StartTimer(ctx, d, onExpire), StopTimer(), cancel-цепочка
    titles.go           — список титулов, RandomTitle(userID)

  questions/
    maat.go             — []MaatQuestion{Text string}
    seth.go             — []SethQuestion{Text, Options [4]string, CorrectIdx int}
    loader.go           — загрузка из JSON-файлов при старте

  store/
    memory.go           — map[roomCode]*Room + sync.RWMutex (MVP)
    interface.go        — Store interface (для замены на Redis без правки логики)

  transport/
    callback.go         — парсинг callback data ("vote:ANKH42:1", "seth:ANKH42:2")

config/
  config.go             — ENV: BOT_TOKEN, DEBUG, MAX_PLAYERS

data/
  questions_maat.json
  questions_seth.json

3. Форматы данных
Вопросы (JSON)
// questions_maat.json
[
  { "id": "m001", "text": "Признайся: что ты сделал и о чём стыдишься?" },
  { "id": "m002", "text": "Какую ложь ты говоришь чаще всего?" }
]

// questions_seth.json
[
  {
    "id": "s001",
    "text": "Кто взвешивает сердце в Дуате?",
    "options": ["Ра", "Озирис", "Анубис", "Тот"],
    "correct_idx": 2
  }
]
Room (Go struct / Redis hash)
{
  "code": "ANKH42",
  "admin_id": 123456,
  "phase": "voting",
  "players": [
    { "user_id": 123456, "username": "ivan", "title": "Наместник Осириса", "score": 38 },
    { "user_id": 789012, "username": "anna", "title": "Прекраснейшая как Нефертити", "score": 29 }
  ],
  "active_idx": 0,
  "current_question": { "id": "m001", "type": "maat" },
  "current_answer": { "type": "text", "text": "Однажды я взял денег в долг…" },
  "votes": { "789012": true, "345678": false },
  "phase_deadline": "2025-12-05T00:01:30Z"
}
Callback data (кнопки Telegram, ≤ 64 байта)
vote:ANKH42:1        — голос "чисто"  (1=trust, 0=lie)
seth:ANKH42:2        — ответ на вопрос Сета, вариант 2
Событие начисления очков (лог / аналитика)
{
  "room": "ANKH42",
  "round": 3,
  "event": "vote_result",
  "target_user_id": 123456,
  "delta": 3,
  "reason": "majority_trust",
  "timestamp": "2025-12-05T00:01:31Z"
}

Стек: Go 1.22+, go-telegram-bot-api, sync.Mutex для store, context для таймеров. Хранилище MVP — in-memory, интерфейс готов под Redis. Деплой — один бинарь + два JSON-файла вопросов.
