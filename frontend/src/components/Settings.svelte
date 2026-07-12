<script>
    import { GetSettings, SaveSettings } from '../../wailsjs/go/main/App.js'

    let { onClose } = $props()

    let form = $state({
        encoder: 'libx265',
        crf: 24,
        keepAudio: true,
        discardIfNoGain: true,
        acceptedFormats: ['mp4', 'mov', 'mkv', 'avi', 'webm'],
    })

    const allFormats = ['mp4', 'mov', 'mkv', 'avi', 'webm']

    $effect(() => {
        GetSettings().then(s => { form = s })
    })

    function toggleFormat(fmt) {
        if (form.acceptedFormats.includes(fmt)) {
            form.acceptedFormats = form.acceptedFormats.filter(f => f !== fmt)
        } else {
            form.acceptedFormats = [...form.acceptedFormats, fmt]
        }
    }

    async function save() {
        await SaveSettings(form)
        onClose()
    }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="overlay" onclick={onClose}>
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="modal modal-shell" onclick={(e) => e.stopPropagation()}>
        <h2>Settings</h2>

        <section>
            <span class="label">Encoder</span>
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="radio-card" class:active={form.encoder === 'hevc_videotoolbox'} onclick={() => form.encoder = 'hevc_videotoolbox'}>
                <strong>VideoToolbox HEVC</strong>
                <small>Hardware accelerated, much faster</small>
            </div>
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="radio-card" class:active={form.encoder === 'libx265'} onclick={() => form.encoder = 'libx265'}>
                <strong>libx265</strong>
                <small>Best compression, slower (software)</small>
            </div>
        </section>

        {#if form.encoder === 'libx265'}
        <section>
            <span class="label">Quality (CRF): {form.crf}</span>
            <input type="range" min="18" max="35" step="1" bind:value={form.crf} class="slider" />
            <div class="slider-labels">
                <span>Better quality (18)</span>
                <span>Smaller file (35)</span>
            </div>
        </section>
        {/if}

        <section>
            <label class="toggle-row">
                <span>Keep audio track</span>
                <input type="checkbox" bind:checked={form.keepAudio} />
            </label>
        </section>

        <section>
            <label class="toggle-row">
                <span>
                    Discard output if no gain
                    <small class="toggle-hint">Keep original when compressed file is larger</small>
                </span>
                <input type="checkbox" bind:checked={form.discardIfNoGain} />
            </label>
        </section>

        <section>
            <span class="label">Accepted formats</span>
            <div class="chips">
                {#each allFormats as fmt}
                    <button
                        class="chip"
                        class:active={form.acceptedFormats.includes(fmt)}
                        onclick={() => toggleFormat(fmt)}
                    >{fmt}</button>
                {/each}
            </div>
        </section>

        <div class="actions">
            <button class="btn-cancel" onclick={onClose}>Cancel</button>
            <button class="btn-save" onclick={save}>Save</button>
        </div>
    </div>
</div>

<style>
    .overlay {
        position: fixed;
        inset: 0;
        background: var(--bg-overlay);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 100;
        backdrop-filter: blur(2px);
    }

    .modal {
        padding: 24px;
        width: 400px;
    }

    h2 {
        margin: 0 0 20px;
        font: 600 16px var(--font-sans);
        color: var(--ink);
    }

    section { margin-bottom: 18px; }

    .label {
        display: block;
        font: 500 11px var(--font-mono);
        color: var(--ink-3);
        text-transform: uppercase;
        letter-spacing: 0.06em;
        margin-bottom: 10px;
    }

    .radio-card {
        display: flex;
        flex-direction: column;
        gap: 2px;
        padding: 10px 12px;
        border-radius: var(--radius-md);
        border: 1px solid var(--line);
        background: var(--bg-3);
        cursor: pointer;
        margin-bottom: 8px;
        transition: border-color var(--dur-fast) var(--ease), background var(--dur-fast) var(--ease);
    }

    .radio-card:last-of-type { margin-bottom: 0; }
    .radio-card.active {
        border-color: var(--accent-line);
        background: var(--accent-dim);
    }

    .radio-card strong { font: 500 13px var(--font-sans); color: var(--ink); }
    .radio-card small  { font: 400 11px var(--font-sans); color: var(--ink-3); }

    .slider {
        width: 100%;
        accent-color: var(--accent);
        margin-top: 4px;
    }

    .slider-labels {
        display: flex;
        justify-content: space-between;
        font: 400 10.5px var(--font-mono);
        color: var(--ink-3);
        margin-top: 4px;
    }

    .toggle-row {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        font: 400 13px var(--font-sans);
        cursor: pointer;
        color: var(--ink);
        gap: 12px;
    }

    .toggle-row input[type="checkbox"] { accent-color: var(--accent); margin-top: 2px; }

    .toggle-hint {
        display: block;
        font: 400 11px var(--font-sans);
        color: var(--ink-3);
        margin-top: 2px;
    }

    .chips {
        display: flex;
        gap: 6px;
        flex-wrap: wrap;
    }

    .chip {
        font: 500 11.5px var(--font-mono);
        padding: 4px 10px;
        border-radius: var(--radius-lg);
        border: 1px solid var(--line);
        background: var(--bg-3);
        color: var(--ink-3);
        cursor: pointer;
        transition: border-color var(--dur-fast) var(--ease), background var(--dur-fast) var(--ease), color var(--dur-fast) var(--ease);
    }

    .chip.active {
        background: var(--accent-dim);
        border-color: var(--accent-line);
        color: var(--accent);
    }

    .actions {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
        margin-top: 20px;
        padding-top: 16px;
        border-top: 1px solid var(--line);
    }

    .btn-cancel, .btn-save {
        padding: 6px 14px;
        border-radius: var(--radius-md);
        font: 500 13px var(--font-sans);
        cursor: pointer;
        border: none;
    }

    .btn-cancel {
        background: var(--bg-3);
        color: var(--ink-2);
        border: 1px solid var(--line-2);
    }

    .btn-cancel:hover { background: var(--bg-btn-hover); }

    .btn-save {
        background: var(--accent);
        color: white;
    }

    .btn-save:hover { background: var(--accent-hover); }
</style>
