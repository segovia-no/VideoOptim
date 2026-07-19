<script>
    import { onMount } from 'svelte'
    import { EventsOn } from '../../wailsjs/runtime/runtime.js'
    import { AddFiles } from '../../wailsjs/go/main/App.js'
    import { upsertJob } from '../stores/queue.js'

    let { children } = $props()
    let dropping = $state(false)

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

        const onDragEnter = () => dropping = true
        const onDragLeave = (e) => { if (!e.relatedTarget) dropping = false }
        window.addEventListener('dragenter', onDragEnter)
        window.addEventListener('dragleave', onDragLeave)
        return () => {
            window.removeEventListener('dragenter', onDragEnter)
            window.removeEventListener('dragleave', onDragLeave)
        }
    })
</script>

<div class="app" class:dropping>
    {@render children()}
    {#if dropping}
        <div class="drop-overlay">
            <span class="drop-overlay-label">Drop to compress</span>
        </div>
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
</style>
