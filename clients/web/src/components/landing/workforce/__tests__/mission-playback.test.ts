import { describe, expect, expectTypeOf, it } from 'vitest'
import {
  workforceScenarioAccents,
  type WorkforceAccent,
  type WorkforceScenario,
  type WorkforceTranslationKey,
} from '../workforce-scenarios'
import { getNextMissionStep } from '../mission-playback'

describe('workforce scenarios', () => {
  it('exposes only supported accent values', () => {
    expect(workforceScenarioAccents).toEqual(['mint', 'coral', 'amber', 'violet', 'sky', 'blue'])
    expectTypeOf<WorkforceScenario['accent']>().toEqualTypeOf<WorkforceAccent>()
    expectTypeOf<WorkforceScenario['goalKey']>().toMatchTypeOf<WorkforceTranslationKey>()
    expectTypeOf<'landing.workforce.scenarios.research.unknown'>().not.toMatchTypeOf<WorkforceTranslationKey>()
  })
})

describe('getNextMissionStep', () => {
  it('advances to the next step', () => {
    expect(getNextMissionStep(0, 4)).toBe(1)
  })

  it('wraps after the final step', () => {
    expect(getNextMissionStep(3, 4)).toBe(0)
  })

  it('stays at zero when there are no steps', () => {
    expect(getNextMissionStep(2, 0)).toBe(0)
  })
})
