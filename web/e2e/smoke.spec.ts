import { test, expect } from '@playwright/test'

test.describe('Smoke Tests', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[type="password"]')).toBeVisible()
  })

  test('login with wrong password shows error', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')
    // Should show error message
    await expect(page.locator('[role="alert"], .field-error, .text-destructive')).toBeVisible()
  })

  test('login with correct password redirects to posts', async ({ page }) => {
    // This test requires the server to be running with a known password
    // Skip if no server is available
    const baseURL = process.env.BASE_URL || 'http://localhost:8080/studio/'
    try {
      const response = await page.goto(baseURL + 'api/session')
      if (!response || response.status() >= 500) {
        test.skip()
        return
      }
    } catch {
      test.skip()
      return
    }

    await page.goto('/login')
    const password = process.env.TEST_PASSWORD || 'testpassword'
    await page.fill('input[type="password"]', password)
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL(/\/posts/)
  })

  test('health page loads for authenticated user', async ({ page }) => {
    // Skip if no server
    try {
      const response = await page.goto((process.env.BASE_URL || 'http://localhost:8080/studio/') + 'api/session')
      if (!response || response.status() >= 500) {
        test.skip()
        return
      }
    } catch {
      test.skip()
      return
    }

    // Login first
    await page.goto('/login')
    const password = process.env.TEST_PASSWORD || 'testpassword'
    await page.fill('input[type="password"]', password)
    await page.click('button[type="submit"]')

    // Navigate to health
    await page.goto('/health')
    await expect(page.locator('text=健康, text=Health')).toBeVisible()
  })
})
