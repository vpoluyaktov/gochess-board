const { test, expect } = require('@playwright/test');

test.describe('Engine Selection', () => {
  test('should display white player engine selection dropdown', async ({ page }) => {
    await page.goto('/');
    
    // Look for white player selector
    const whitePlayerSelect = page.locator('select#whitePlayer');
    await expect(whitePlayerSelect).toBeVisible({ timeout: 5000 });
  });

  test('should display black player engine selection dropdown', async ({ page }) => {
    await page.goto('/');
    
    // Look for black player selector
    const blackPlayerSelect = page.locator('select#blackPlayer');
    await expect(blackPlayerSelect).toBeVisible({ timeout: 5000 });
  });

  test('should list available engines for white player', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Get engine options for white player
    const whitePlayerSelect = page.locator('select#whitePlayer');
    const options = whitePlayerSelect.locator('option');
    const count = await options.count();
    
    // Should have at least Human + built-in engine
    expect(count).toBeGreaterThan(1);
  });

  test('should include built-in engine in white player list', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Check for built-in engine option (GoChess Basic)
    const builtinOption = page.locator('select#whitePlayer option:has-text("GoChess")');
    await expect(builtinOption.first()).toBeAttached();
  });
});
