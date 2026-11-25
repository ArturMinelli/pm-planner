# 003 — Extensible theme presets via config YAML

## Config shape (~/.config/pm/config.yaml, also JSON over the API)
```yaml
default_theme: midnight      # preset id used when the user has no saved choice
theme_name: midnight         # (managed by the app) currently selected preset id; "auto" = follow OS
theme:                       # (managed by the app) active colors; overrides everything when non-empty
  accent: "#ff8800"
themes:                      # user presets; ids are slugs [a-z0-9-_]
  midnight:
    label: Meia-noite
    colors: {bg: "#000000", panel: "#0a0a0a", text: "#f5f5f5", muted: "#a0a0a0", border: "#262626", accent: "#7c5cff", ok: "#4ade80", warning: "#facc15", danger: "#f87171"}
```
Keys: bg panel text muted border accent ok warning danger. Missing keys in a preset are filled from the built-in Claro (light) preset; invalid hex values are dropped (never fatal).

## Resolution order (effective colors)
1. `theme` non-empty → those colors (user's live edits/custom).
2. else `theme_name` (built-in id, user preset id, or `auto`) → preset colors / none.
3. else `default_theme` → preset colors.
4. else none (Automático: OS light/dark).
A user preset with the same id as a built-in overrides it. Built-ins: auto, light, dark, ocean, warm.

## UI (Aparência card)
- Preset row lists built-ins + user presets (user ones flagged), picking one sets `theme_name` and clears `theme`.
- Editing a color creates a custom `theme` (no preset highlighted).
- "Salvar como tema": name input → writes `themes[slug] = {label, colors}` and selects it; custom presets get a delete (×) button. Built-ins and yaml-defined presets are all editable/deletable the same way (they live in the same map).
- "Restaurar padrão" → clears `theme`/`theme_name` (falls to `default_theme`, else Automático).

## Docs
README section "Temas personalizados" with the yaml example and the resolution order (pt-BR).
