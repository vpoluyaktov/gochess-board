const { test, expect } = require('@playwright/test');

test.describe('Computer Moves', () => {
  test('should allow selecting an engine for white player', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Select an engine for white player
    const whitePlayerSelect = page.locator('select#whitePlayer');
    await expect(whitePlayerSelect).toBeVisible();
    
    // Get the first non-human option (built-in engine)
    const options = await whitePlayerSelect.locator('option').all();
    let engineValue = null;
    for (const option of options) {
      const value = await option.getAttribute('value');
      if (value !== 'human') {
        engineValue = value;
        break;
      }
    }
    
    // Select the engine
    if (engineValue) {
      await whitePlayerSelect.selectOption(engineValue);
      
      // Verify selection
      const selectedValue = await whitePlayerSelect.inputValue();
      expect(selectedValue).toBe(engineValue);
    }
  });

  test('should allow selecting an engine for black player', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Select an engine for black player
    const blackPlayerSelect = page.locator('select#blackPlayer');
    await expect(blackPlayerSelect).toBeVisible();
    
    // Get the first non-human option (built-in engine)
    const options = await blackPlayerSelect.locator('option').all();
    let engineValue = null;
    for (const option of options) {
      const value = await option.getAttribute('value');
      if (value !== 'human') {
        engineValue = value;
        break;
      }
    }
    
    // Select the engine
    if (engineValue) {
      await blackPlayerSelect.selectOption(engineValue);
      
      // Verify selection
      const selectedValue = await blackPlayerSelect.inputValue();
      expect(selectedValue).toBe(engineValue);
    }
  });

  test('should have start game button', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Look for start/pause button
    const startButton = page.locator('button#startPauseBtn');
    await expect(startButton).toBeVisible();
  });
});
