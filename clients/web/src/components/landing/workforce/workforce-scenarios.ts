export type WorkforceScenarioId =
  | 'research'
  | 'content'
  | 'operations'
  | 'sales'
  | 'knowledge'
  | 'product'

export const workforceScenarioAccents = ['mint', 'coral', 'amber', 'violet', 'sky', 'blue'] as const

export type WorkforceAccent = (typeof workforceScenarioAccents)[number]

type WorkforceScenarioRecord = {
  id: WorkforceScenarioId
  goalKey: `landing.workforce.scenarios.${WorkforceScenarioId}.goal`
  workers: readonly [
    `landing.workforce.scenarios.${WorkforceScenarioId}.workers.${string}`,
    `landing.workforce.scenarios.${WorkforceScenarioId}.workers.${string}`,
    `landing.workforce.scenarios.${WorkforceScenarioId}.workers.${string}`,
  ]
  steps: readonly [
    `landing.workforce.scenarios.${WorkforceScenarioId}.steps.${string}`,
    `landing.workforce.scenarios.${WorkforceScenarioId}.steps.${string}`,
    `landing.workforce.scenarios.${WorkforceScenarioId}.steps.${string}`,
    `landing.workforce.scenarios.${WorkforceScenarioId}.steps.${string}`,
  ]
  deliverableKey: `landing.workforce.scenarios.${WorkforceScenarioId}.deliverable`
  accent: WorkforceAccent
}

export const workforceScenarios = [
  {
    id: 'research',
    goalKey: 'landing.workforce.scenarios.research.goal',
    workers: [
      'landing.workforce.scenarios.research.workers.scout',
      'landing.workforce.scenarios.research.workers.analyst',
      'landing.workforce.scenarios.research.workers.editor',
    ],
    steps: [
      'landing.workforce.scenarios.research.steps.scope',
      'landing.workforce.scenarios.research.steps.gather',
      'landing.workforce.scenarios.research.steps.synthesize',
      'landing.workforce.scenarios.research.steps.review',
    ],
    deliverableKey: 'landing.workforce.scenarios.research.deliverable',
    accent: 'mint',
  },
  {
    id: 'content',
    goalKey: 'landing.workforce.scenarios.content.goal',
    workers: [
      'landing.workforce.scenarios.content.workers.strategist',
      'landing.workforce.scenarios.content.workers.writer',
      'landing.workforce.scenarios.content.workers.editor',
    ],
    steps: [
      'landing.workforce.scenarios.content.steps.brief',
      'landing.workforce.scenarios.content.steps.research',
      'landing.workforce.scenarios.content.steps.draft',
      'landing.workforce.scenarios.content.steps.approve',
    ],
    deliverableKey: 'landing.workforce.scenarios.content.deliverable',
    accent: 'coral',
  },
  {
    id: 'operations',
    goalKey: 'landing.workforce.scenarios.operations.goal',
    workers: [
      'landing.workforce.scenarios.operations.workers.coordinator',
      'landing.workforce.scenarios.operations.workers.analyst',
      'landing.workforce.scenarios.operations.workers.operator',
    ],
    steps: [
      'landing.workforce.scenarios.operations.steps.intake',
      'landing.workforce.scenarios.operations.steps.map',
      'landing.workforce.scenarios.operations.steps.execute',
      'landing.workforce.scenarios.operations.steps.report',
    ],
    deliverableKey: 'landing.workforce.scenarios.operations.deliverable',
    accent: 'amber',
  },
  {
    id: 'sales',
    goalKey: 'landing.workforce.scenarios.sales.goal',
    workers: [
      'landing.workforce.scenarios.sales.workers.researcher',
      'landing.workforce.scenarios.sales.workers.strategist',
      'landing.workforce.scenarios.sales.workers.writer',
    ],
    steps: [
      'landing.workforce.scenarios.sales.steps.target',
      'landing.workforce.scenarios.sales.steps.qualify',
      'landing.workforce.scenarios.sales.steps.personalize',
      'landing.workforce.scenarios.sales.steps.review',
    ],
    deliverableKey: 'landing.workforce.scenarios.sales.deliverable',
    accent: 'violet',
  },
  {
    id: 'knowledge',
    goalKey: 'landing.workforce.scenarios.knowledge.goal',
    workers: [
      'landing.workforce.scenarios.knowledge.workers.librarian',
      'landing.workforce.scenarios.knowledge.workers.analyst',
      'landing.workforce.scenarios.knowledge.workers.curator',
    ],
    steps: [
      'landing.workforce.scenarios.knowledge.steps.collect',
      'landing.workforce.scenarios.knowledge.steps.organize',
      'landing.workforce.scenarios.knowledge.steps.verify',
      'landing.workforce.scenarios.knowledge.steps.publish',
    ],
    deliverableKey: 'landing.workforce.scenarios.knowledge.deliverable',
    accent: 'sky',
  },
  {
    id: 'product',
    goalKey: 'landing.workforce.scenarios.product.goal',
    workers: [
      'landing.workforce.scenarios.product.workers.researcher',
      'landing.workforce.scenarios.product.workers.designer',
      'landing.workforce.scenarios.product.workers.builder',
    ],
    steps: [
      'landing.workforce.scenarios.product.steps.discover',
      'landing.workforce.scenarios.product.steps.define',
      'landing.workforce.scenarios.product.steps.build',
      'landing.workforce.scenarios.product.steps.validate',
    ],
    deliverableKey: 'landing.workforce.scenarios.product.deliverable',
    accent: 'blue',
  },
] as const satisfies readonly WorkforceScenarioRecord[]

export type WorkforceScenario = (typeof workforceScenarios)[number]

export type WorkforceTranslationKey =
  | WorkforceScenario['goalKey']
  | WorkforceScenario['workers'][number]
  | WorkforceScenario['steps'][number]
  | WorkforceScenario['deliverableKey']
