import { test, expect, type Page } from '@playwright/test'

// Shared helper: create a project and navigate to its dashboard by clicking through the UI.
// Returns the project name used.
async function createAndOpenProject(page: Page): Promise<string> {
  const projectName = `Dashboard E2E ${Date.now()}`

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

  // Find the link that corresponds to the newly created project card
  const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
  const openLink = projectCard.getByRole('link', { name: /Open|View/ })
  await openLink.click()

  await page.waitForLoadState('networkidle')
  await expect(page).toHaveURL(/\/projects\/[^/]+$/)

  return projectName
}

test.describe('Project Dashboard', () => {
  test.describe.serial('dashboard with a real project', () => {
    let projectName: string

    test('navigate from project list to dashboard by clicking the card', async ({ page }) => {
      projectName = await createAndOpenProject(page)
      await expect(page.getByRole('heading', { name: projectName })).toBeVisible()
    })

    test('dashboard shows project name and status badge', async ({ page }) => {
      projectName = await createAndOpenProject(page)

      await expect(page.getByRole('heading', { name: projectName })).toBeVisible()

      // Status badge — draft is the default for new projects
      const badge = page.locator('main').getByText(/draft|planning|active|paused|completed/i).first()
      await expect(badge).toBeVisible()
    })

    test('kanban board renders with all four column headers', async ({ page }) => {
      await createAndOpenProject(page)

      // KanbanColumn renders an h3 for each column title
      await expect(page.getByRole('heading', { name: 'Backlog', exact: true })).toBeVisible()
      await expect(page.getByRole('heading', { name: 'In Progress', exact: true })).toBeVisible()
      await expect(page.getByRole('heading', { name: 'Review', exact: true })).toBeVisible()
      await expect(page.getByRole('heading', { name: 'Done', exact: true })).toBeVisible()
    })

    test('agent status bar is rendered (empty state acceptable)', async ({ page }) => {
      await createAndOpenProject(page)

      // Either the "No agents running." message or individual agent chips are present
      const noAgents = page.getByText('No agents running.')
      const agentChip = page.locator('div.flex.flex-wrap.gap-2 > div').first()

      const hasNoAgents = await noAgents.isVisible()
      const hasAgents = await agentChip.isVisible()

      expect(hasNoAgents || hasAgents).toBe(true)
    })

    test('"Spec" button opens the spec sheet', async ({ page }) => {
      await createAndOpenProject(page)

      await page.getByRole('button', { name: /Spec/i }).click()

      // Sheet opens with "Project Spec" as its title
      await expect(page.getByRole('heading', { name: 'Project Spec', exact: true })).toBeVisible()
    })

    test('"Tokens" button opens the token dashboard sheet', async ({ page }) => {
      await createAndOpenProject(page)

      await page.getByRole('button', { name: /Tokens/i }).click()

      await expect(page.getByRole('heading', { name: 'Token Usage', exact: true })).toBeVisible()
    })

    test('"PO Chat" button opens the PO chat sheet', async ({ page }) => {
      await createAndOpenProject(page)

      await page.getByRole('button', { name: /PO Chat/i }).click()

      await expect(page.getByRole('heading', { name: 'PO Chat', exact: true })).toBeVisible()
      await expect(page.getByText('Coming Soon')).toBeVisible()
    })

    test('settings icon link navigates to settings page', async ({ page }) => {
      await createAndOpenProject(page)

      // Settings is a ghost icon button wrapping a Link — click it
      const currentUrl = page.url()
      await page.getByRole('link', { name: 'Settings' }).click()

      await expect(page).toHaveURL(/\/projects\/[^/]+\/settings$/)

      // Navigate back for cleanup
      await page.goBack()
      await expect(page).toHaveURL(new RegExp(currentUrl))
    })

    test('"Start" button is present for a draft project', async ({ page }) => {
      await createAndOpenProject(page)

      // New projects start as draft, so the Start button should be visible
      await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible()
    })
  })
})
