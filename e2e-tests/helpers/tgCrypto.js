/**
 * Generates a valid Telegram Mini App initData string signed with HMAC-SHA256.
 *
 * Algorithm mirrors backend auth.go: ValidateInitData.
 * secret_key = HMAC-SHA256("WebAppData", botToken)
 * hash       = HMAC-SHA256(checkString,   secret_key)
 * checkString = sorted "key=value" pairs (excluding hash) joined by \n
 */

const crypto = require('crypto')

/**
 * @param {string} botToken  - Telegram bot token (must match the running backend)
 * @param {{ id: number, first_name: string, username?: string }} user
 * @param {number} [authDate] - Unix timestamp, defaults to now
 * @returns {string}  URL-encoded initData ready for X-Telegram-Init-Data header
 */
function generateInitData(botToken, user, authDate = Math.floor(Date.now() / 1000)) {
  const params = {
    auth_date: String(authDate),
    user: JSON.stringify(user),
  }

  const checkString = Object.entries(params)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([k, v]) => `${k}=${v}`)
    .join('\n')

  const secretKey = crypto.createHmac('sha256', 'WebAppData')
    .update(botToken)
    .digest()

  const hash = crypto.createHmac('sha256', secretKey)
    .update(checkString)
    .digest('hex')

  return new URLSearchParams({ ...params, hash }).toString()
}

module.exports = { generateInitData }
