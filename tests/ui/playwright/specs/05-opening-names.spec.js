const { test, expect } = require('@playwright/test');

test.describe('Opening Names', () => {
  test('should have opening display element', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for opening name display (it's hidden by default)
    const openingDisplay = page.locator('#openingDisplay');
    await expect(openingDisplay).toBeAttached();
  });

  test('should have opening text element', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for opening text span
    const openingText = page.locator('#openingText');
    await expect(openingText).toBeAttached();
  });
});
