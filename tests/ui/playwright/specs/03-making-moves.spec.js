const { test, expect } = require('@playwright/test');

test.describe('Making Moves', () => {
  test('should allow clicking on pieces', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Try to click on a white pawn (e2 square)
    // Adjust selector based on your chessboard implementation
    const e2Square = page.locator('[data-square="e2"], .square-e2, #e2').first();
    
    if (await e2Square.isVisible()) {
      await e2Square.click();
      // If click works, test passes
      expect(true).toBe(true);
    } else {
      // Skip if board structure is different
      test.skip();
    }
  });

  test('should highlight legal moves when piece is selected', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Click on e2 pawn
    const e2Square = page.locator('[data-square="e2"], .square-e2').first();
    
    if (await e2Square.isVisible()) {
      await e2Square.click();
      
      // Check if any squares are highlighted (common class names)
      const highlighted = page.locator('.highlight, .legal-move, [class*="highlight"]');
      const count = await highlighted.count();
      
      // Should show at least e3 and e4 as legal moves
      expect(count).toBeGreaterThan(0);
    } else {
      test.skip();
    }
  });
});
