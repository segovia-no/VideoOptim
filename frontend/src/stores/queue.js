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
