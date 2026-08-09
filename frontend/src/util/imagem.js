// Utilitário pra converter uma imagem PNG codificada em base64 (como vem
// dos bindings Go, ex.: ObterImagemGrafico) num Blob, pronto pra usar com
// navigator.clipboard.write ou como src de <img>.
export function base64ParaBlob(base64, mime) {
    const binario = atob(base64);
    const bytes = new Uint8Array(binario.length);
    for (let i = 0; i < binario.length; i++) {
        bytes[i] = binario.charCodeAt(i);
    }
    return new Blob([bytes], { type: mime });
}
