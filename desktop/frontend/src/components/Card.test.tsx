import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { Card } from './ui'

describe('Card', () => {
  it('renders a plain section by default', () => {
    const html = renderToStaticMarkup(<Card title="Atualizações">conteúdo</Card>)

    expect(html).toContain('<section')
    expect(html).not.toContain('<details')
    expect(html).toContain('conteúdo')
  })

  it('folds the content behind the title when collapsible, closed by default', () => {
    const html = renderToStaticMarkup(
      <Card collapsible title="Planner" intro="Introdução">
        conteúdo
      </Card>,
    )

    expect(html).toContain('<details')
    expect(html).not.toMatch(/<details[^>]*\sopen/)
    expect(html).toContain('<summary')
    expect(html).toContain('Planner')
    expect(html).toContain('Introdução')
    expect(html).toContain('conteúdo')
  })

  it('can start open', () => {
    const html = renderToStaticMarkup(
      <Card collapsible defaultOpen title="Planner">
        conteúdo
      </Card>,
    )

    expect(html).toMatch(/<details[^>]*\sopen/)
  })
})
