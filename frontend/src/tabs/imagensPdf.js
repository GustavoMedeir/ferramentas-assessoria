import { novoBlocoId } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { CriarPDFDeImagens } from "../../wailsjs/go/main/App.js";

// Resolução assumida pra converter pixels das imagens escolhidas em pontos
// de página PDF (1/72 polegada) — fotos/capturas de tela raramente trazem
// DPI embutido, então assumimos um valor razoável de leitura/impressão em
// vez de gerar páginas gigantes (fotos de celular têm milhares de pixels
// de lado).
const DPI_IMAGEM = 150;

let refs = {};
let imagens = []; // { id, nome, img (Image já carregada) }
let indiceArrastado = null;

export function mount(container) {
    clear(container);
    refs = {};
    imagens = [];
    indiceArrastado = null;

    refs.vazio = el("div", { class: "estado-vazio" }, [
        el("div", { text: "Nenhuma imagem adicionada ainda." }),
        btn("Adicionar imagens", { classe: "pri", icon: icons.iconMais, onClick: adicionarImagens }),
    ]);
    container.appendChild(refs.vazio);

    refs.conteudo = el("div", { class: "imgpdf" });
    refs.conteudo.style.display = "none";

    refs.conteudo.appendChild(
        el("div", { class: "mail-r-h" }, [
            el("h3", { text: "Imagens" }),
            btn("Adicionar imagens", { icon: icons.iconMais, onClick: adicionarImagens }),
        ])
    );

    refs.lista = el("div", { class: "imgpdf-lista" });
    refs.conteudo.appendChild(refs.lista);

    refs.status = el("p", { class: "cfg-sub" });
    refs.conteudo.appendChild(refs.status);

    refs.btnGerar = btn("Gerar PDF", { classe: "pri", icon: icons.iconSalvar, onClick: gerarPDF });
    refs.conteudo.appendChild(el("div", { class: "des-foot" }, [refs.btnGerar]));

    container.appendChild(refs.conteudo);
}

function atualizarVisibilidade() {
    const temImagens = imagens.length > 0;
    refs.vazio.style.display = temImagens ? "none" : "flex";
    refs.conteudo.style.display = temImagens ? "flex" : "none";
}

function carregarImagem(src) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error("não foi possível carregar a imagem"));
        img.src = src;
    });
}

function lerComoDataURL(arquivo) {
    return new Promise((resolve, reject) => {
        const leitor = new FileReader();
        leitor.onload = () => resolve(leitor.result);
        leitor.onerror = () => reject(new Error("não foi possível ler o arquivo"));
        leitor.readAsDataURL(arquivo);
    });
}

// Pode ser clicado várias vezes — cada seleção soma no fim da lista, não
// substitui a anterior.
function adicionarImagens() {
    const input = el("input", { type: "file", accept: "image/*", multiple: "true" });
    input.addEventListener("change", async () => {
        const arquivos = Array.from(input.files || []);
        if (!arquivos.length) return;
        refs.status.textContent = "Carregando imagens...";
        for (const arquivo of arquivos) {
            try {
                const dataUrl = await lerComoDataURL(arquivo);
                const img = await carregarImagem(dataUrl);
                imagens.push({ id: novoBlocoId(), nome: arquivo.name, img });
            } catch (e) {
                refs.status.textContent = "Erro ao carregar " + arquivo.name + ": " + e;
            }
        }
        if (refs.status.textContent === "Carregando imagens...") refs.status.textContent = "";
        renderLista();
        atualizarVisibilidade();
    });
    input.click();
}

function mover(indice, delta) {
    const alvo = indice + delta;
    if (alvo < 0 || alvo >= imagens.length) return;
    [imagens[indice], imagens[alvo]] = [imagens[alvo], imagens[indice]];
    renderLista();
}

function moverPara(origem, destino) {
    if (origem === destino) return;
    const [item] = imagens.splice(origem, 1);
    imagens.splice(destino, 0, item);
    renderLista();
}

function removerImagem(id) {
    imagens = imagens.filter((im) => im.id !== id);
    renderLista();
    atualizarVisibilidade();
}

function renderLista() {
    clear(refs.lista);
    imagens.forEach((imagem, i) => {
        const linha = el("div", { class: "imgpdf-item", draggable: "true" });

        const alca = el("span", { class: "imgpdf-alca" });
        alca.insertAdjacentHTML("beforeend", icons.iconArrastar);
        linha.appendChild(alca);

        linha.appendChild(el("span", { class: "imgpdf-num", text: String(i + 1) }));

        const thumb = el("img", { class: "imgpdf-thumb" });
        thumb.src = imagem.img.src;
        linha.appendChild(thumb);

        linha.appendChild(el("span", { class: "imgpdf-nome", text: imagem.nome }));

        const botaoCima = el("button", { class: "ordem-seta up", type: "button", "aria-label": "Mover para cima" });
        botaoCima.insertAdjacentHTML("beforeend", icons.iconChevronBaixo);
        botaoCima.disabled = i === 0;
        botaoCima.addEventListener("click", () => mover(i, -1));

        const botaoBaixo = el("button", { class: "ordem-seta", type: "button", "aria-label": "Mover para baixo" });
        botaoBaixo.insertAdjacentHTML("beforeend", icons.iconChevronBaixo);
        botaoBaixo.disabled = i === imagens.length - 1;
        botaoBaixo.addEventListener("click", () => mover(i, 1));

        const botaoRemover = el("button", { class: "ordem-seta imgpdf-remover", type: "button", "aria-label": "Remover" });
        botaoRemover.insertAdjacentHTML("beforeend", icons.iconLixeira);
        botaoRemover.addEventListener("click", () => removerImagem(imagem.id));

        linha.appendChild(el("div", { class: "imgpdf-acoes" }, [botaoCima, botaoBaixo, botaoRemover]));

        // Mesmo padrão de arrastar-e-soltar de main.js:montarAbaOrdemNav —
        // dragover precisa de preventDefault() pra o drop ser aceito.
        linha.addEventListener("dragstart", (e) => {
            indiceArrastado = i;
            linha.classList.add("arrastando");
            e.dataTransfer.effectAllowed = "move";
        });
        linha.addEventListener("dragend", () => {
            indiceArrastado = null;
            linha.classList.remove("arrastando");
        });
        linha.addEventListener("dragover", (e) => e.preventDefault());
        linha.addEventListener("drop", (e) => {
            e.preventDefault();
            if (indiceArrastado === null) return;
            moverPara(indiceArrastado, i);
        });

        refs.lista.appendChild(linha);
    });
}

// paraJPEGBase64 achata a imagem (fundo branco antes de desenhar, pro caso
// de PNG com transparência) num JPEG — mesmo padrão de
// tabs/editorPdf.js:paraJPEGBase64.
function paraJPEGBase64(img) {
    const c = document.createElement("canvas");
    c.width = img.naturalWidth;
    c.height = img.naturalHeight;
    const ctx = c.getContext("2d");
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, c.width, c.height);
    ctx.drawImage(img, 0, 0);
    return c.toDataURL("image/jpeg", 0.92).split(",")[1];
}

async function gerarPDF() {
    if (!imagens.length) return;
    refs.btnGerar.disabled = true;
    refs.status.textContent = "Gerando PDF...";
    try {
        const paginas = imagens.map((imagem) => ({
            JPEGBase64: paraJPEGBase64(imagem.img),
            LarguraPt: (imagem.img.naturalWidth * 72) / DPI_IMAGEM,
            AlturaPt: (imagem.img.naturalHeight * 72) / DPI_IMAGEM,
        }));
        const caminho = await CriarPDFDeImagens(paginas);
        refs.status.textContent = caminho ? `Salvo em: ${caminho}` : "";
    } catch (e) {
        refs.status.textContent = "Erro ao gerar PDF: " + e;
    } finally {
        refs.btnGerar.disabled = false;
    }
}
