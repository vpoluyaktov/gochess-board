const { test, expect } = require('@playwright/test');

test.describe('Position Analysis', () => {
  test('should have analysis toggle button', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for analysis button
    const analysisButton = page.locator('button#analysisToggle');
    await expect(analysisButton).toBeVisible();
  });

  test('should have analysis engine selector', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for analysis engine dropdown
    const analysisEngine = page.locator('select#analysisEngine');
    await expect(analysisEngine).toBeVisible();
  });

  test('should have analysis depth display', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for analysis depth indicator
    const analysisDepth = page.locator('#analysisDepth');
    await expect(analysisDepth).toBeVisible();
  });

  test('should have analysis display options', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for PV arrows radio button
    const pvOption = page.locator('input#showPVArrows');
    await expect(pvOption).toBeAttached();
    
    // Look for best moves radio button
    const bestMovesOption = page.locator('input#showBestMoves');
    await expect(bestMovesOption).toBeAttached();
  });
});
