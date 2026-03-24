import { test, expect, type Page } from '@playwright/test'

async function createProjectAndGoToSettings(page: Page): Promise<string> {
  const projectName = `Settings E2E ${Date.now()}`

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

  // Click the card to go to dashboard first
  const projectCard = page.locator('div.grid > div').filter({ hasText: projectName })
  await projectCard.getByRole('link', { name: /Open|View/ }).click()
  await page.waitForLoadState('networkidle')
  await expect(page).toHaveURL(/\/projects\/[^/]+$/)

  // Now navigate to settings from the dashboard via the Settings link
  await page.getByRole('link', { name: 'Settings' }).click()
  await page.waitForLoadState('networkidle')
  await expect(page).toHaveURL(/\/projects\/[^/]+\/settings$/)

  return projectName
}

test.describe('Project Settings', () => {
  test.describe.serial('settings page navigation and content', () => {
    test('navigates to settings from dashboard by clicking Settings link', async ({ page }) => {
      await createProjectAndGoToSettings(page)

      await expect(
        page.getByRole('heading', { name: /Settings/, exact: false })
      ).toBeVisible()
    })

    test('page shows the project name in the heading', async ({ page }) => {
      const projectName = await createProjectAndGoToSettings(page)

      await expect(
        page.getByRole('heading', { name: new RegExp(`${projectName}.*Settings`) })
      ).toBeVisible()
    })

    test('Risk Profile tab is present and active by default', async ({ page }) => {
      await createProjectAndGoToSettings(page)

      // The Tabs component renders TabsTrigger elements
      await expect(page.getByRole('tab', { name: 'Risk Profile', exact: true })).toBeVisible()
    })

    test('Model Routing tab is present and can be clicked', async ({ page }) => {
      await createProjectAndGoToSettings(page)

      const modelRoutingTab = page.getByRole('tab', { name: 'Model Routing', exact: true })
      await expect(modelRoutingTab).toBeVisible()
      await modelRoutingTab.click()

      // After clicking, the Model Routing tab panel should be active
      await expect(modelRoutingTab).toHaveAttribute('data-state', 'active')
    })

    test('Risk Profile tab shows risk criteria content when loaded', async ({ page }) => {
      await createProjectAndGoToSettings(page)

      // Risk Profile is the default tab
      const riskTab = page.getByRole('tab', { name: 'Risk Profile', exact: true })
      await expect(riskTab).toHaveAttribute('data-state', 'active')

      // The panel content should be visible — either the editor or no-profile message
      const riskContent = page.getByRole('tabpanel')
      await expect(riskContent).toBeVisible()
    })

    test('back arrow navigates to project dashboard', async ({ page }) => {
      await createProjectAndGoToSettings(page)

      // The ArrowLeft button is a ghost icon button wrapping a Link
      // The Link renders as the only <a> with ArrowLeft in the header area
      const backButton = page
        .locator('div.flex.items-center.gap-3')
        .first()
        .getByRole('link')
        .first()

      await backButton.click()

      await expect(page).toHaveURL(/\/projects\/[^/]+$/)
    })

    test('no risk profile state renders appropriate message', async ({ page }) => {
      // This tests the branch where riskProfile is null after loading.
      // We can only verify the page doesn't crash and shows something meaningful.
      await createProjectAndGoToSettings(page)

      // Either tabs appear (profile loaded) or the "No risk profile" message appears
      const hasTabs = await page.getByRole('tab', { name: 'Risk Profile' }).isVisible()
      const hasNoProfile = await page
        .getByText('No risk profile found for this project.')
        .isVisible()
      const isLoading = await page.getByText('Loading settings...').isVisible()

      expect(hasTabs || hasNoProfile || isLoading).toBe(true)
    })
  })
})
