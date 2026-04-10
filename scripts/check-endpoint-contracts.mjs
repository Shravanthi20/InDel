import fs from 'fs'
import path from 'path'

const root = process.cwd()

function read(relPath) {
  return fs.readFileSync(path.join(root, relPath), 'utf8')
}

function assertContains(filePath, needle, label) {
  const content = read(filePath)
  if (!content.includes(needle)) {
    throw new Error(`${label} mismatch in ${filePath}\nExpected to find: ${needle}`)
  }
}

function assertRegex(filePath, pattern, label) {
  const content = read(filePath)
  if (!pattern.test(content)) {
    throw new Error(`${label} mismatch in ${filePath}\nExpected to match: ${pattern}`)
  }
}

const checks = [
  {
    label: 'worker batch accept route',
    backend: ['backend/internal/router/worker_router.go', 'PUT("/worker/batches/:batch_id/accept"'],
    frontend: ['worker-app/app/src/main/java/com/imaginai/indel/data/api/WorkerApiService.kt', /@PUT\("api\/v1\/worker\/batches\/\{batch_id\}\/accept"\)/],
  },
  {
    label: 'worker batch deliver route',
    backend: ['backend/internal/router/worker_router.go', 'PUT("/worker/batches/:batch_id/deliver"'],
    frontend: ['worker-app/app/src/main/java/com/imaginai/indel/data/api/WorkerApiService.kt', /@PUT\("api\/v1\/worker\/batches\/\{batch_id\}\/deliver"\)/],
  },
  {
    label: 'worker plan routes',
    backend: ['backend/internal/router/worker_router.go', 'GET("/worker/plans"'],
    frontend: ['worker-app/app/src/main/java/com/imaginai/indel/data/api/WorkerApiService.kt', /@GET\("api\/v1\/worker\/plans"\)/],
  },
  {
    label: 'platform zone paths',
    backend: ['backend/internal/router/platform_router.go', 'GET("/zone-paths"'],
    frontend: ['platform-dashboard/src/api/platform.ts', 'getZonePaths'],
  },
  {
    label: 'insurer batches endpoint',
    backend: ['backend/internal/router/worker_router.go', 'GET("/worker/batches"'],
    frontend: ['insurer-dashboard/src/api/insurer.ts', '/api/v1/worker/batches'],
  },
  {
    label: 'simulator batch accept method',
    backend: ['backend/internal/router/worker_router.go', 'PUT("/worker/batches/:batch_id/accept"'],
    frontend: ['delivery_batch_pickup_simulator.html', /method:\s*'PUT'/],
  },
]

for (const check of checks) {
  const [backendFile, backendNeedle] = check.backend
  const [frontendFile, frontendNeedle] = check.frontend

  if (typeof backendNeedle === 'string') {
    assertContains(backendFile, backendNeedle, `${check.label} backend`)
  } else {
    assertRegex(backendFile, backendNeedle, `${check.label} backend`)
  }

  if (typeof frontendNeedle === 'string') {
    assertContains(frontendFile, frontendNeedle, `${check.label} frontend`)
  } else {
    assertRegex(frontendFile, frontendNeedle, `${check.label} frontend`)
  }
}

console.log('Endpoint contract checks passed.')