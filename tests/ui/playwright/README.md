# Go-Chess UI Tests

Playwright-based end-to-end tests for the go-chess web interface.

## Setup

```bash
# Run the setup script
../scripts/test_ui.sh --setup

# Or manually:
cd tests/ui
npm install
npx playwright install chromium
```

## Running Tests

```bash
# Run all tests (headless)
npm test

# Run with browser visible
npm run test:headed

# Run in debug mode
npm run test:debug

# Run with UI mode (interactive)
npm run test:ui

# View test report
npm run report
```

## Test Structure

- `specs/01-page-load.spec.js` - Basic page loading and rendering
- `specs/02-engine-selection.spec.js` - Engine selection dropdown
- `specs/03-making-moves.spec.js` - User move interactions
- `specs/04-computer-moves.spec.js` - Computer move requests
- `specs/05-opening-names.spec.js` - Opening name recognition
- `specs/06-analysis.spec.js` - Position analysis features
- `specs/07-responsive.spec.js` - Responsive design testing

## Writing Tests

Tests use Playwright's test framework. Example:

```javascript
const { test, expect } = require('@playwright/test');

test('should do something', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('#element')).toBeVisible();
});
```

## Selectors

Update selectors in test files to match your actual HTML structure:
- Board: `.board-b72b1`, `#board`, `.chessboard`
- Squares: `[data-square="e2"]`, `.square-e2`
- Buttons: `button:has-text("Computer")`, `#computerMove`
- Engine select: `select#engine`, `select[name="engine"]`

## CI/CD Integration

Set `CI=true` environment variable for CI mode:
```bash
CI=true npm test
```
