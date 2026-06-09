import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerBalanceCard from './PlannerBalanceCard'

describe('PlannerBalanceCard', () => {
  it('shows debt repayment, multiplier, and health cap', () => {
    const html = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={{ employeeId: '42', balanceSecs: -8 * 60 * 60 }}
        adjustment={{
          balanceSecs: -8 * 60 * 60,
          appliesToday: true,
          targetAdjustmentSecs: 3 * 60 * 60,
          adjustedTargetSecs: 11.5 * 60 * 60,
          estimatedBalanceChangeSecs: 4.5 * 60 * 60,
          remainingBalanceSecs: -3.5 * 60 * 60,
          multiplier: 1.5,
          capped: true,
        }}
        busy={false}
        error={null}
        onRefresh={() => undefined}
      />,
    )

    expect(html).toContain('-08:00:00')
    expect(html).toContain('1.5x')
    expect(html).toContain('Limite saudável aplicado')
  })

  it('shows unavailable and non-today states', () => {
    const unavailable = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={null}
        busy={false}
        error="API indisponível"
        onRefresh={() => undefined}
      />,
    )
    expect(unavailable).toContain('Saldo indisponível')
    expect(unavailable).toContain('API indisponível')

    const informational = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={{ employeeId: '42', balanceSecs: 60 * 60 }}
        adjustment={{
          balanceSecs: 60 * 60,
          appliesToday: false,
          targetAdjustmentSecs: 0,
          adjustedTargetSecs: 8.5 * 60 * 60,
          estimatedBalanceChangeSecs: 0,
          remainingBalanceSecs: 60 * 60,
          multiplier: 1.5,
          capped: false,
        }}
        busy={false}
        error={null}
        onRefresh={() => undefined}
      />,
    )
    expect(informational).toContain('apenas informativo')
  })

  it('shows loading, positive-credit, and zero-balance states', () => {
    const loading = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={null}
        busy
        error={null}
        onRefresh={() => undefined}
      />,
    )
    expect(loading).toContain('Carregando saldo atual')
    expect(loading).toContain('Atualizando...')

    const positive = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={{ employeeId: '42', balanceSecs: 60 * 60 }}
        adjustment={{
          balanceSecs: 60 * 60,
          appliesToday: true,
          targetAdjustmentSecs: -60 * 60,
          adjustedTargetSecs: 7.5 * 60 * 60,
          estimatedBalanceChangeSecs: -60 * 60,
          remainingBalanceSecs: 0,
          multiplier: 1.5,
          capped: false,
        }}
        busy={false}
        error={null}
        onRefresh={() => undefined}
      />,
    )
    expect(positive).toContain('Você tem horas disponíveis')
    expect(positive).toContain('encerrar mais cedo')

    const zero = renderToStaticMarkup(
      <PlannerBalanceCard
        balance={{ employeeId: '42', balanceSecs: 0 }}
        busy={false}
        error={null}
        onRefresh={() => undefined}
      />,
    )
    expect(zero).toContain('Banco de horas zerado')
  })
})
