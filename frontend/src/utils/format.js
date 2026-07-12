export function formatBytes(bytes) {
    if (!bytes) return '—'
    const gb = bytes / 1073741824
    const mb = bytes / 1048576
    return gb >= 1 ? `${gb.toFixed(2)} GB` : mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}
