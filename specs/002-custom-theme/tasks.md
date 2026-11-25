# Tasks — 002 custom theme

Legend: [P] parallel-safe. Executor: haiku subagent; orchestrator reviews each.

## Phase 1 (parallel)
- [x] T1 [P] Go: add `Theme map[string]string` to `pkg/config` `File` and `fileRaw` (json/yaml `theme,omitempty`), carry through load/save, test round trip. Files: pkg/config/config.go, config_test.go
- [x] T2 [P] TS: `PmConfig.theme?`, `util/theme.ts` (THEME_KEYS, PRESETS, isHex, buildThemeVars, applyTheme, contrastRatio) + `util/theme.test.ts`. Files: src/types.ts, src/util/theme.ts(+test)

## Phase 2 (parallel, after T2)
- [x] T3 [P] Wire apply: effect applying `config.theme`, localStorage cache pre-render in main.tsx. Files: src/main.tsx, src/context/ConfigContext.tsx (+test)
- [x] T4 [P] UI: ThemeSettingsForm + SettingsPage mount + pt-BR strings + CSS. Files: src/features/settings/ThemeSettingsForm.tsx(+test), src/pages/SettingsPage.tsx, src/locales/pt-BR.json, src/index.css (append only)

## Phase 3 (orchestrator)
- [x] T5 Run go test, vitest, tsc/lint, build.
- [x] T6 Browser validate Settings (desktop + narrow, light + dark), live preview, persistence, reset.
- [x] T7 Impeccable polish pass; fix findings.
