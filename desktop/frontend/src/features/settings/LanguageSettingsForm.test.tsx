import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import LanguageSettingsForm from './LanguageSettingsForm'

describe('LanguageSettingsForm', () => {
  it('renders locale options and save action', () => {
    const html = renderToStaticMarkup(
      <LanguageSettingsForm
        locale="en"
        busy={false}
        onLocaleChange={() => undefined}
        onSaved={() => undefined}
        onError={() => undefined}
      />,
    )

    expect(html).toContain('Language')
    expect(html).toContain('English')
    expect(html).toContain('Portuguese (Brazil)')
    expect(html).toContain('Save language')
  })
})
