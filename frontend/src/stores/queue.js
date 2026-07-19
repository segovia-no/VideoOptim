import { writable, derived } from 'svelte/store'

export const jobs = writable([])
export const ffmpegMissing = writable(null)
export const openMenuId = writable(null)

export function upsertJob(job) {
    jobs.update(list => {
        const idx = list.findIndex(j => j.id === job.id)
        if (idx >= 0) {
            const next = [...list]
            next[idx] = { ...next[idx], ...job }
            return next
        }
        return [...list, job]
    })
}

export function updateJob(id, patch) {
    jobs.update(list => list.map(j => j.id === id ? { ...j, ...patch } : j))
}

export const hasCompleted = derived(jobs, $jobs =>
    $jobs.some(j => ['done', 'skipped', 'error', 'cancelled'].includes(j.status))
)

export const hasDone = derived(jobs, $jobs =>
    $jobs.some(j => j.status === 'done')
)

export const hasActive = derived(jobs, $jobs =>
    $jobs.some(j => j.status === 'waiting' || j.status === 'processing')
)

export const stats = derived(jobs, $jobs => {
    const done = $jobs.filter(j => j.status === 'done')
    const totalOrigBytes = done.reduce((s, j) => s + (j.originalSize || 0), 0)
    const totalOutBytes  = done.reduce((s, j) => s + (j.outputSize  || 0), 0)
    return {
        doneCount:     done.length,
        skippedCount:  $jobs.filter(j => j.status === 'skipped').length,
        errorCount:    $jobs.filter(j => j.status === 'error').length,
        totalOrigBytes,
        totalOutBytes,
        reduction: totalOrigBytes > 0 ? (1 - totalOutBytes / totalOrigBytes) * 100 : 0,
    }
})

export const showSummary = derived([jobs, hasActive, hasCompleted], ([$jobs, $hasActive, $hasCompleted]) =>
    !$hasActive && $jobs.length > 0 && $hasCompleted
)
