import { test, expect } from '@playwright/test'
import { getMockWailsInitScript } from './fixtures/mock-wails'

// window.go モックを全テストで注入（addInitScript はページ読み込み前に実行される）
test.beforeEach(async ({ page }) => {
  await page.addInitScript(getMockWailsInitScript())
})

test.describe('VRChat Tweaker', () => {
  test('shows dashboard', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1')).toContainText('ダッシュボード')
  })

  test('navigates to launcher', async ({ page }) => {
    await page.goto('/')
    await page.click('text=ランチャー')
    await expect(page.locator('h1')).toContainText('ランチャー')
  })

  test('navigates to settings', async ({ page }) => {
    await page.goto('/')
    await page.click('text=設定')
    await expect(page.locator('h1')).toContainText('設定')
  })

  test.describe('Dashboard', () => {
    test('displays default profile and status buttons', async ({ page }) => {
      await page.goto('/')
      await expect(page.locator('h1')).toContainText('ダッシュボード')
      // モックでデフォルトプロファイル1件を返すため、起動ボタンが有効になる
      await expect(page.getByRole('button', { name: /VRChat 起動/ })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Join Me' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Ask Me' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Busy' })).toBeVisible()
    })

    test('can click status button', async ({ page }) => {
      await page.goto('/')
      await page.getByRole('button', { name: 'Ask Me' }).click()
      // クリック後も画面が正常であることを確認
      await expect(page.locator('h1')).toContainText('ダッシュボード')
    })
  })

  test.describe('Launcher', () => {
    test('displays profile list and editor', async ({ page }) => {
      await page.goto('/launcher')
      await expect(page.locator('h1')).toContainText('ランチャー')
      // モックのデフォルトプロファイルが表示される
      await expect(page.getByText('デフォルトプロファイル')).toBeVisible()
      await expect(page.getByText('既定')).toBeVisible()
    })

    test('can add new profile', async ({ page }) => {
      await page.goto('/launcher')
      await page.getByRole('button', { name: '+ 新規プロファイル' }).click()
      await expect(page.locator('input[type="text"]').first()).toHaveValue('新しいプロファイル')
    })

    test('can edit profile and click save', async ({ page }) => {
      await page.goto('/launcher')
      await page.getByText('デフォルトプロファイル').click()
      const nameInput = page.locator('.profile-editor input[type="text"]').first()
      await nameInput.fill('編集したプロファイル')
      await expect(nameInput).toHaveValue('編集したプロファイル')
      await page.getByRole('button', { name: '保存' }).click()
      // 保存後も画面が正常であることを確認（モックは固定データのためリストは変わらない）
      await expect(page.locator('h1')).toContainText('ランチャー')
    })
  })

  test.describe('Settings', () => {
    test('displays path settings and log retention', async ({ page }) => {
      await page.goto('/settings')
      await expect(page.locator('h1')).toContainText('設定')
      await expect(page.getByText('VRChat ログイン')).toBeVisible()
      await expect(page.getByText('パス設定')).toBeVisible()
      await expect(page.getByText('ログ・データ管理')).toBeVisible()
      await expect(page.getByText('遭遇記録の保存期間')).toBeVisible()
    })

    test('log retention input has default value', async ({ page }) => {
      await page.goto('/settings')
      const retentionInput = page.locator('input[type="number"]').first()
      await expect(retentionInput).toHaveValue('30')
    })
  })
})
