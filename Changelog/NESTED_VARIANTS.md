# Nested Variants Support

The chess UI now supports **nested variants** - variants within variants, allowing for deep analysis of alternative lines.

## How It Works

### Creating Nested Variants

1. **Open a variant** from the main window
   - Click "Open Variation" button when viewing a move with variants
   - Variant window opens showing the variant line

2. **Create a sub-variant** in the variant window
   - Navigate to any move in the variant
   - Click "Create Variant" button
   - A new variant window opens (sub-variant window)

3. **Merge back** the sub-variant
   - Make moves in the sub-variant window
   - Click "Merge Variant" to merge back to parent variant window
   - Parent variant window now has the sub-variant

4. **Merge to main** window
   - In the parent variant window, click "Merge Variant"
   - The variant (with all sub-variants) merges to main window

## Example Hierarchy

```
Main Window
├── Move 14. Kh3
│   └── Variant: 14. Kxf4 Qe4+ 15. Kg3 Rg8+ ...
│       └── Sub-variant at move 15: 15. Kf3 (alternative to 15. Kg3)
│           └── Sub-sub-variant at move 16: 16. Qe2 (alternative within 15. Kf3 line)
```

## Technical Implementation

### Window Types
- **Main Window**: Original game window, can open variants
- **Variant Window**: Shows a variant line, can open sub-variants
- **Sub-variant Window**: Variant within a variant, can open more sub-variants

### Message Passing
All windows use `window.postMessage()` to communicate:

1. **start-variant**: Parent → Child
   - Sends game state and variant moves
   - Includes `variantIndex` to track which variant

2. **merge-variant**: Child → Parent
   - Sends modified variant moves
   - Includes `variants` object with any sub-variants
   - Includes `variantIndex` to update correct variant

### Position Adjustment
When merging sub-variants, positions are adjusted:
- Sub-variant positions are relative to variant window (0-based)
- When merging, positions are adjusted by adding `variantStartPosition`
- Example: Sub-variant at position 5 in variant window → position 31 in main window

### Data Structure

```javascript
gameState.variants = {
    26: [  // Position 26 (move 14 White)
        ['e2e4', 'e7e5', ...],  // Variant 0
        ['e2e3', 'e7e6', ...]   // Variant 1
    ],
    31: [  // Position 31 (move 16 White in variant)
        ['f3f4', 'g8g6', ...]   // Sub-variant
    ]
}
```

## User Workflow

### Scenario: Analyzing a complex position

1. **Main line**: 14. Kh3 Rxh4+ 15. Kxh4 Ne4+ ...

2. **Create variant**: What if 14. Kxf4 instead?
   - Open variant window
   - Play out: 14. Kxf4 Qe4+ 15. Kg3 Rg8+ ...

3. **Analyze sub-line**: What if 15. Kf3 instead of 15. Kg3?
   - In variant window, create sub-variant at move 15
   - Play out: 15. Kf3 Qe3+ 16. Kg4 Qg5# ...

4. **Merge everything back**:
   - Merge sub-variant to variant window
   - Merge variant (with sub-variant) to main window
   - Now main window has complete analysis tree

## Display in PGN

```
14. Kh3    Rxh4+
   └─ (14. Kxf4  Qe4+
       15. Kg3   Rg8+
          └─ (15. Kf3  Qe3+
              16. Kg4  Qg5#)
       16. Qg4   Rxg4+
       17. Kh2   Qxg2#)
15. Kxh4   Ne4+
```

## Benefits

1. **Deep Analysis**: Explore multiple levels of alternatives
2. **Organized**: Each variant window focuses on one line
3. **Flexible**: Create, modify, and merge variants at any level
4. **Preserves Work**: All sub-variants are preserved when merging

## Limitations

- Currently always opens the first variant (index 0)
- Could be extended to choose which variant to open
- Could add UI to show variant depth/hierarchy
- No limit on nesting depth (be careful with too many windows!)

## Future Enhancements

- Variant selector: Choose which variant to open when multiple exist
- Breadcrumb navigation: Show variant path (Main → Var1 → SubVar2)
- Variant tree view: Visual representation of all variants
- Variant comparison: Side-by-side comparison of variants
