// @ts-check
require('dotenv').config({ path: '../backend/.env' })

const { defineConfig, devices } = require('@playwright/test')

const FRONTEND_URL = process.env.FRONTEND_URL ?? 'http://localhost:5173'

module.exports = defineConfig({
  testDir: '.',
  testMatch: '**/*.spec.js',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false, // game tests share backend state — run sequentially
  retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],

  // Auto-start the Vite dev server before tests. The backend must be started
  // manually: cd backend && go run ./cmd/bot/
  webServer: {
    command: 'npm run dev',
    cwd: '../frontend',
    url: FRONTEND_URL,
    reuseExistingServer: true,
    timeout: 30_000,
  },

  use: {
    baseURL: FRONTEND_URL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 7'] },
    },
  ],
})
