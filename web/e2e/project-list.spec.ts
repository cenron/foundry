import { test, expect } from '@playwright/test'

test.describe('Project List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('page loads with Foundry branding', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Foundry', exact: true })).toBeVisible()
  })

  test('page heading is "Projects"', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Projects', exact: true })).toBeVisible()
  })

  test('"New Project" button is always visible in the header', async ({ page }) => {
    // The header-level button is always present regardless of empty/populated state
    const header = page.locator('div.flex.items-center.justify-between').first()
    await expect(header.getByRole('button', { name: 'New Project' })).toBeVisible()
  })

  test('shows empty state when no projects exist', async ({ page }) => {
    const projectsResponse = page.waitForResponse((res) =>
      res.url().includes('/api/projects') && res.status() === 200
    )
    await page.goto('/')
    await projectsResponse

    const projectsList = await page.locator('[class*="grid"]').count()

    if (projectsList === 0) {
      await expect(page.getByText('No projects yet')).toBeVisible()
      await expect(
        page.getByText('Create your first project to get started.')
      ).toBeVisible()
      // Empty state also renders a "New Project" button
      const emptyStateButton = page
        .locator('div.rounded-xl')
        .getByRole('button', { name: 'New Project' })
      await expect(emptyStateButton).toBeVisible()
    }
  })

  test('"New Project" button opens create dialog', async ({ page }) => {
    const header = page.locator('div.flex.items-center.justify-between').first()
    await header.getByRole('button', { name: 'New Project' }).click()

    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(
      page.getByRole('heading', { name: 'Create Project', exact: true })
    ).toBeVisible()
    await expect(page.getByLabel('Name')).toBeVisible()
    await expect(page.getByLabel('Description')).toBeVisible()
    await expect(page.getByLabel('Repository URL')).toBeVisible()
  })

  test('create project: fills form, submits, verifies API, shows new card', async ({ page }) => {
    const projectName = `E2E Project ${Date.now()}`

    const header = page.locator('div.flex.items-center.justify-between').first()
    await header.getByRole('button', { name: 'New Project' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    await page.getByLabel('Name').fill(projectName)
    await page.getByLabel('Description').fill('Created by E2E test')

    const createResponse = page.waitForResponse(
      (res) =>
        res.url().includes('/api/projects') &&
        res.request().method() === 'POST' &&
        res.status() === 201
    )

    await page.getByRole('button', { name: 'Create Project', exact: true }).click()

    await createResponse

    // Dialog should close after success
    await expect(page.getByRole('dialog')).not.toBeVisible()

    // New card should appear in the grid
    await expect(page.getByText(projectName)).toBeVisible()
  })

  test('project card shows name and status badge', async ({ page }) => {
    // Only meaningful when at least one project exists; skip gracefully if grid is empty
    const grid = page.locator('div.grid').first()
    const cardCount = await grid.locator('[class*="CardTitle"], [class*="card"]').count()

    if (cardCount === 0) {
      test.skip()
      return
    }

    const firstCard = grid.locator('> div').first()
    // The card renders a CardTitle (div) with the project name
    await expect(firstCard.locator('div[class*="text-base"]')).not.toBeEmpty()
    // Badge with status text is present
    await expect(firstCard.locator('[class*="badge"], [class*="Badge"]')).toBeVisible()
  })

  test('clicking project card navigates to dashboard', async ({ page }) => {
    // Wait for projects to load
    await page.waitForLoadState('networkidle')

    const grid = page.locator('div.grid').first()
    const openButton = grid.getByRole('link', { name: /Open|View/ }).first()

    if (!(await openButton.isVisible())) {
      test.skip()
      return
    }

    await openButton.click()

    await expect(page).toHaveURL(/\/projects\/[^/]+$/)
  })
})
