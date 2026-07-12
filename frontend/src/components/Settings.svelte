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
    <div class="modal" onclick={(e) => e.stopPropagation()}>
        <h2>Settings</h2>

        <section>
            <span class="section-label">Encoder</span>
            <div class="radio-group">
                <label class="radio">
                    <input type="radio" bind:group={form.encoder} value="libx265" />
                    <span class="radio-label">
                        <strong>libx265</strong>
                        <small>Best compression, slower (software)</small>
                    </span>
                </label>
                <label class="radio">
                    <input type="radio" bind:group={form.encoder} value="hevc_videotoolbox" />
                    <span class="radio-label">
                        <strong>VideoToolbox HEVC</strong>
                        <small>Hardware accelerated, much faster</small>
                    </span>
                </label>
            </div>
        </section>

        {#if form.encoder === 'libx265'}
        <section>
            <span class="section-label">Quality (CRF): {form.crf}</span>
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
            <span class="section-label">Accepted formats</span>
            <div class="format-chips">
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
    }

    .modal {
        background: var(--bg-modal);
        border-radius: 12px;
        padding: 24px;
        width: 380px;
        box-shadow: 0 24px 64px rgba(0,0,0,0.35);
        border: 1px solid var(--border);
    }

    h2 {
        margin: 0 0 20px;
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);
    }

    section { margin-bottom: 18px; }

    .section-label {
        display: block;
        font-size: 11px;
        font-weight: 600;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        margin-bottom: 8px;
    }

    .radio-group {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .radio {
        display: flex;
        align-items: flex-start;
        gap: 8px;
        cursor: pointer;
    }

    .radio input { margin-top: 2px; }

    .radio-label {
        display: flex;
        flex-direction: column;
        gap: 1px;
    }

    .radio-label strong { font-size: 13px; color: var(--text-primary); }
    .radio-label small  { font-size: 11px; color: var(--text-muted); }

    .slider {
        width: 100%;
        accent-color: var(--accent);
        margin-top: 4px;
    }

    .slider-labels {
        display: flex;
        justify-content: space-between;
        font-size: 10px;
        color: var(--text-muted);
        margin-top: 3px;
    }

    .toggle-row {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        font-size: 13px;
        cursor: pointer;
        color: var(--text-primary);
        gap: 12px;
    }

    .toggle-hint {
        display: block;
        font-size: 11px;
        color: var(--text-muted);
        font-weight: normal;
        margin-top: 2px;
    }

    .format-chips {
        display: flex;
        gap: 6px;
        flex-wrap: wrap;
    }

    .chip {
        padding: 3px 10px;
        border-radius: 10px;
        border: 1px solid var(--border-btn);
        background: var(--bg-chip);
        font-size: 12px;
        cursor: pointer;
        color: var(--text-secondary);
        font-family: inherit;
    }

    .chip.active {
        background: var(--accent);
        border-color: var(--accent);
        color: white;
    }

    .actions {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
        margin-top: 20px;
        padding-top: 16px;
        border-top: 1px solid var(--border);
    }

    .btn-cancel, .btn-save {
        padding: 6px 14px;
        border-radius: 7px;
        border: none;
        font-size: 13px;
        cursor: pointer;
        font-family: inherit;
    }

    .btn-cancel {
        background: var(--bg-btn);
        color: var(--text-btn);
        border: 1px solid var(--border-btn);
    }

    .btn-cancel:hover { background: var(--bg-btn-hover); }

    .btn-save {
        background: var(--accent);
        color: white;
        font-weight: 500;
    }

    .btn-save:hover { background: var(--accent-hover); }
</style>
