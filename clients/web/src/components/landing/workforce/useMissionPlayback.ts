'use client'

import { useCallback, useEffect, useState, useSyncExternalStore } from 'react'
import { getNextMissionStep } from './mission-playback'

const STEP_INTERVAL_MS = 1800
const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)'

function subscribeToReducedMotion(onStoreChange: () => void) {
  const mediaQuery = window.matchMedia(REDUCED_MOTION_QUERY)
  mediaQuery.addEventListener('change', onStoreChange)
  return () => mediaQuery.removeEventListener('change', onStoreChange)
}

function getReducedMotionSnapshot() {
  return window.matchMedia(REDUCED_MOTION_QUERY).matches
}

export function useMissionPlayback(totalSteps: number) {
  const [currentStep, setCurrentStep] = useState(0)
  const [userPaused, setUserPaused] = useState(false)
  const reducedMotion = useSyncExternalStore(
    subscribeToReducedMotion,
    getReducedMotionSnapshot,
    () => false,
  )
  const isPaused = userPaused || reducedMotion

  useEffect(() => {
    if (isPaused || reducedMotion || totalSteps === 0) return

    const timer = window.setInterval(() => {
      setCurrentStep((step) => getNextMissionStep(step, totalSteps))
    }, STEP_INTERVAL_MS)

    return () => window.clearInterval(timer)
  }, [isPaused, reducedMotion, totalSteps])

  const togglePlayback = useCallback(() => setUserPaused((paused) => !paused), [])
  const nextStep = useCallback(() => {
    setCurrentStep((step) => getNextMissionStep(step, totalSteps))
  }, [totalSteps])
  const replay = useCallback(() => {
    setCurrentStep(0)
    setUserPaused(reducedMotion)
  }, [reducedMotion])
  const resetAndPause = useCallback(() => {
    setCurrentStep(0)
    setUserPaused(true)
  }, [])

  return {
    currentStep,
    isPaused,
    reducedMotion,
    nextStep,
    replay,
    resetAndPause,
    togglePlayback,
  }
}
