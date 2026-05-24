# Continuum Brand System — Circuit Board Edition

## Core Identity

**Name:** Continuum  
**Tagline:** The Persistent Brain for Your Agent Swarm  
**Positioning:** Cross-agent memory mesh — self-hosted, open source, agent-agnostic  
**Personality:** Technical, precise, powerful, retro-futuristic

---

## Locked Color Palette

| Token | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Deep Navy** | `#0A1628` | 10, 22, 40 | Primary background, dark mode base, code blocks |
| **Electric Cyan** | `#00D4FF` | 0, 212, 255 | Primary accent, active links, CTAs, highlights |
| **Soft Lavender** | `#A78BFA` | 167, 139, 250 | Secondary accent, connections, borders, secondary text |
| **Hot Magenta** | `#FF006E` | 255, 0, 110 | Alerts, errors, warnings, attention-grabbing elements |
| **Pure White** | `#FFFFFF` | 255, 255, 255 | Primary text on dark, headings |
| **Muted Slate** | `#1A2B4A` | 26, 43, 74 | Secondary background, cards, borders, table headers |
| **Circuit Green** | `#00FF88` | 0, 255, 136 | Success states, positive indicators, status OK |
| **Dim Grey** | `#6B7280` | 107, 114, 128 | Muted text, captions, disabled states |

### Usage Rules
- **Backgrounds:** Always Deep Navy (`#0A1628`) or Muted Slate (`#1A2B4A`)
- **Text on dark:** Pure White (`#FFFFFF`) or Dim Grey (`#6B7280`) for secondary
- **Text on light:** Deep Navy (`#0A1628`)
- **Accent hierarchy:** Electric Cyan → Soft Lavender → Hot Magenta
- **Success/Error:** Circuit Green for OK, Hot Magenta for errors
- **Never use:** Orange, yellow, brown, or pastel colors

---

## Typography

| Role | Font | Weight | Size | Letter-Spacing |
|------|------|--------|------|----------------|
| **Wordmark** | JetBrains Mono / SF Mono | 700 | 24–32px | 4–6px |
| **H1** | Inter / SF Pro Display | 700 | 32–40px | -0.5px |
| **H2** | Inter / SF Pro Display | 600 | 24–28px | -0.3px |
| **H3** | Inter / SF Pro Display | 600 | 18–20px | 0px |
| **Body** | Inter / SF Pro Text | 400 | 14–16px | 0px |
| **Code** | JetBrains Mono / SF Mono | 400 | 13–14px | 0px |
| **Tagline** | JetBrains Mono / SF Mono | 400 | 11–13px | 2–3px |
| **Caption** | Inter / SF Pro Text | 400 | 12px | 0.5px |

### Typography Rules
- **Headings:** Always white on dark backgrounds
- **Taglines:** Always ALL CAPS, monospace, lavender or cyan
- **Code blocks:** Deep Navy background (`#0A1628`), Electric Cyan syntax highlights
- **Never use:** Serif fonts, script fonts, or decorative typefaces

---

## Logo System

### Primary Logo
- Circuit board pattern background
- Central processor chip with connection pins
- Traces leading to agent nodes at corners
- Monospace wordmark below

### Logo Variants
1. **Full Logo** — Icon + wordmark (horizontal)
2. **Icon Only** — Circuit chip for favicon, app icon
3. **Wordmark Only** — "CONTINUUM" in monospace for headers
4. **Inverted** — Light background variant (rarely used)

### Clear Space
- Minimum 24px padding around logo on all sides
- Never place on busy backgrounds

### Minimum Sizes
- Digital: 32px height
- Print: 0.5 inch height
- Favicon: 16×16, 32×32

---

## Visual Language

### Patterns
- **Circuit traces:** 1–2px lines, cyan or lavender, connecting nodes
- **Grid backgrounds:** Subtle 20px grid, Muted Slate (`#1A2B4A`) on Deep Navy
- **Node dots:** 6–10px circles at connection points
- **Glow effects:** Soft cyan glow on active elements (box-shadow: 0 0 20px rgba(0,212,255,0.3))

### Shapes
- **Primary:** Rectangles with 4px border-radius (chips, cards)
- **Secondary:** Circles (nodes, status indicators)
- **Accent:** Hexagons (rare, for special badges)

### Icons
- Style: Line icons, 2px stroke
- Color: Electric Cyan or Soft Lavender
- Library: Phosphor Icons or Feather Icons (customized)

---

## Voice & Tone

- **Technical but approachable** — Speak to developers, not investors
- **Confident, not arrogant** — "It just works" not "We're the best"
- **Action-oriented** — "Spin it up" not "Consider using"
- **Agent-first** — The agents are the heroes, Continuum is the sidekick
- **Retro-futuristic** — Respect the terminal era, embrace the future

### Writing Style
- Use monospace for code, commands, and technical terms
- Use ALL CAPS for taglines and section labels
- Use sentence case for body text
- Avoid em-dashes; use en-dashes or colons instead
- No exclamation marks in documentation (save for marketing)

---

## Application Guidelines

### GitHub Repository
- Header: Full logo on Deep Navy background
- README: Icon + wordmark at top, circuit accents in diagrams

### Dashboard
- Background: Deep Navy
- Cards: Muted Slate with 1px border
- Active elements: Electric Cyan glow
- Status indicators: Circuit Green (OK), Hot Magenta (Error)

### Documentation
- Background: White or Deep Navy (toggle)
- Code blocks: Deep Navy with cyan highlights
- Tables: Muted Slate headers, alternating rows

### Social Media (X, Reddit, Discord)
- Square posts: Icon centered on Deep Navy
- Banners: Full logo left-aligned, tagline right
- Thumbnails: Icon only, high contrast

### Merch / Swag
- T-shirts: Deep Navy or black, cyan logo
- Stickers: Icon only, holographic finish

---

## Markdown Styling Rules

### Headers
```markdown
# H1 — Inter 700, white, 32px
## H2 — Inter 600, white, 24px  
### H3 — Inter 600, lavender, 18px
```

### Code Blocks
- Background: `#0A1628`
- Text: `#FFFFFF`
- Keywords: `#00D4FF`
- Strings: `#A78BFA`
- Comments: `#6B7280`

### Tables
- Header background: `#1A2B4A`
- Header text: `#00D4FF`, uppercase, monospace
- Row text: `#FFFFFF`
- Border: `#1A2B4A`
- Alternate rows: subtle `#0F1D33`

### Blockquotes
- Left border: 4px `#00D4FF`
- Background: `#1A2B4A`
- Text: `#FFFFFF`

### Links
- Default: `#00D4FF`
- Hover: `#FFFFFF` with cyan underline
- Visited: `#A78BFA`

### Badges
- Background: `#1A2B4A`
- Text: `#00D4FF`
- Border: 1px `#00D4FF`
- Border-radius: 4px
- Font: monospace, 12px

---

## Assets

### Generated Files
- `logo-full.svg` — Full logo with wordmark
- `logo-icon.svg` — Icon only (circuit chip)
- `logo-wordmark.svg` — Wordmark only
- `favicon-16.png`, `favicon-32.png` — Favicon sizes
- `social-banner.png` — 1500×500 for X/Discord
- `social-square.png` — 1080×1080 for posts

### File Locations
```
branding/
├── BRAND_SYSTEM.md          # This file
├── assets/
│   ├── logo-full.svg
│   ├── logo-icon.svg
│   ├── logo-wordmark.svg
│   ├── favicon-16.png
│   ├── favicon-32.png
│   ├── social-banner.png
│   └── social-square.png
└── concepts/                # Original 10 concepts (not committed)
```

---

## Enforcement

**All future `.md` files, documentation, and UI must follow this system.**
- No exceptions for "quick fixes" or "one-off pages"
- Any deviation requires brand system update + approval
- When in doubt, refer to this document

---

*Locked: 2026-05-24*  
*Branch: brand/circuit-board-identity*  
*Concept: Circuit Board (Concept 6)*
