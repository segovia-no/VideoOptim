<script>
    import { onMount } from 'svelte'
    import { RevealInFinder, OpenFile, MoveToTrash } from '../../wailsjs/go/main/App.js'
    import { openMenuId } from '../stores/queue.js'
    import { formatBytes } from '../utils/format.js'

    let { job } = $props()

    let menuX = $state(0)
    let menuY = $state(0)
    let showMenu = $derived($openMenuId === job.id)

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
        await MoveToTrash(job.id, job.path)
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
    <div class="col-orig num">{formatBytes(job.originalSize)}</div>
    <div class="col-output">
        {#if job.status === 'processing'}
            <div class="progress-track">
                <div class="progress-fill" style="width: {job.progress}%"></div>
            </div>
            <span class="progress-pct num">{Math.round(job.progress)}%</span>
            <span class="elapsed num">{job.elapsed || '0:00'}</span>
        {:else if job.status === 'done'}
            <span class="output-size num">{formatBytes(job.outputSize)}</span>
        {:else if job.status === 'error'}
            <span class="error-val" title={job.error}>Error</span>
        {:else}
            <span class="dim">—</span>
        {/if}
    </div>
    <div class="col-savings">
        {#if job.status === 'done'}
            <span class="savings-val num">−{job.savings.toFixed(1)}%</span>
        {:else if job.status === 'skipped' && !job.skipReason}
            <span class="no-gain">No gain</span>
        {:else}
            <span class="dim">—</span>
        {/if}
    </div>
    <div class="col-status">
        {#if job.status === 'processing'}
            <span class="spinner"></span>
        {:else if job.status === 'waiting'}
            <span class="spinner spinner-dim"></span>
        {:else if job.status === 'skipped' && job.skipReason === 'hevc'}
            <span class="icon-warn">⚠<span class="warn-tip">Already H.265 — re-encoding is unlikely to reduce file size</span></span>
        {:else if job.status === 'done'}
            <span class="icon-done">✓</span>
        {:else if job.status === 'error'}
            <span class="icon-error">✕</span>
        {:else}
            <span class="dim">·</span>
        {/if}
    </div>
</div>

{#if showMenu}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="menu" style="left:{menuX}px; top:{menuY}px" onclick={(e) => e.stopPropagation()}>
        <button class="menu-item" onclick={revealInFinder}>Reveal in Finder</button>
        <button class="menu-item" onclick={openFile}>Open file</button>
        {#if canDeleteOriginal}
            <div class="menu-sep"></div>
            <button class="menu-item menu-danger" onclick={deleteOriginal}>Delete original</button>
        {/if}
    </div>
{/if}

<style>
    .row {
        display: flex;
        align-items: center;
        padding: 0 16px;
        height: 42px;
        border-bottom: 1px solid var(--line);
        gap: 8px;
        background: var(--bg-2);
        cursor: pointer;
        user-select: none;
        transition: background var(--dur-fast) var(--ease);
    }

    .row:last-child { border-bottom: none; }
    .row:hover { background: var(--bg-3); }
    .row.processing { background: var(--accent-dim); }

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
        font: 400 12.5px var(--font-mono);
        color: var(--ink);
    }

    .num { font-variant-numeric: tabular-nums; }

    .col-orig {
        width: 80px;
        text-align: right;
        font: 400 12.5px var(--font-mono);
        color: var(--ink-2);
    }

    .col-output {
        width: 190px;
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 6px;
    }

    .col-savings {
        width: 76px;
        text-align: right;
    }

    .col-status {
        width: 26px;
        text-align: center;
        flex-shrink: 0;
    }

    .progress-track {
        flex: 1;
        height: 5px;
        background: var(--line-2);
        border-radius: 3px;
        overflow: hidden;
    }

    .progress-fill {
        height: 100%;
        background: var(--accent);
        border-radius: 3px;
        transition: width 0.25s var(--ease);
    }

    .progress-pct {
        font: 400 11px var(--font-mono);
        color: var(--ink-2);
        min-width: 28px;
        text-align: right;
    }

    .elapsed {
        font: 400 11px var(--font-mono);
        color: var(--ink-3);
        min-width: 30px;
        text-align: right;
    }

    .savings-val { color: var(--accent); font: 600 12.5px var(--font-mono); }
    .no-gain     { font: 400 11px var(--font-mono); color: var(--ink-3); }
    .output-size { font: 400 12.5px var(--font-mono); color: var(--ink-2); }
    .error-val   { font: 500 11.5px var(--font-mono); color: var(--danger); }
    .dim         { color: var(--ink-4); }

    .icon-done  { color: var(--accent); font-weight: 700; font-size: 13px; }
    .icon-error { color: var(--danger); font-size: 13px; }

    .icon-warn {
        position: relative;
        color: var(--seg);
        cursor: default;
        font-size: 13px;
    }

    .warn-tip {
        display: none;
        position: absolute;
        bottom: calc(100% + 6px);
        right: 0;
        background: var(--bg-3);
        border: 1px solid var(--line-2);
        border-radius: var(--radius-lg);
        padding: 6px 9px;
        font: 400 11px var(--font-sans);
        color: var(--ink-2);
        white-space: nowrap;
        box-shadow: var(--shadow-menu);
        pointer-events: none;
        z-index: 400;
    }

    .icon-warn:hover .warn-tip { display: block; }

    .spinner { width: 13px; height: 13px; }
    .spinner-dim { border-top-color: var(--ink-4); opacity: 0.6; }

    /* Context menu */
    .menu {
        position: fixed;
        z-index: 300;
        background: var(--bg-3);
        border: 1px solid var(--line-2);
        border-radius: var(--radius-lg);
        padding: 4px;
        min-width: 190px;
        box-shadow: var(--shadow-menu);
        backdrop-filter: blur(16px);
    }

    .menu-item {
        display: block;
        width: 100%;
        padding: 6px 10px;
        border: none;
        background: none;
        font: 400 13px var(--font-sans);
        color: var(--ink);
        text-align: left;
        cursor: pointer;
        border-radius: var(--radius-md);
    }

    .menu-item:hover { background: var(--accent); color: white; }

    .menu-danger { color: var(--danger); }
    .menu-danger:hover { background: var(--danger); color: white; }

    .menu-sep {
        height: 1px;
        background: var(--line);
        margin: 4px 0;
    }
</style>
