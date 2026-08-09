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
export const iconDocumento = svg('<path d="M7 3h7l4 4v13a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/><path d="M14 3v4h4"/><path d="M9 13h6M9 16.5h6"/>');
export const iconDesagio = svg('<rect x="3" y="5" width="13" height="13" rx="2"/><path d="M3 10h13M8 10v8"/><path d="M19 14v6m0 0l-2.3-2.3M19 20l2.3-2.3"/>');
export const iconCofrinho = svg(
    '<path d="M5 13a5 5 0 0 1 5-5h4a5 5 0 0 1 5 5v1.5a1.5 1.5 0 0 1-1.5 1.5H17l-.7 2.5h-2L14 16h-4l-.3 2H8l-.7-2.5H6.5A1.5 1.5 0 0 1 5 14.5z"/><circle cx="15.5" cy="10.5" r="1"/><path d="M8.5 8V6"/>'
);
export const iconComparadora = svg('<path d="M3 12h6"/><path d="M6 8l3 4-3 4"/><path d="M21 12h-6"/><path d="M18 8l-3 4 3 4"/>');
export const iconCalculadora = svg(
    '<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 7h8M8 12h.01M12 12h.01M16 12h.01M8 16h.01M12 16h.01M16 16h.01"/>'
);
export const iconPessoa = svg(
    '<circle cx="8.5" cy="8" r="3"/><path d="M3 20c0-3.5 2.5-5.6 5.5-5.6s5.5 2.1 5.5 5.6"/><circle cx="17" cy="9" r="2.4"/><path d="M15.3 14.5c2.6.5 4.2 2.4 4.2 5.5"/>'
);
export const iconConfig = svg('<path d="M4 7h9M17 7h3M4 17h5M13 17h7"/><circle cx="15" cy="7" r="2"/><circle cx="11" cy="17" r="2"/>');
// Ícone do item de navegação "Configurações" na sidebar — distinto do
// iconConfig acima (que segue usado na aba "Temas" dentro do popup).
export const iconConfigNav = svg(
    '<circle cx="12" cy="12" r="3"/><path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M18.4 5.6l-2.1 2.1M7.7 16.3l-2.1 2.1"/>'
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
    '<circle cx="10" cy="10" r="6.5"/><path d="M10 6.5v3.5l2.5 1.5"/><circle cx="18" cy="17" r="3.6"/><path d="M16.5 17h3"/>'
);
export const iconOlho = svg('<path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>');
export const iconInfo = svg('<circle cx="12" cy="12" r="9"/><path d="M12 8h.01M12 11v5"/>');
export const iconMaisPontos = svg('<circle cx="5" cy="12" r="1.6"/><circle cx="12" cy="12" r="1.6"/><circle cx="19" cy="12" r="1.6"/>');
export const iconChevronBaixo = svg('<path d="M6 9l6 6 6-6"/>');
export const iconEstrela = svg('<path d="M12 2l2.9 6.3 6.9.7-5.2 4.6 1.5 6.8L12 18.6 5.9 21.7l1.5-6.8L2.2 9.3l6.9-.7z"/>');
export const iconMeta = svg('<circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="5"/><circle cx="12" cy="12" r="1"/>');
export const iconApresentacao = svg('<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M8 20h8M12 16v4"/>');
export const iconTelaCheia = svg('<path d="M4 9V5a1 1 0 0 1 1-1h4M15 4h4a1 1 0 0 1 1 1v4M20 15v4a1 1 0 0 1-1 1h-4M9 20H5a1 1 0 0 1-1-1v-4"/>');
export const iconBorracha = svg(
    '<path d="m7 21-4.3-4.3c-1-1-1-2.5 0-3.4l9.6-9.6c1-1 2.5-1 3.4 0l5.6 5.6c1 1 1 2.5 0 3.4L13 21"/><path d="M22 21H7"/><path d="m5 11 9 9"/>'
);
export const iconRedefinir = svg('<path d="M3 12a9 9 0 1 0 3-6.7"/><path d="M3 4v5h5"/>');
export const iconFormulario = svg(
    '<rect x="5" y="3" width="14" height="18" rx="2"/><path d="M9 3h6v3H9z"/><path d="M8 11l1.5 1.5L12 10"/><path d="M8 16h8"/>'
);
export const iconFesta = svg(
    '<rect x="4" y="9" width="16" height="11" rx="1.5"/><path d="M4 13h16"/><path d="M12 9V5"/><path d="M12 5c-1.5 0-2.5-1-2.5-2S10.5 1.5 12 3c1.5-1.5 2.5-.5 2.5.5S13.5 5 12 5z"/>'
);
export const iconEditarPDF = svg(
    '<path d="M7 3h7l4 4v6"/><path d="M14 3v4h4"/><path d="M14 15l5.5 5.5"/><path d="M12.5 16.5l6-6a1.4 1.4 0 0 1 2 2l-6 6-2.7.7z"/>'
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
    '<path d="M7 3h7l4 4v13a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/><path d="M14 3v4h4"/><circle cx="10" cy="12.5" r="1.1"/><path d="M8 17l2.5-2.5 1.8 1.8L15 13"/>'
);
