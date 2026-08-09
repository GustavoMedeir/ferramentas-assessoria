// Ícones SVG (inline, como string) usados na UI. Markup fixo e confiável
// (não vem de input do usuário), por isso é seguro atribuir via innerHTML.
// Todos viewBox="0 0 24 24", stroke-based, seguindo o estilo do design.

function svg(paths) {
    return `<svg viewBox="0 0 24 24">${paths}</svg>`;
}

export const iconRentabilidade = svg('<path d="M3 17l6-6 4 4 8-8"/><path d="M17 7h4v4"/>');
export const iconEmail = svg('<rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 7l9 6 9-6"/>');
export const iconTabela = svg('<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M3 10h18M9 10v10"/>');
// Ícones do item de navegação na sidebar — versões redesenhadas (v2), mais
// distintas visualmente das usadas em outros lugares (cat-h da Calculadora,
// abas do popup de Configurações), que mantêm os ícones acima.
export const iconVelocimetro = svg('<path d="M4.5 16a7.5 7.5 0 0 1 15 0"/><path d="M12 16l4-6"/><circle cx="12" cy="16" r="1.3"/>');
export const iconDocumento = svg('<path d="M21 11V6a1 1 0 0 0-1-1H4a1 1 0 0 0-1 1v11a1 1 0 0 0 1 1h9"/><path d="M3 7l9 5 9-5"/><path d="M15.5 18.5l2 2 3.5-3.5"/>');
export const iconDesagio = svg('<rect x="2.5" y="6.5" width="14" height="10" rx="2"/><circle cx="9.5" cy="11.5" r="2"/><path d="M21 9v8m0 0l-2.3-2.3M21 17l2.3-2.3"/>');
export const iconCofrinho = svg(
    '<path d="M12 3l7 3v5c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z"/><path d="M9.5 14.5 14.5 9.5"/><circle cx="10" cy="10" r=".9"/><circle cx="14" cy="14" r=".9"/>'
);
export const iconComparadora = svg('<path d="M12 3v17"/><path d="M7.5 20h9"/><path d="M4 7h16"/><path d="M4 7v4.5"/><path d="M20 7v4.5"/><path d="M1.5 11.5a2.5 2.5 0 0 0 5 0z"/><path d="M17.5 11.5a2.5 2.5 0 0 0 5 0z"/><circle cx="12" cy="7" r="1.2"/>');
export const iconCalculadora = svg(
    '<rect x="5" y="3" width="14" height="18" rx="2"/><path d="M8 7h8"/><path d="M8.5 12h.01M12 12h.01M15.5 12h.01M8.5 16h.01M12 16h.01M15.5 16h.01"/>'
);
export const iconPessoa = svg(
    '<circle cx="8.5" cy="8" r="3"/><path d="M3 20c0-3.5 2.5-5.6 5.5-5.6s5.5 2.1 5.5 5.6"/><circle cx="17" cy="9" r="2.4"/><path d="M15.3 14.5c2.6.5 4.2 2.4 4.2 5.5"/>'
);
export const iconConfig = svg('<path d="M4 7h9M17 7h3M4 17h5M13 17h7"/><circle cx="15" cy="7" r="2"/><circle cx="11" cy="17" r="2"/>');
// Ícone do item de navegação "Configurações" na sidebar — distinto do
// iconConfig acima (que segue usado na aba "Temas" dentro do popup).
export const iconConfigNav = svg(
    '<circle cx="12" cy="12" r="3"/><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>'
);
export const iconPasta = svg('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>');
export const iconAtualizar = svg('<path d="M21 12a9 9 0 1 1-3-6.7"/><path d="M21 4v5h-5"/>');
export const iconExportar = svg('<path d="M12 3v12"/><path d="M8 11l4 4 4-4"/><path d="M4 21h16"/>');
export const iconEnviar = svg('<path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4 20-7z"/>');
export const iconBusca = svg('<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>');
export const iconImagem = svg('<rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="8.5" cy="9.5" r="1.5"/><path d="M21 15l-5-5-8 8"/>');
export const iconCopiar = svg('<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/>');
export const iconSalvar = svg('<path d="M5 4h11l3 3v13H5z"/><path d="M8 4v5h7"/><rect x="8" y="13" width="8" height="5" rx="1"/>');
export const iconCheck = svg('<path d="M20 6L9 17l-5-5"/>');
export const iconMais = svg('<path d="M12 5v14M5 12h14"/>');
export const iconMenos = svg('<path d="M5 12h14"/>');
export const iconArrastar = svg(
    '<circle cx="9" cy="6" r="1.2"/><circle cx="15" cy="6" r="1.2"/><circle cx="9" cy="12" r="1.2"/><circle cx="15" cy="12" r="1.2"/><circle cx="9" cy="18" r="1.2"/><circle cx="15" cy="18" r="1.2"/>'
);
export const iconAlerta = svg('<circle cx="12" cy="12" r="9"/><path d="M12 11v5M12 8h.01"/>');
export const iconRelogio = svg('<circle cx="12" cy="12" r="8"/><path d="M12 8v4l3 2"/>');
export const iconLista = svg('<path d="M4 7h16M4 12h16M4 17h16"/>');
export const iconLixeira = svg('<path d="M4 7h16M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3m-9 0 1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13"/>');
export const iconFechar = svg('<path d="M18 6L6 18M6 6l12 12"/>');
export const iconPrevidencia = svg('<path d="M12 3l7 3v6c0 5-3.5 8-7 9-3.5-1-7-4-7-9V6l7-3z"/><path d="M9 12l2 2 4-4"/>');
export const iconWhatsApp = svg(
    '<path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>'
);
export const iconCompromissada = svg(
    '<circle cx="10" cy="10" r="6.5"/><path d="M10 6.5V10l2.5 1.5"/><circle cx="17.5" cy="17" r="3.6"/><path d="M17.5 15.4v3.2"/>'
);
export const iconOlho = svg('<path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>');
export const iconInfo = svg('<circle cx="12" cy="12" r="9"/><path d="M12 8h.01M12 11v5"/>');
export const iconMaisPontos = svg('<circle cx="5" cy="12" r="1.6"/><circle cx="12" cy="12" r="1.6"/><circle cx="19" cy="12" r="1.6"/>');
export const iconChevronBaixo = svg('<path d="M6 9l6 6 6-6"/>');
export const iconEstrela = svg('<path d="M12 2l2.9 6.3 6.9.7-5.2 4.6 1.5 6.8L12 18.6 5.9 21.7l1.5-6.8L2.2 9.3l6.9-.7z"/>');
export const iconMeta = svg('<path d="M4 20c3.5-5 12-5 16 0"/><path d="M9 15V4l7 3-7 3"/>');
export const iconApresentacao = svg('<rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8M12 17v4"/><path d="M10.5 8l4 2.5-4 2.5z"/>');
export const iconTelaCheia = svg('<path d="M4 9V5a1 1 0 0 1 1-1h4M15 4h4a1 1 0 0 1 1 1v4M20 15v4a1 1 0 0 1-1 1h-4M9 20H5a1 1 0 0 1-1-1v-4"/>');
export const iconBorracha = svg(
    '<path d="m7 21-4.3-4.3c-1-1-1-2.5 0-3.4l9.6-9.6c1-1 2.5-1 3.4 0l5.6 5.6c1 1 1 2.5 0 3.4L13 21"/><path d="M22 21H7"/><path d="m5 11 9 9"/>'
);
export const iconRedefinir = svg('<path d="M3 12a9 9 0 1 0 3-6.7"/><path d="M3 4v5h5"/>');
export const iconFormulario = svg(
    '<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 8h.01M8 12h.01M8 16h.01"/><path d="M11.5 8h5M11.5 12h5M11.5 16h5"/>'
);
export const iconFesta = svg(
    '<rect x="4" y="9" width="16" height="11" rx="1.5"/><path d="M4 13h16"/><path d="M12 9V5"/><path d="M12 5c-1.5 0-2.5-1-2.5-2S10.5 1.5 12 3c1.5-1.5 2.5-.5 2.5.5S13.5 5 12 5z"/>'
);
export const iconEditarPDF = svg(
    '<path d="M14 3H7a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h6"/><path d="M13 3v5h5"/><path d="M20.4 13.6 15 19l-3 .8.8-3 5.4-5.4a1.35 1.35 0 0 1 2.2 2.2z"/>'
);
export const iconCaneta = svg('<path d="M4 20l1-4L15.5 5.5a2 2 0 0 1 3 3L8 19l-4 1z"/><path d="M13 7l3 3"/>');
export const iconTexto = svg('<path d="M5 6h14M12 6v14"/>');
export const iconTarja = svg('<rect x="3" y="9" width="18" height="6" rx="1"/>');
export const iconDesfazer = svg('<path d="M9 7L4 12l5 5"/><path d="M4 12h11a5 5 0 0 1 0 10h-1"/>');
export const iconMarcaX = svg('<path d="M6 6l12 12M18 6L6 18"/>');
export const iconAssinatura = svg(
    '<path d="M3 17c2-1 3-3 3.5-5.5S7 6 8.5 6c1.8 0 1 5 2.5 7s3.5-1 4-3 1 3 3 3 2-1.5 3-1.5"/>'
);
export const iconOrdenar = svg('<path d="M7 4v16M7 4L4 7M7 4l3 3"/><path d="M17 20V4M17 20l-3-3M17 20l3-3"/>');
export const iconImagensPDF = svg(
    '<path d="M14 3H7a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V7z"/><path d="M13 3v5h5"/><circle cx="9.8" cy="13" r="1.2"/><path d="M8 18l2.5-2.5 1.6 1.6L15 14.5"/>'
);

// Ícone dedicado do item de navegação "Validar Assinatura" — separado de
// iconAssinatura (que continua com o rabisco antigo) porque as duas telas
// são conceitualmente diferentes apesar do nome parecido: validação de
// assinatura digital ICP-Brasil aqui, cadastro de imagem de assinatura
// pra carimbar PDF em Configurações → Assinatura.
export const iconValidarAssinatura = svg(
    '<path d="M14 3H7a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h5"/><path d="M13 3v5h5"/><circle cx="16.5" cy="16.5" r="4.2"/><path d="M14.7 16.6l1.2 1.2 2.3-2.3"/>'
);
