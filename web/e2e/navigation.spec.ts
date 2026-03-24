import { test, expect, type Page } from '@playwright/test'

// Creates a project and returns its name. Used to ensure at least one project
// exists before navigation tests that depend on a populated list.
async function ensureProject(page: Page): Promise<string> {
  const projectName = `Nav E2E ${Date.now()}`

  await page.goto('/')
  await page.waitForLoadState('networkidle')

  const header = page.locator('div.flex.items-center.justify-between').first()
  await header.getByRole('button', { name: 'New Project' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()

  await page.getByLabel('Name').fill(projectName)

  const createResponse = page.waitForResponse(
    (res) =>
      res.url().includes('/api/projects') &&
      res.request().method() === 'POST' &&
      res.status() === 201
  )

  await page.getByRole('button', { name: 'Create Project', exact: true }).click()
  await createResponse
  await expect(page.getByRole('dialog')).not.toBeVisible()

  return projectName
}

test.describe('Navigation', () => {
  test('start at "/" renders the project list', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL('/')
    await expect(page.getByRole('heading', { name: 'Projects', exact: true })).toBeVisible()
  })

  test('clicking a project card navigates to dashboard URL', async ({ page }) => {
    const projectName = await ensureProject(page)

    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()

    await expect(page).toHaveURL(/\/projects\/[^/]+$/)
    await page.waitForLoadState('networkidle')
  })

  test('clicking Settings in nav bar navigates to settings URL', async ({ page }) => {
    const projectName = await ensureProject(page)

    // Go to dashboard first by clicking the card
    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/projects\/[^/]+$/)

    // The Layout nav bar shows a "Settings" link when on a project page
    const navBar = page.locator('nav')
    await navBar.getByRole('link', { name: 'Settings', exact: true }).click()

    await expect(page).toHaveURL(/\/projects\/[^/]+\/settings$/)
    await page.waitForLoadState('networkidle')
  })

  test('clicking back arrow on settings returns to dashboard', async ({ page }) => {
    const projectName = await ensureProject(page)

    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')

    const navBar = page.locator('nav')
    await navBar.getByRole('link', { name: 'Settings', exact: true }).click()
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/projects\/[^/]+\/settings$/)

    // ArrowLeft back button in settings page header
    const backLink = page
      .locator('main div.flex.items-center.gap-3')
      .first()
      .getByRole('link')
      .first()
    await backLink.click()

    await expect(page).toHaveURL(/\/projects\/[^/]+$/)
  })

  test('clicking the Foundry logo returns to project list', async ({ page }) => {
    const projectName = await ensureProject(page)

    // Go to dashboard first
    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/projects\/[^/]+$/)

    // Foundry logo link is in the nav
    await page.getByRole('link', { name: 'Foundry', exact: true }).click()

    await expect(page).toHaveURL('/')
    await expect(page.getByRole('heading', { name: 'Projects', exact: true })).toBeVisible()
  })

  test('clicking Projects nav link from dashboard returns to project list', async ({ page }) => {
    const projectName = await ensureProject(page)

    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/projects\/[^/]+$/)

    await page.locator('nav').getByRole('link', { name: 'Projects', exact: true }).click()

    await expect(page).toHaveURL('/')
    await expect(page.getByRole('heading', { name: 'Projects', exact: true })).toBeVisible()
  })

  test('clicking agent chip navigates to agent detail URL', async ({ page }) => {
    const projectName = await ensureProject(page)

    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')

    // Agent chips only appear when agents are running
    const agentLink = page
      .locator('div.flex.flex-wrap.gap-2 a[href*="/agents/"]')
      .first()

    const hasAgent = await agentLink.isVisible()
    if (!hasAgent) {
      // No agents running for this project — skip the agent navigation assertion
      test.skip()
      return
    }

    await agentLink.click()
    await expect(page).toHaveURL(/\/projects\/[^/]+\/agents\/[^/]+$/)
    await page.waitForLoadState('networkidle')
  })

  test('agent detail back arrow returns to dashboard', async ({ page }) => {
    const projectName = await ensureProject(page)

    const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
    await projectCard.getByRole('link', { name: /Open|View/ }).click()
    await page.waitForLoadState('networkidle')

    const agentLink = page
      .locator('div.flex.flex-wrap.gap-2 a[href*="/agents/"]')
      .first()

    const hasAgent = await agentLink.isVisible()
    if (!hasAgent) {
      test.skip()
      return
    }

    await agentLink.click()
    await expect(page).toHaveURL(/\/projects\/[^/]+\/agents\/[^/]+$/)
    await page.waitForLoadState('networkidle')

    // ArrowLeft back button in agent detail header
    const backLink = page
      .locator('div.flex.items-center.gap-3')
      .first()
      .getByRole('link')
      .first()
    await backLink.click()

    await expect(page).toHaveURL(/\/projects\/[^/]+$/)
  })
})
