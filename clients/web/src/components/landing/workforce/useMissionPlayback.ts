'use client'

import { useCallback, useEffect, useState } from 'react'
import { getNextMissionStep } from './mission-playback'

const STEP_INTERVAL_MS = 1800
const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)'

function prefersReducedMotion() {
  return typeof window !== 'undefined' && window.matchMedia(REDUCED_MOTION_QUERY).matches
}

export function useMissionPlayback(totalSteps: number) {
  const [currentStep, setCurrentStep] = useState(0)
  const [isPaused, setIsPaused] = useState(prefersReducedMotion)
  const [reducedMotion, setReducedMotion] = useState(prefersReducedMotion)

  useEffect(() => {
    const mediaQuery = window.matchMedia(REDUCED_MOTION_QUERY)
    const handleChange = (event: MediaQueryListEvent) => {
      setReducedMotion(event.matches)
      if (event.matches) setIsPaused(true)
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  useEffect(() => {
    if (isPaused || reducedMotion || totalSteps === 0) return

    const timer = window.setInterval(() => {
      setCurrentStep((step) => getNextMissionStep(step, totalSteps))
    }, STEP_INTERVAL_MS)

    return () => window.clearInterval(timer)
  }, [isPaused, reducedMotion, totalSteps])

  const pause = useCallback(() => setIsPaused(true), [])
  const togglePlayback = useCallback(() => setIsPaused((paused) => !paused), [])
  const replay = useCallback(() => {
    setCurrentStep(0)
    setIsPaused(false)
  }, [])
  const resetAndPause = useCallback(() => {
    setCurrentStep(0)
    setIsPaused(true)
  }, [])

  return { currentStep, isPaused, pause, replay, resetAndPause, togglePlayback }
}
