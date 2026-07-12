<script>
    import { onMount } from 'svelte'
    import { EventsOn } from '../wailsjs/runtime/runtime.js'
    import { AddFiles, Cleanup, ClearCompleted, OpenFilePicker } from '../wailsjs/go/main/App.js'
    import FileList from './components/FileList.svelte'
    import Settings from './components/Settings.svelte'
    import { jobs, ffmpegMissing, upsertJob, updateJob, hasCompleted, hasDone, hasActive } from './stores/queue.js'

    let showSettings = $state(false)
    let showCleanupConfirm = $state(false)
    let dropping = $state(false)
    let cleanupResult = $state(null)
    let cleanupTimer = null

    let doneJobs     = $derived($jobs.filter(j => j.status === 'done'))
    let skippedCount = $derived($jobs.filter(j => j.status === 'skipped').length)
    let errorCount   = $derived($jobs.filter(j => j.status === 'error').length)
    let totalOrigBytes = $derived(doneJobs.reduce((s, j) => s + (j.originalSize || 0), 0))
    let totalOutBytes  = $derived(doneJobs.reduce((s, j) => s + (j.outputSize  || 0), 0))
    let reduction      = $derived(totalOrigBytes > 0 ? (1 - totalOutBytes / totalOrigBytes) * 100 : 0)
    let showSummary    = $derived(!$hasActive && $jobs.length > 0 && $hasCompleted)

    function formatBytes(bytes) {
        if (!bytes) return '—'
        const gb = bytes / 1073741824
        const mb = bytes / 1048576
        return gb >= 1 ? `${gb.toFixed(2)} GB` : mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
    }

    onMount(() => {
        let dropTimer = null
        let pendingPaths = new Set()
        EventsOn('wails:file-drop', (_x, _y, paths) => {
            dropping = false
            if (paths && paths.length > 0) {
                paths.forEach(p => pendingPaths.add(p))
                clearTimeout(dropTimer)
                dropTimer = setTimeout(() => {
                    const batch = [...pendingPaths]
                    pendingPaths.clear()
                    AddFiles(batch).then(newJobs => {
                        if (newJobs) newJobs.forEach(j => upsertJob(j))
                    })
                }, 50)
            }
        })

        EventsOn('ffmpeg:missing', (data) => ffmpegMissing.set(data))
        EventsOn('job:start',    (data) => updateJob(data.id, { status: 'processing', progress: 0 }))
        EventsOn('job:progress', (data) => updateJob(data.id, { progress: data.percent, elapsed: data.elapsed, fps: data.fps }))
        EventsOn('job:complete', (data) => {
            const hasOutput = !!data.outputPath
            const savings = hasOutput && data.outputSize < data.originalSize
                ? (1 - data.outputSize / data.originalSize) * 100 : 0
            updateJob(data.id, {
                status: hasOutput ? 'done' : 'skipped',
                progress: 100, outputPath: data.outputPath,
                originalSize: data.originalSize, outputSize: data.outputSize, savings,
                skipReason: data.skipReason || null,
            })
        })
        EventsOn('job:error', (data) => updateJob(data.id, { status: 'error', error: data.message }))

        const onDragEnter = () => { dropping = true }
        const onDragLeave = (e) => { if (!e.relatedTarget) dropping = false }
        window.addEventListener('dragenter', onDragEnter)
        window.addEventListener('dragleave', onDragLeave)
        return () => {
            window.removeEventListener('dragenter', onDragEnter)
            window.removeEventListener('dragleave', onDragLeave)
        }
    })

    async function openPicker() {
        const paths = await OpenFilePicker()
        if (paths && paths.length > 0) {
            const newJobs = await AddFiles(paths)
            if (newJobs) newJobs.forEach(j => upsertJob(j))
        }
    }

    async function runCleanup() {
        showCleanupConfirm = false
        const result = await Cleanup()
        cleanupResult = result
        clearTimeout(cleanupTimer)
        cleanupTimer = setTimeout(() => cleanupResult = null, 4000)
    }

    function clearAll() {
        ClearCompleted()
        jobs.set([])
    }
</script>

<div class="app" class:dropping>

    <!-- ffmpeg missing banner -->
    {#if $ffmpegMissing}
        <div class="warning-banner">
            <strong>ffmpeg not found.</strong> Run <code>brew install ffmpeg</code> then relaunch.
        </div>
    {/if}

    <!-- Main content -->
    <div class="content">
        {#if $jobs.length === 0}
            <div class="drop-zone">
                <svg class="drop-icon" width="52" height="52" viewBox="0 0 52 52" fill="none">
                    <rect x="6" y="4" width="40" height="44" rx="5" stroke="var(--border)" stroke-width="2" stroke-dasharray="5 3"/>
                    <path d="M26 16v20M18 28l8 8 8-8" stroke="var(--text-placeholder)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <p class="drop-hint">Drop videos or folders here</p>
                <button class="btn-open" onclick={openPicker}>Choose files…</button>
            </div>
        {:else}
            <FileList />
        {/if}
    </div>

    <!-- Summary bar -->
    {#if showSummary}
    <div class="summary-bar">
        <div class="summary-counts">
            <span class="sum-done">{doneJobs.length} transcoded</span>
            <span class="sum-sep">·</span>
            <span class="sum-skip">{skippedCount} skipped</span>
            {#if errorCount > 0}
                <span class="sum-sep">·</span>
                <span class="sum-error">{errorCount} failed</span>
            {/if}
            <span class="sum-sep">·</span>
            <span class="sum-total">{$jobs.length} total</span>
        </div>
        {#if totalOrigBytes > 0}
        <div class="summary-sizes">
            <span class="sum-orig">{formatBytes(totalOrigBytes)}</span>
            <span class="sum-arrow">→</span>
            <span class="sum-out">{formatBytes(totalOutBytes)}</span>
            <span class="sum-pct">·&nbsp;{reduction.toFixed(1)}% smaller</span>
        </div>
        {/if}
    </div>
    {/if}

    <!-- Bottom toolbar -->
    <div class="toolbar">
        <div class="toolbar-left">
            {#if $jobs.length > 0}
                <button class="btn-sm btn-muted" onclick={openPicker}>Add more…</button>
            {/if}
            {#if $hasActive}
                <span class="toolbar-spinner"></span>
                <span class="toolbar-progress">
                    Transcoded {$jobs.filter(j => j.status === 'done').length} of {$jobs.length} videos
                </span>
            {/if}
        </div>
        <div class="toolbar-right">
            {#if cleanupResult}
                <span class="cleanup-feedback">
                    {cleanupResult.moved} moved to Trash
                    {#if cleanupResult.deleted > 0}, {cleanupResult.deleted} deleted{/if}
                </span>
            {/if}
            {#if $hasDone}
                <button class="btn-sm" onclick={() => showCleanupConfirm = true}>Clean up originals</button>
            {/if}
            {#if $hasCompleted}
                <button class="btn-sm btn-muted" onclick={clearAll}>Clear list</button>
            {/if}
            <button class="btn-icon" onclick={() => showSettings = true} title="Settings">
                <svg width="15" height="15" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" clip-rule="evenodd"
                        d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"/>
                </svg>
            </button>
        </div>
    </div>

    <!-- Drop overlay -->
    {#if dropping}
        <div class="drop-overlay">
            <span class="drop-overlay-label">Drop to compress</span>
        </div>
    {/if}

    {#if showCleanupConfirm}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div class="confirm-overlay" onclick={() => showCleanupConfirm = false}>
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="confirm-dialog" onclick={(e) => e.stopPropagation()}>
                <p class="confirm-title">Move originals to Trash?</p>
                <p class="confirm-body">
                    Original files with a smaller <code>_optimized</code> version will be moved to the Trash.
                    This cannot be undone from within the app.
                </p>
                <div class="confirm-actions">
                    <button class="confirm-cancel" onclick={() => showCleanupConfirm = false}>Cancel</button>
                    <button class="confirm-ok" onclick={runCleanup}>Move to Trash</button>
                </div>
            </div>
        </div>
    {/if}

    {#if showSettings}
        <Settings onClose={() => showSettings = false} />
    {/if}
</div>

<style>
    .app {
        display: flex;
        flex-direction: column;
        height: 100vh;
        background: var(--bg);
        position: relative;
        overflow: hidden;
        font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", sans-serif;
        color: var(--text-primary);
    }

    /* Warning banner */
    .warning-banner {
        background: var(--warning-bg);
        border-bottom: 1px solid var(--warning-border);
        padding: 7px 16px;
        font-size: 12px;
        color: var(--warning-text);
        flex-shrink: 0;
    }

    .warning-banner code {
        background: rgba(0,0,0,0.1);
        padding: 1px 5px;
        border-radius: 3px;
        font-family: "SF Mono", Menlo, monospace;
        font-size: 11px;
    }

    /* Content */
    .content {
        flex: 1;
        display: flex;
        flex-direction: column;
        min-height: 0;
        overflow: hidden;
    }

    /* Empty drop zone */
    .drop-zone {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 10px;
    }

    .drop-icon { margin-bottom: 4px; }

    .drop-hint {
        margin: 0;
        font-size: 15px;
        color: var(--text-muted);
    }

    .btn-open {
        margin-top: 4px;
        padding: 7px 18px;
        border-radius: 8px;
        border: 1px solid var(--border-btn);
        background: var(--bg-btn);
        font-size: 13px;
        cursor: pointer;
        color: var(--text-btn);
        font-family: inherit;
        box-shadow: 0 1px 2px rgba(0,0,0,0.06);
    }

    .btn-open:hover { background: var(--bg-btn-hover); }

    /* Bottom toolbar */
    .toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        height: 36px;
        padding: 0 8px;
        background: var(--bg-toolbar);
        border-top: 1px solid var(--border);
        flex-shrink: 0;
    }

    .toolbar-left,
    .toolbar-right {
        display: flex;
        align-items: center;
        gap: 6px;
    }

    .btn-sm {
        padding: 3px 10px;
        border-radius: 5px;
        border: 1px solid var(--border-btn);
        background: var(--bg-btn);
        font-size: 12px;
        cursor: pointer;
        color: var(--text-btn);
        font-family: inherit;
    }

    .btn-sm:hover { background: var(--bg-btn-hover); }
    .btn-muted { color: var(--text-muted); }

    .btn-icon {
        width: 28px;
        height: 28px;
        border-radius: 6px;
        border: none;
        background: transparent;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-secondary);
    }

    .btn-icon:hover { background: var(--bg-btn-hover); }

    /* Summary bar */
    .summary-bar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 6px 16px;
        background: var(--bg-header);
        border-top: 1px solid var(--border);
        flex-shrink: 0;
        gap: 8px;
    }

    .summary-counts, .summary-sizes {
        display: flex;
        align-items: center;
        gap: 5px;
        font-size: 12px;
        font-variant-numeric: tabular-nums;
    }

    .sum-done  { color: var(--green); font-weight: 600; }
    .sum-skip  { color: var(--text-muted); }
    .sum-error { color: var(--red); }
    .sum-total { color: var(--text-secondary); }
    .sum-sep   { color: var(--text-placeholder); }
    .sum-orig  { color: var(--text-secondary); }
    .sum-arrow { color: var(--text-secondary); }
    .sum-out   { color: var(--green); font-weight: 600; }
    .sum-pct   { color: var(--green); font-weight: 600; }

    .toolbar-spinner {
        display: inline-block;
        width: 11px;
        height: 11px;
        border: 1.5px solid var(--spinner-track);
        border-top-color: var(--spinner-head);
        border-radius: 50%;
        animation: spin 0.75s linear infinite;
        flex-shrink: 0;
    }

    @keyframes spin { to { transform: rotate(360deg); } }

    .toolbar-progress {
        font-size: 12px;
        color: var(--text-muted);
    }

    .cleanup-feedback {
        font-size: 12px;
        color: var(--green);
    }

    /* Drop overlay */
    .drop-overlay {
        position: absolute;
        inset: 0 0 36px 0;
        background: color-mix(in srgb, var(--accent) 8%, transparent);
        border: 2px dashed var(--accent);
        margin: 8px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
        z-index: 50;
    }

    .drop-overlay-label {
        font-size: 18px;
        font-weight: 600;
        color: var(--accent);
    }

    /* Confirmation dialog */
    .confirm-overlay {
        position: fixed;
        inset: 0;
        background: var(--bg-overlay);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 200;
    }

    .confirm-dialog {
        background: var(--bg-modal);
        border: 1px solid var(--border);
        border-radius: 12px;
        padding: 24px 24px 20px;
        width: 340px;
        box-shadow: 0 20px 60px rgba(0,0,0,0.3);
    }

    .confirm-title {
        margin: 0 0 8px;
        font-size: 15px;
        font-weight: 600;
        color: var(--text-primary);
    }

    .confirm-body {
        margin: 0 0 20px;
        font-size: 13px;
        color: var(--text-secondary);
        line-height: 1.5;
    }

    .confirm-body code {
        font-family: "SF Mono", Menlo, monospace;
        font-size: 11px;
        background: var(--bg-btn);
        padding: 1px 4px;
        border-radius: 3px;
    }

    .confirm-actions {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
    }

    .confirm-cancel, .confirm-ok {
        padding: 6px 14px;
        border-radius: 7px;
        font-size: 13px;
        font-family: inherit;
        cursor: pointer;
        border: none;
    }

    .confirm-cancel {
        background: var(--bg-btn);
        color: var(--text-btn);
        border: 1px solid var(--border-btn);
    }

    .confirm-cancel:hover { background: var(--bg-btn-hover); }

    .confirm-ok {
        background: var(--red);
        color: white;
        font-weight: 500;
    }

    .confirm-ok:hover { opacity: 0.88; }
</style>
