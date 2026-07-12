<script>
    import { onMount } from 'svelte'
    import { RevealInFinder, OpenFile, MoveToTrash } from '../../wailsjs/go/main/App.js'
    import { updateJob, openMenuId } from '../stores/queue.js'

    let { job } = $props()

    let menuX = $state(0)
    let menuY = $state(0)
    let showMenu = $derived($openMenuId === job.id)

    // Prefer optimized file for Reveal/Open when done
    let primaryPath = $derived(
        (job.status === 'done' && job.outputPath) ? job.outputPath : job.path
    )

    let canDeleteOriginal = $derived(
        job.status === 'done' && !!job.outputPath && !job.originalDeleted
    )

    function openMenu(e) {
        e.stopPropagation()
        const menuW = 190
        const menuH = canDeleteOriginal ? 114 : 80
        menuX = (e.clientX + menuW > window.innerWidth) ? e.clientX - menuW : e.clientX
        menuY = (e.clientY + menuH > window.innerHeight) ? e.clientY - menuH : e.clientY
        openMenuId.set(showMenu ? null : job.id)
    }

    function closeMenu() { openMenuId.set(null) }

    async function revealInFinder(e) {
        e.stopPropagation()
        closeMenu()
        await RevealInFinder(primaryPath)
    }

    async function openFile(e) {
        e.stopPropagation()
        closeMenu()
        await OpenFile(primaryPath)
    }

    async function deleteOriginal(e) {
        e.stopPropagation()
        closeMenu()
        const err = await MoveToTrash(job.path)
        if (!err) updateJob(job.id, { originalDeleted: true })
    }

    function formatBytes(bytes) {
        if (!bytes) return '—'
        const mb = bytes / 1048576
        return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
    }

    function statusIcon(status) {
        switch (status) {
            case 'done':      return '✓'
            case 'skipped':   return '—'
            case 'error':     return '✕'
            case 'cancelled': return '⊘'
            default:          return '·'
        }
    }

    function statusClass(status) {
        switch (status) {
            case 'done':  return 'icon-done'
            case 'error': return 'icon-error'
            default:      return 'icon-neutral'
        }
    }

    onMount(() => {
        const close = () => openMenuId.set(null)
        window.addEventListener('click', close)
        return () => window.removeEventListener('click', close)
    })
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="row" class:processing={job.status === 'processing'} onclick={openMenu}>
    <div class="col-name">
        <span class="filename" title={job.path}>{job.filename}</span>
    </div>
    <div class="col-orig">{formatBytes(job.originalSize)}</div>
    <div class="col-output">
        {#if job.status === 'processing'}
            <div class="progress-wrap">
                <div class="progress-bar" style="width: {job.progress}%"></div>
            </div>
            <span class="progress-label">{Math.round(job.progress)}%</span>
            <span class="elapsed">{job.elapsed || '0:00'}</span>
        {:else if job.status === 'done'}
            <span class="output-size">{formatBytes(job.outputSize)}</span>
        {:else if job.status === 'error'}
            <span class="error-msg" title={job.error}>Error</span>
        {:else}
            <span class="placeholder">—</span>
        {/if}
    </div>
    <div class="col-savings">
        {#if job.status === 'done'}
            <span class="savings">−{job.savings.toFixed(1)}%</span>
        {:else if job.status === 'skipped' && !job.skipReason}
            <span class="skipped-label">No gain</span>
        {:else}
            <span class="placeholder">—</span>
        {/if}
    </div>
    <div class="col-status">
        {#if job.status === 'processing' || job.status === 'waiting'}
            <span class="spinner" class:spinner-dim={job.status === 'waiting'}></span>
        {:else if job.status === 'skipped' && job.skipReason === 'hevc'}
            <span class="icon-warn">⚠<span class="warn-tip">Already H.265 — re-encoding is unlikely to reduce file size</span></span>
        {:else}
            <span class={statusClass(job.status)}>{statusIcon(job.status)}</span>
        {/if}
    </div>
</div>

{#if showMenu}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="menu" style="left:{menuX}px; top:{menuY}px" onclick={(e) => e.stopPropagation()}>
        <button class="menu-item" onclick={revealInFinder}>Reveal in Finder</button>
        <button class="menu-item" onclick={openFile}>Open file</button>
        {#if canDeleteOriginal}
            <div class="menu-separator"></div>
            <button class="menu-item menu-danger" onclick={deleteOriginal}>Delete original</button>
        {/if}
    </div>
{/if}

<style>
    .row {
        display: flex;
        align-items: center;
        padding: 0 16px;
        height: 40px;
        border-bottom: 1px solid var(--border-row);
        gap: 8px;
        font-size: 13px;
        background: var(--bg-row);
        cursor: pointer;
        user-select: none;
    }

    .row:last-child { border-bottom: none; }
    .row:hover { background: var(--bg-row-hover); }
    .row.processing { background: var(--bg-row-processing); }

    .col-name {
        flex: 1;
        min-width: 0;
        overflow: hidden;
    }

    .filename {
        display: block;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        color: var(--text-primary);
    }

    .col-orig {
        width: 80px;
        text-align: right;
        color: var(--text-secondary);
        font-variant-numeric: tabular-nums;
    }

    .col-output {
        width: 170px;
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 6px;
        font-variant-numeric: tabular-nums;
    }

    .col-savings {
        width: 72px;
        text-align: right;
        font-variant-numeric: tabular-nums;
    }

    .col-status {
        width: 22px;
        text-align: center;
        flex-shrink: 0;
    }

    .progress-wrap {
        flex: 1;
        height: 5px;
        background: var(--progress-track);
        border-radius: 3px;
        overflow: hidden;
    }

    .progress-bar {
        height: 100%;
        background: var(--progress-fill);
        border-radius: 3px;
        transition: width 0.25s ease;
    }

    .progress-label {
        font-size: 11px;
        color: var(--text-secondary);
        min-width: 28px;
        text-align: right;
    }

    .elapsed {
        font-size: 11px;
        color: var(--text-muted);
        min-width: 30px;
        text-align: right;
    }

    .savings      { color: var(--green); font-weight: 600; font-size: 13px; }
    .skipped-label { color: var(--text-muted); font-size: 11px; }
    .output-size  { color: var(--text-secondary); }
    .error-msg    { color: var(--red); font-size: 11px; }
    .placeholder  { color: var(--text-placeholder); }

    .icon-done    { color: var(--green); font-weight: 700; }
    .icon-error   { color: var(--red); }
    .icon-neutral { color: var(--text-placeholder); }
    .icon-warn {
        position: relative;
        color: var(--orange);
        cursor: default;
    }

    .warn-tip {
        display: none;
        position: absolute;
        bottom: calc(100% + 6px);
        right: 0;
        background: var(--bg-modal);
        border: 1px solid var(--border);
        border-radius: 6px;
        padding: 5px 8px;
        font-size: 11px;
        color: var(--text-secondary);
        white-space: nowrap;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        pointer-events: none;
        z-index: 400;
    }

    .icon-warn:hover .warn-tip { display: block; }

    .spinner {
        display: inline-block;
        width: 13px;
        height: 13px;
        border: 2px solid var(--spinner-track);
        border-top-color: var(--spinner-head);
        border-radius: 50%;
        animation: spin 0.75s linear infinite;
    }

    @keyframes spin { to { transform: rotate(360deg); } }

    .spinner-dim { border-top-color: var(--text-muted); opacity: 0.5; }

    /* Context menu */
    .menu {
        position: fixed;
        z-index: 300;
        background: var(--bg-modal);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 4px;
        min-width: 190px;
        box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        backdrop-filter: blur(12px);
    }

    .menu-item {
        display: block;
        width: 100%;
        padding: 6px 10px;
        border: none;
        background: none;
        font-size: 13px;
        font-family: inherit;
        color: var(--text-primary);
        text-align: left;
        cursor: pointer;
        border-radius: 5px;
    }

    .menu-item:hover { background: var(--accent); color: white; }

    .menu-danger { color: var(--red); }
    .menu-danger:hover { background: var(--red); color: white; }

    .menu-separator {
        height: 1px;
        background: var(--border);
        margin: 4px 0;
    }
</style>
