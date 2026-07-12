<script>
    import { onMount } from 'svelte'
    import { EventsOn } from '../wailsjs/runtime/runtime.js'
    import { AddFiles, Cleanup, ClearCompleted, OpenFilePicker, OpenFolderPicker, PauseQueue, ResumeQueue, StopQueue } from '../wailsjs/go/main/App.js'
    import FileList from './components/FileList.svelte'
    import Settings from './components/Settings.svelte'
    import { jobs, ffmpegMissing, upsertJob, updateJob, hasCompleted, hasDone, hasActive } from './stores/queue.js'
    import { formatBytes } from './utils/format.js'

    let isPaused = $state(false)
    let showSettings = $state(false)
    let showAbout = $state(false)
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

    $effect(() => { if (!$hasActive) isPaused = false })

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

        EventsOn('menu:settings', () => showSettings = true)
        EventsOn('menu:about',    () => showAbout = true)
        EventsOn('menu:open',        () => openPicker())
        EventsOn('menu:open-folder', () => openFolder())
        EventsOn('menu:clear',    () => clearAll())

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

    async function openFolder() {
        const path = await OpenFolderPicker()
        if (path) {
            const newJobs = await AddFiles([path])
            if (newJobs) newJobs.forEach(j => upsertJob(j))
        }
    }

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

    function pause() {
        isPaused = true
        PauseQueue()
    }

    function resume() {
        isPaused = false
        ResumeQueue()
    }

    function stop() {
        isPaused = false
        jobs.update(list => list.filter(j => j.status !== 'waiting' && j.status !== 'processing'))
        StopQueue()
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

    <!-- Queue controls -->
    {#if $hasActive || isPaused}
    <div class="control-bar">
        {#if isPaused}
            <button class="ctrl-btn ctrl-resume" onclick={resume}>▶ Resume</button>
        {:else}
            <button class="ctrl-btn ctrl-pause" onclick={pause}>⏸ Pause</button>
        {/if}
        <button class="ctrl-btn ctrl-stop" onclick={stop}>⏹ Stop</button>
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
                <div class="drop-actions">
                    <button class="btn-open" onclick={openPicker}>Choose files…</button>
                    <button class="btn-open btn-open-secondary" onclick={openFolder}>Choose folder…</button>
                </div>
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
                <button class="btn-sm btn-muted" onclick={openPicker}>Add files…</button>
                <button class="btn-sm btn-muted" onclick={openFolder}>Add folder…</button>
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
            <div class="confirm-dialog modal-shell" onclick={(e) => e.stopPropagation()}>
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

    {#if showAbout}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div class="confirm-overlay" onclick={() => showAbout = false}>
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="confirm-dialog about-dialog modal-shell" onclick={(e) => e.stopPropagation()}>
                <p class="about-name">VideoOptim</p>
                <p class="about-version">Version 0.2.0</p>
                <p class="about-desc">Video compression for macOS.<br>Powered by ffmpeg + HEVC.</p>
                <p class="about-author">Diego Segovia @ 2026</p>
                <div class="confirm-actions">
                    <button class="confirm-ok" onclick={() => showAbout = false}>OK</button>
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
        font-family: var(--font-sans);
        color: var(--ink);
    }

    /* Warning banner */
    .warning-banner {
        background: var(--warning-bg);
        border-bottom: 1px solid var(--warning-border);
        padding: 8px 16px;
        font: 400 12px var(--font-sans);
        color: var(--warning-text);
        flex-shrink: 0;
    }

    .warning-banner code {
        background: rgba(0,0,0,0.25);
        padding: 1px 6px;
        border-radius: var(--radius-sm);
        font-family: var(--font-mono);
        font-size: 11px;
    }

    /* Control bar */
    .control-bar {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 8px 14px;
        background: var(--bg-3);
        border-bottom: 1px solid var(--line);
        flex-shrink: 0;
    }

    .ctrl-btn {
        font: 500 12px var(--font-mono);
        padding: 4px 11px;
        border-radius: var(--radius-md);
        border: 1px solid var(--line-2);
        background: var(--bg-2);
        color: var(--ink-2);
        cursor: pointer;
        transition: background var(--dur-fast) var(--ease);
    }

    .ctrl-btn:hover { background: var(--bg-btn-hover); }
    .ctrl-resume { color: var(--accent); border-color: var(--accent-line); background: var(--accent-dim); }
    .ctrl-stop   { color: var(--danger); }

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
        gap: 14px;
        background: var(--bg-2);
    }

    .drop-icon { color: var(--ink-4); margin-bottom: 4px; }

    .drop-hint {
        margin: 0;
        font: 400 14px var(--font-sans);
        color: var(--ink-3);
    }

    .btn-open {
        padding: 6px 16px;
        border-radius: var(--radius-md);
        border: 1px solid var(--line-2);
        background: var(--bg-3);
        font: 500 12.5px var(--font-sans);
        cursor: pointer;
        color: var(--ink-2);
        transition: background var(--dur-fast) var(--ease);
    }

    .btn-open:hover { background: var(--bg-btn-hover); }

    .drop-actions { display: flex; gap: 8px; }
    .btn-open-secondary { color: var(--ink-3); }

    /* Bottom toolbar */
    .toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        height: 44px;
        padding: 0 12px;
        background: var(--bg-3);
        border-top: 1px solid var(--line);
        flex-shrink: 0;
    }

    .toolbar-left,
    .toolbar-right {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .btn-sm {
        padding: 3px 10px;
        border-radius: var(--radius-md);
        border: 1px solid var(--line-2);
        background: transparent;
        font: 400 12px var(--font-sans);
        cursor: pointer;
        color: var(--ink-2);
        transition: background var(--dur-fast) var(--ease);
    }

    .btn-sm:hover { background: var(--bg-btn-hover); }
    .btn-muted { color: var(--ink-3); }

    .btn-icon {
        width: 28px;
        height: 28px;
        border-radius: var(--radius-md);
        border: none;
        background: transparent;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--ink-2);
    }

    .btn-icon:hover { background: var(--bg-2); }

    /* Summary bar */
    .summary-bar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 8px 16px;
        background: var(--bg-3);
        border-top: 1px solid var(--line);
        flex-shrink: 0;
        gap: 8px;
    }

    .summary-counts, .summary-sizes {
        display: flex;
        align-items: center;
        gap: 6px;
        font: 400 12px var(--font-mono);
        font-variant-numeric: tabular-nums;
    }

    .sum-done  { color: var(--accent); font-weight: 600; }
    .sum-skip  { color: var(--ink-3); }
    .sum-error { color: var(--danger); }
    .sum-total { color: var(--ink-2); }
    .sum-sep   { color: var(--ink-4); }
    .sum-orig  { color: var(--ink-2); }
    .sum-arrow { color: var(--ink-3); }
    .sum-out   { color: var(--accent); font-weight: 600; }
    .sum-pct   { color: var(--accent); font-weight: 600; }

    .toolbar-spinner {
        width: 11px;
        height: 11px;
        flex-shrink: 0;
    }

    .toolbar-progress {
        font: 400 11.5px var(--font-mono);
        color: var(--ink-3);
    }

    .cleanup-feedback {
        font: 400 12px var(--font-mono);
        color: var(--accent);
    }

    /* Drop overlay */
    .drop-overlay {
        position: absolute;
        inset: 0 0 44px 0;
        background: var(--accent-dim);
        border: 2px dashed var(--accent-line);
        margin: 8px;
        border-radius: var(--radius-xl);
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
        z-index: 50;
    }

    .drop-overlay-label {
        font: 600 18px var(--font-sans);
        color: var(--accent);
    }

    /* Modals */
    .confirm-overlay {
        position: fixed;
        inset: 0;
        background: var(--bg-overlay);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 200;
        backdrop-filter: blur(2px);
    }

    .confirm-dialog {
        padding: 24px 24px 20px;
        width: 360px;
    }

    .confirm-title {
        margin: 0 0 8px;
        font: 600 15px var(--font-sans);
        color: var(--ink);
    }

    .confirm-body {
        margin: 0 0 20px;
        font: 400 13px/1.6 var(--font-sans);
        color: var(--ink-2);
    }

    .confirm-body code {
        font-family: var(--font-mono);
        font-size: 11.5px;
        background: var(--bg-3);
        padding: 1px 5px;
        border-radius: var(--radius-sm);
    }

    .confirm-actions {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
    }

    .confirm-cancel, .confirm-ok {
        padding: 6px 14px;
        border-radius: var(--radius-md);
        font: 500 13px var(--font-sans);
        cursor: pointer;
        border: none;
    }

    .confirm-cancel {
        background: var(--bg-3);
        color: var(--ink-2);
        border: 1px solid var(--line-2);
    }

    .confirm-cancel:hover { background: var(--bg-btn-hover); }

    .confirm-ok {
        background: var(--danger);
        color: white;
    }

    .confirm-ok:hover { opacity: 0.88; }

    .about-dialog { text-align: center; width: 280px; padding: 28px 24px 20px; }
    .about-name    { margin: 0 0 4px; font: 700 17px var(--font-sans); color: var(--ink); }
    .about-version, .about-author { margin: 0 0 12px; font: 400 11.5px var(--font-mono); color: var(--ink-3); }
    .about-desc    { margin: 0 0 20px; font: 400 12.5px/1.6 var(--font-sans); color: var(--ink-2); }
    .about-dialog .confirm-ok { background: var(--accent); }
</style>
