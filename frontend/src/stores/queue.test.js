import { describe, it, expect, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import { jobs, upsertJob, updateJob, hasActive, hasDone, hasCompleted, stats, showSummary } from './queue.js'

beforeEach(() => { jobs.set([]) })

// ── upsertJob ──────────────────────────────────────────────────────────────

describe('upsertJob', () => {
    it('adds new job', () => {
        upsertJob({ id: '1', status: 'waiting' })
        expect(get(jobs)).toHaveLength(1)
    })

    it('merges when job already exists', () => {
        upsertJob({ id: '1', status: 'waiting', progress: 0 })
        upsertJob({ id: '1', status: 'processing', progress: 42 })
        const list = get(jobs)
        expect(list).toHaveLength(1)
        expect(list[0].status).toBe('processing')
        expect(list[0].progress).toBe(42)
    })

    it('preserves fields not in patch', () => {
        upsertJob({ id: '1', status: 'waiting', filename: 'video.mp4' })
        upsertJob({ id: '1', status: 'done' })
        expect(get(jobs)[0].filename).toBe('video.mp4')
    })

    it('appends multiple distinct jobs in order', () => {
        upsertJob({ id: '1', status: 'waiting' })
        upsertJob({ id: '2', status: 'waiting' })
        upsertJob({ id: '3', status: 'waiting' })
        const ids = get(jobs).map(j => j.id)
        expect(ids).toEqual(['1', '2', '3'])
    })
})

// ── updateJob ──────────────────────────────────────────────────────────────

describe('updateJob', () => {
    it('patches matching job', () => {
        jobs.set([{ id: '1', status: 'waiting', progress: 0 }])
        updateJob('1', { status: 'processing', progress: 55 })
        expect(get(jobs)[0]).toMatchObject({ status: 'processing', progress: 55 })
    })

    it('ignores non-matching id', () => {
        jobs.set([{ id: '1', status: 'waiting' }])
        updateJob('999', { status: 'done' })
        expect(get(jobs)[0].status).toBe('waiting')
    })

    it('does not mutate other jobs', () => {
        jobs.set([
            { id: '1', status: 'waiting' },
            { id: '2', status: 'waiting' },
        ])
        updateJob('1', { status: 'done' })
        expect(get(jobs)[1].status).toBe('waiting')
    })
})

// ── hasActive ──────────────────────────────────────────────────────────────

describe('hasActive', () => {
    it('true when a job is waiting', () => {
        jobs.set([{ id: '1', status: 'waiting' }])
        expect(get(hasActive)).toBe(true)
    })

    it('true when a job is processing', () => {
        jobs.set([{ id: '1', status: 'processing' }])
        expect(get(hasActive)).toBe(true)
    })

    it('false when all jobs are done', () => {
        jobs.set([{ id: '1', status: 'done' }, { id: '2', status: 'skipped' }])
        expect(get(hasActive)).toBe(false)
    })

    it('false when list is empty', () => {
        expect(get(hasActive)).toBe(false)
    })
})

// ── hasDone ────────────────────────────────────────────────────────────────

describe('hasDone', () => {
    it('true when at least one done job', () => {
        jobs.set([{ id: '1', status: 'done' }])
        expect(get(hasDone)).toBe(true)
    })

    it('false when no done jobs', () => {
        jobs.set([{ id: '1', status: 'skipped' }, { id: '2', status: 'error' }])
        expect(get(hasDone)).toBe(false)
    })
})

// ── hasCompleted ───────────────────────────────────────────────────────────

describe('hasCompleted', () => {
    it('true for done', () => {
        jobs.set([{ id: '1', status: 'done' }])
        expect(get(hasCompleted)).toBe(true)
    })

    it('true for skipped', () => {
        jobs.set([{ id: '1', status: 'skipped' }])
        expect(get(hasCompleted)).toBe(true)
    })

    it('true for error', () => {
        jobs.set([{ id: '1', status: 'error' }])
        expect(get(hasCompleted)).toBe(true)
    })

    it('true for cancelled', () => {
        jobs.set([{ id: '1', status: 'cancelled' }])
        expect(get(hasCompleted)).toBe(true)
    })

    it('false when only active jobs', () => {
        jobs.set([{ id: '1', status: 'waiting' }, { id: '2', status: 'processing' }])
        expect(get(hasCompleted)).toBe(false)
    })
})

// ── stats ──────────────────────────────────────────────────────────────────

describe('stats', () => {
    it('zeros when empty', () => {
        const s = get(stats)
        expect(s.doneCount).toBe(0)
        expect(s.skippedCount).toBe(0)
        expect(s.errorCount).toBe(0)
        expect(s.totalOrigBytes).toBe(0)
        expect(s.totalOutBytes).toBe(0)
        expect(s.reduction).toBe(0)
    })

    it('counts statuses correctly', () => {
        jobs.set([
            { id: '1', status: 'done',    originalSize: 1000, outputSize: 600 },
            { id: '2', status: 'done',    originalSize: 2000, outputSize: 1000 },
            { id: '3', status: 'skipped', originalSize: 500,  outputSize: 500 },
            { id: '4', status: 'error' },
            { id: '5', status: 'waiting' },
        ])
        const s = get(stats)
        expect(s.doneCount).toBe(2)
        expect(s.skippedCount).toBe(1)
        expect(s.errorCount).toBe(1)
    })

    it('sums bytes from done jobs only', () => {
        jobs.set([
            { id: '1', status: 'done',    originalSize: 1000, outputSize: 600 },
            { id: '2', status: 'done',    originalSize: 2000, outputSize: 1000 },
            { id: '3', status: 'skipped', originalSize: 9999, outputSize: 9999 },
        ])
        const s = get(stats)
        expect(s.totalOrigBytes).toBe(3000)
        expect(s.totalOutBytes).toBe(1600)
    })

    it('calculates reduction percentage', () => {
        jobs.set([
            { id: '1', status: 'done', originalSize: 1000, outputSize: 500 },
        ])
        expect(get(stats).reduction).toBeCloseTo(50, 5)
    })

    it('reduction is 0 when no done jobs', () => {
        jobs.set([{ id: '1', status: 'skipped' }])
        expect(get(stats).reduction).toBe(0)
    })

    it('handles missing size fields gracefully', () => {
        jobs.set([{ id: '1', status: 'done' }])
        const s = get(stats)
        expect(s.totalOrigBytes).toBe(0)
        expect(s.reduction).toBe(0)
    })
})

// ── showSummary ────────────────────────────────────────────────────────────

describe('showSummary', () => {
    it('false when no jobs', () => {
        expect(get(showSummary)).toBe(false)
    })

    it('false when active jobs remain', () => {
        jobs.set([
            { id: '1', status: 'processing' },
            { id: '2', status: 'done' },
        ])
        expect(get(showSummary)).toBe(false)
    })

    it('true when all done and no active', () => {
        jobs.set([
            { id: '1', status: 'done' },
            { id: '2', status: 'skipped' },
        ])
        expect(get(showSummary)).toBe(true)
    })
})
