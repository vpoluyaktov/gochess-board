const { test, expect, devices } = require('@playwright/test');

test.describe('Responsive Design', () => {
  test('should display correctly on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Check that board is visible (using actual board ID)
    const board = page.locator('#myBoard');
    await expect(board).toBeVisible();
  });

  test('should display correctly on tablet', async ({ page }) => {
    // Set tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Check that board is visible
    const board = page.locator('#myBoard');
    await expect(board).toBeVisible();
  });

  test('should display correctly on desktop', async ({ page }) => {
    // Set desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Check that board is visible
    const board = page.locator('#myBoard');
    await expect(board).toBeVisible();
  });
});
