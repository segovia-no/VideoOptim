import { EventsOn } from '../../wailsjs/runtime/runtime.js'
import { ffmpegMissing, updateJob } from '../stores/queue.js'

export function setupEvents({ onSettings, onAbout, onOpen, onFolder, onClear }) {
    EventsOn('menu:settings',    onSettings)
    EventsOn('menu:about',       onAbout)
    EventsOn('menu:open',        onOpen)
    EventsOn('menu:open-folder', onFolder)
    EventsOn('menu:clear',       onClear)

    EventsOn('ffmpeg:missing', data => ffmpegMissing.set(data))

    EventsOn('job:start',    data => updateJob(data.id, { status: 'processing', progress: 0 }))
    EventsOn('job:progress', data => updateJob(data.id, { progress: data.percent, elapsed: data.elapsed, fps: data.fps }))
    EventsOn('job:complete', data => {
        const hasOutput = !!data.outputPath
        const savings = hasOutput && data.outputSize < data.originalSize
            ? (1 - data.outputSize / data.originalSize) * 100 : 0
        updateJob(data.id, {
            status:       hasOutput ? 'done' : 'skipped',
            progress:     100,
            outputPath:   data.outputPath,
            originalSize: data.originalSize,
            outputSize:   data.outputSize,
            savings,
            skipReason:   data.skipReason || null,
        })
    })
    EventsOn('job:error', data => updateJob(data.id, { status: 'error', error: data.message }))
}
