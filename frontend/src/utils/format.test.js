import { describe, it, expect } from 'vitest'
import { formatBytes } from './format.js'

describe('formatBytes', () => {
    it('returns em dash for falsy values', () => {
        expect(formatBytes(0)).toBe('—')
        expect(formatBytes(null)).toBe('—')
        expect(formatBytes(undefined)).toBe('—')
    })

    it('formats bytes as KB', () => {
        expect(formatBytes(1024)).toBe('1 KB')
        expect(formatBytes(512)).toBe('1 KB')   // rounds to 1
        expect(formatBytes(2048)).toBe('2 KB')
    })

    it('formats as MB when >= 1 MB', () => {
        expect(formatBytes(1048576)).toBe('1.0 MB')
        expect(formatBytes(1572864)).toBe('1.5 MB')
        expect(formatBytes(10485760)).toBe('10.0 MB')
    })

    it('formats as GB when >= 1 GB', () => {
        expect(formatBytes(1073741824)).toBe('1.00 GB')
        expect(formatBytes(2147483648)).toBe('2.00 GB')
        expect(formatBytes(1610612736)).toBe('1.50 GB')
    })

    it('MB boundary: just under 1 GB uses MB', () => {
        expect(formatBytes(1073741823)).toMatch(/MB$/)
    })

    it('KB boundary: just under 1 MB uses KB', () => {
        expect(formatBytes(1048575)).toMatch(/KB$/)
    })
})
