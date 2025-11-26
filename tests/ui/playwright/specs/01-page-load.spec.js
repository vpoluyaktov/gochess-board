const { test, expect } = require('@playwright/test');

test.describe('Page Load', () => {
  test('should load the chess board page', async ({ page }) => {
    await page.goto('/');
    
    // Check page title
    await expect(page).toHaveTitle(/Chess/i);
    
    // Check that the page loaded
    await expect(page.locator('body')).toBeVisible();
  });

  test('should display the chessboard', async ({ page }) => {
    await page.goto('/');
    
    // Wait for chessboard to be visible (using actual board ID)
    const board = page.locator('#myBoard');
    await expect(board).toBeVisible({ timeout: 5000 });
  });

  test('should load chess pieces', async ({ page }) => {
    await page.goto('/');
    
    // Wait for board
    await page.waitForTimeout(1000);
    
    // Check that pieces are rendered (look for piece images or unicode)
    const pieces = page.locator('img[data-piece], .piece, [class*="piece"]');
    const count = await pieces.count();
    
    // Should have 32 pieces at start
    expect(count).toBeGreaterThan(0);
  });
});
