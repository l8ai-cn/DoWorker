import { describe, expect, it } from 'vitest'
import { getNextMissionStep } from '../mission-playback'

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
