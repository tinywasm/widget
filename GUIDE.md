# User Guide — `tinywasm/widget`

This guide helps you choose the correct layout, styling, and structural components using `tinywasm/widget`.

## The Decision Table

The substitute for design judgement: you do not choose, you look up.

| I want… | Use |
|---|---|---|
| a column of things | `Stack(Space2)` |
| a row of buttons | `Row(Space1)` |
| a grid that adapts by itself | `Grid(ColumnNarrow, Space2)` |
| list plus detail | `Split(SplitTwoThirds, Space3)` |
| a centred column of text | `Center(Readable)` |
| a horizontal scrolling strip | `ScrollRow(Space2)` |
| an image with a fixed proportion | `MediaBox(Aspect16x9)` |
| an svg icon that keeps its box | `IconBox(IconLg)` |
| the item in a row that pushes the rest aside | `Grow()` |
| an item pinned to the trailing edge of its line | `PushEnd()` |
| the frame of a whole application | `Cover()` |
| list and detail swiped on a phone | `On(css.Mobile, "", MasterDetail(Most))` |
| a fixed nav rail beside the content | `Sidebar(SideEnd, RailNarrow, SpaceNone)` |
| a dropdown that does not push the page | `Anchor()` on the trigger + `Flyout(SideEnd)` on the list |
| a panel that slides in from an edge | `Drawer(SideEnd, TwoThirds)` + `RevealedBy(widget.Open)` |
| a different arrangement on phones | `On(css.Mobile, "part", …)` |
| an element that exists only on phones | `OnlyOn(css.Mobile, "part", …)` |
| the page background | `As(Page)` |
| a card or panel | `As(Panel)` |
| something clickable | `Interactive(Primary)` |
| something clickable, secondary | `Interactive(Secondary)` |
| the selected item of a list | `When(widget.Selected, "item", As(Highlight))` |
| secondary text | `As(Subtle)` |
| an error | `As(Danger)` |
| to fill the remaining height | `Fill()` |
| to scroll internally | `Scroll()` |
| something that expands | `RevealedBy(widget.Open)` |
| a modal dialog | `Backdrop(Viewport)` + `Veil()` |

## Principles

1. **Deterministic CSS:** What you declare in Go is exactly what is generated in CSS.
2. **Zero-value compatibility:** Component stylesheet builders should not read fields; they should work on `&T{}`.
3. **No custom property leakage:** All styling relies on a closed catalog of CSS variables and design tokens.
