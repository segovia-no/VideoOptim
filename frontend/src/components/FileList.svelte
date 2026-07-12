<script>
    import FileRow from './FileRow.svelte'
    import { jobs } from '../stores/queue.js'
</script>

<div class="file-list">
    {#if $jobs.length > 0}
        <div class="header-row">
            <div class="col-name">File</div>
            <div class="col-orig">Original</div>
            <div class="col-output">Output</div>
            <div class="col-savings">Savings</div>
            <div class="col-status"></div>
        </div>
        <div class="rows">
            {#each $jobs as job (job.id)}
                <FileRow {job} />
            {/each}
        </div>
    {/if}
</div>

<style>
    .file-list {
        flex: 1;
        overflow-y: auto;
        min-height: 0;
        background: var(--bg-content);
    }

    .header-row {
        display: flex;
        align-items: center;
        padding: 0 16px;
        height: 26px;
        border-bottom: 1px solid var(--border);
        background: var(--bg-header);
        gap: 8px;
        font-size: 11px;
        font-weight: 600;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        position: sticky;
        top: 0;
        z-index: 1;
    }

    .col-name    { flex: 1; }
    .col-orig    { width: 80px;  text-align: right; }
    .col-output  { width: 170px; text-align: right; }
    .col-savings { width: 72px;  text-align: right; }
    .col-status  { width: 22px; }
</style>
