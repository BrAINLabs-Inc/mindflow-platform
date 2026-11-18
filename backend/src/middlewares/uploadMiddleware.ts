import multer from 'multer';

const ALLOWED_AUDIO_MIMES = new Set([
    'audio/wav',
    'audio/x-wav',
    'audio/wave',
    'audio/mpeg',
    'audio/mp4',
    'audio/webm',
    'audio/ogg',
    'application/octet-stream',
]);

const AUDIO_MIME_BY_EXT: Record<string, string> = {
    '.wav': 'audio/wav',
    '.mp3': 'audio/mpeg',
    '.m4a': 'audio/mp4',
    '.webm': 'audio/webm',
    '.ogg': 'audio/ogg',
};

// The fileFilter below accepts a file if EITHER its declared mimetype OR its extension is
// allowlisted, so the declared mimetype alone can't be trusted as the stored Content-Type:
// "x.wav" declared as text/html would otherwise be served from the public R2 bucket as HTML.
export function storedAudioContentType(mimetype: string | undefined, originalname: string | undefined): string {
    const mime = (mimetype || '').toLowerCase();
    if (ALLOWED_AUDIO_MIMES.has(mime)) return mime;
    const ext = (originalname || '').toLowerCase().match(/\.(wav|mp3|m4a|webm|ogg)$/)?.[0];
    return (ext && AUDIO_MIME_BY_EXT[ext]) || 'application/octet-stream';
}

const storage = multer.memoryStorage();

export const upload = multer({
    storage,
    limits: {
        fileSize: 10 * 1024 * 1024,
    },
    fileFilter: (_req, file, cb) => {
        const mime = (file.mimetype || '').toLowerCase();
        const ext = (file.originalname || '').toLowerCase();
        const extOk = /\.(wav|mp3|m4a|webm|ogg)$/.test(ext);
        if (ALLOWED_AUDIO_MIMES.has(mime) || extOk) {
            cb(null, true);
        } else {
            cb(new Error('Only audio files (wav, mp3, m4a, webm, ogg) are allowed'));
        }
    },
});
